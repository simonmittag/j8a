package j8a

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/rs/zerolog/log"
	"strconv"
)

type KeySet []KidPair

func (ks *KeySet) Upsert(kp KidPair) {
	updated := false
	for _, k := range *ks {
		if k.Kid == kp.Kid {
			k.Key = kp.Key
			updated = true
		}
	}
	if !updated {
		*ks = append(*ks, kp)
	}
}

func (ks *KeySet) Find(kid string) interface{} {
	for _, k := range *ks {
		if k.Kid == kid {
			return k.Key
		}
	}
	return nil
}

type KidPair struct {
	Kid string
	Key interface{}
}

type Jwt struct {
	Name string
	Alg  string
	// Jwt key supports pem encoding for public keys, certificates unencoded secrets for hmac.
	Key string
	// JwksUrl loads remotely.
	JwksUrl               string
	RSAPublic             KeySet
	ECDSAPublic           KeySet
	Secret                KeySet
	AcceptableSkewSeconds string
}

var validAlgNoNone = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "HS256", "HS384", "HS512", "ES256", "ES384", "ES512"}
var validAlg = append(validAlgNoNone, "none")

const pemOverflow = "jwt key [%s] only type PUBLIC KEY or CERTIFICATE allowed but found additional or invalid data, check your PEM block"
const pemTypeBad = "jwt key [%s] is not of type PUBLIC KEY or CERTIFICATE, check your PEM Block preamble"
const pemAsn1Bad = "jwt key [%s] asn data not valid, check your PEM Block"
const pemRsaNotFound = "jwt key [%s] RSA public key not found in your certificate, check your PEM Block"
const pemEcdsaNotFound = "jwt key [%s] ECDSA public key not found in your certificate, check your PEM Block"

const keyTypeInvalid = "jwt [%s] unable to determine key type. Must be one of %s"
const unknownAlg = "jwt [%s] unknown alg [%s]. Must be one of %s"
const missingAlg = "jwt [%s] missing mandatory alg parameter next to jwksUrl. Must be one of %s"
const noneWithKeyData = "jwt [%s] none type signature does not allow key data, check your configuration"
const missingKeyOrJwks = "jwt [%s] alg [%s] must specify one of key or jwksUrl"
const skewInvalid = "jwt [%s] acceptable skew seconds, must be 0 or greater, was %s"

const ecdsaKeySizeBad = "jwt [%s] invalid key size for alg [%s], parsed bitsize %d, check your configuration"

const defaultSkew = "120"

func (jwt *Jwt) validate() error {
	if jwt.RSAPublic == nil {
		jwt.RSAPublic = make([]KidPair, 0)
	}
	if jwt.ECDSAPublic == nil {
		jwt.ECDSAPublic = make([]KidPair, 0)
	}
	if jwt.Secret == nil {
		jwt.Secret = make([]KidPair, 0)
	}

	var err error
	alg := *new(jwa.SignatureAlgorithm)
	err = alg.Accept(jwt.Alg)

	if len(jwt.Name) == 0 {
		return errors.New("invalid jwt name not specified")
	}

	if len(jwt.Alg) > 0 {
		matched := false
		for _, alg := range validAlg {
			if alg == jwt.Alg {
				matched = true
			}
		}
		if !matched {
			return errors.New(fmt.Sprintf(unknownAlg, jwt.Name, jwt.Alg, validAlg))
		}
	}

	if len(jwt.Alg) == 0 && len(jwt.JwksUrl) > 0 {
		return errors.New(fmt.Sprintf(missingAlg, jwt.Name, validAlgNoNone))
	}

	if len(jwt.Alg) == 0 && len(jwt.Key) > 0 {
		return errors.New(fmt.Sprintf(missingAlg, jwt.Name, validAlgNoNone))
	}

	if alg == jwa.NoSignature && len(jwt.Key) > 0 {
		return errors.New(fmt.Sprintf(noneWithKeyData, jwt.Name))
	}

	if alg != jwa.NoSignature && len(jwt.Key) == 0 && len(jwt.JwksUrl) == 0 {
		err = errors.New(fmt.Sprintf(missingKeyOrJwks, jwt.Name, alg))
	}

	if len(jwt.AcceptableSkewSeconds) > 0 {
		secs, nonnumeric := strconv.Atoi(jwt.AcceptableSkewSeconds)
		if nonnumeric != nil || secs < 0 {
			err = errors.New(fmt.Sprintf(skewInvalid, jwt.Name, jwt.AcceptableSkewSeconds))
			return err
		}
	} else {
		jwt.AcceptableSkewSeconds = defaultSkew
	}

	if len(jwt.Key) > 0 {
		err = jwt.parseKey(alg)
	} else if len(jwt.JwksUrl) > 0 {
		err = jwt.LoadJwks()
	}

	return err
}

func (jwt *Jwt) LoadJwks() error {
	keyset, err := jwk.Fetch(jwt.JwksUrl)
	if err == nil {
		log.Debug().Msgf("jwt [%s] fetched %d jwk from jwks URL %s", jwt.Name, keyset.Len(), jwt.JwksUrl)
	} else {
		return err
	}

	if keyset.Keys == nil || len(keyset.Keys) == 0 {
		return errors.New(fmt.Sprintf("jwt [%s] unable to parse keys in keyset", jwt.Name))
	}

	for _, key := range keyset.Keys {
		//here, use the key's signature algorithm, not what's supplied in the config.
		alg := *new(jwa.SignatureAlgorithm)
		err = alg.Accept(key.Algorithm())

		if jwt.Alg != key.Algorithm() {
			return errors.New(fmt.Sprintf("jwt [%s] key algorithm %s in jwks keyset does not match configured alg %s", jwt.Name, key.Algorithm(), jwt.Alg))
		}

		switch alg {
		case jwa.RS256, jwa.RS384, jwa.RS512, jwa.PS256, jwa.PS384, jwa.PS512:
			k := KidPair{
				Kid: key.KeyID(),
				Key: &rsa.PublicKey{
					N: nil,
					E: 0,
				},
			}
			err = key.Raw(k.Key)
			if err == nil {
				jwt.RSAPublic.Upsert(k)
			}
		//Note, removed support for HS256, secret keys make no sense for JWKS even over TLS.
		case jwa.ES256, jwa.ES384, jwa.ES512:
			k := KidPair{
				Kid: key.KeyID(),
				Key: &ecdsa.PublicKey{
					Curve: nil,
					X:     nil,
					Y:     nil,
				},
			}
			err = key.Raw(k.Key)
			err = jwt.checkECDSABitSize(alg, k.Key.(*ecdsa.PublicKey))
			if err == nil {
				jwt.ECDSAPublic.Upsert(k)
			}
		default:
			err = errors.New(fmt.Sprintf("unknown key type in Jwks %v", alg.String()))
		}
		log.Debug().Msgf("jwt [%s] successfully parsed %s key from remote jwk", jwt.Name, alg)
	}
	return err
}

func (jwt *Jwt) parseKey(alg jwa.SignatureAlgorithm) error {
	var p *pem.Block
	var p1 []byte
	var err error

	switch alg {
	case jwa.RS256, jwa.RS384, jwa.RS512, jwa.PS256, jwa.PS384, jwa.PS512:
		p, p1 = pem.Decode([]byte(jwt.Key))
		if len(p1) > 0 {
			return errors.New(fmt.Sprintf(pemOverflow, jwt.Name))
		}
		if p.Type != "PUBLIC KEY" && p.Type != "RSA PUBLIC KEY" && p.Type != "CERTIFICATE" {
			return errors.New(fmt.Sprintf(pemTypeBad, jwt.Name))
		}

		switch p.Type {
		case "PUBLIC KEY", "RSA PUBLIC KEY":
			var pub interface{}
			pub, err = x509.ParsePKIXPublicKey(p.Bytes)
			switch pub.(type) {
			case *rsa.PublicKey:
				jwt.RSAPublic.Upsert(
					KidPair{
						Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
						Key: pub.(*rsa.PublicKey),
					})
			default:
				return errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
			}
		case "CERTIFICATE":
			var cert interface{}
			cert, err = x509.ParseCertificate(p.Bytes)
			switch cert.(type) {
			case *x509.Certificate:
				key := cert.(*x509.Certificate).PublicKey
				switch key.(type) {
				case *rsa.PublicKey:
					jwt.RSAPublic.Upsert(
						KidPair{
							Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
							Key: key.(*rsa.PublicKey),
						})
				default:
					return errors.New(fmt.Sprintf(pemRsaNotFound, jwt.Name))
				}
			default:
				return errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
			}
		}

	case jwa.HS256, jwa.HS384, jwa.HS512:
		if len(jwt.Key) > 0 {
			jwt.Secret.Upsert(
				KidPair{
					Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
					Key: []byte(jwt.Key),
				})
		} else {
			err = errors.New("jwt secret not found, check your configuration")
		}

	case jwa.ES256, jwa.ES384, jwa.ES512:
		p, p1 = pem.Decode([]byte(jwt.Key))
		if len(p1) > 0 {
			return errors.New(fmt.Sprintf(pemOverflow, jwt.Name))
		}
		if p.Type != "PUBLIC KEY" && p.Type != "CERTIFICATE" {
			return errors.New(fmt.Sprintf(pemTypeBad, jwt.Name))
		}

		switch p.Type {
		case "PUBLIC KEY":
			var pub interface{}
			pub, err = x509.ParsePKIXPublicKey(p.Bytes)
			switch pub.(type) {
			case *ecdsa.PublicKey:
				parsed := pub.(*ecdsa.PublicKey)
				err = jwt.checkECDSABitSize(alg, parsed)
				if err == nil {
					jwt.ECDSAPublic.Upsert(
						KidPair{
							Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
							Key: parsed,
						})
				}
			default:
				return errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
			}
		case "CERTIFICATE":
			var cert interface{}
			cert, err = x509.ParseCertificate(p.Bytes)
			switch cert.(type) {
			case *x509.Certificate:
				key := cert.(*x509.Certificate).PublicKey
				switch key.(type) {
				case *ecdsa.PublicKey:
					parsed := key.(*ecdsa.PublicKey)
					err = jwt.checkECDSABitSize(alg, parsed)
					if err == nil {
						jwt.ECDSAPublic.Upsert(
							KidPair{
								Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
								Key: parsed,
							})
					}
				default:
					return errors.New(fmt.Sprintf(pemEcdsaNotFound, jwt.Name))
				}
			default:
				return errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
			}
		}

	case jwa.NoSignature:
		if len(jwt.Key) > 0 {
			return errors.New(fmt.Sprintf("jwt [%s] none type signature does not allow key data, check your configuration", jwt.Name))
		}

	default:
		return errors.New(fmt.Sprintf(keyTypeInvalid, jwt.Name, validAlg))
	}

	log.Debug().Msgf("jwt [%s] successfully parsed %s key", jwt.Name, alg)

	return err
}

func (jwt *Jwt) checkECDSABitSize(alg jwa.SignatureAlgorithm, parsed *ecdsa.PublicKey) error {
	bitsize := parsed.Curve.Params().BitSize

	var err error
	if alg == jwa.ES256 && (bitsize != 256) {
		err = errors.New(fmt.Sprintf(ecdsaKeySizeBad, jwt.Name, alg, bitsize))
	} else if alg == jwa.ES384 && (bitsize != 384) {
		err = errors.New(fmt.Sprintf(ecdsaKeySizeBad, jwt.Name, alg, bitsize))
	} else if alg == jwa.ES512 && (bitsize != 521) {
		err = errors.New(fmt.Sprintf(ecdsaKeySizeBad, jwt.Name, alg, bitsize))
	}
	return err
}
