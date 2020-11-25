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

type RSAKidPair struct {
	Kid string
	Key *rsa.PublicKey
}

type ECDSAKidPair struct {
	Kid string
	Key *ecdsa.PublicKey
}

//uh oh
type SecretKidPair struct {
	Kid string
	Key []byte
}

type Jwt struct {
	Name string
	Alg  string
	// Jwt key supports pem encoding for public keys, certificates unencoded secrets for hmac.
	Key string
	// JwksUrl loads remotely.
	JwksUrl               string
	RSAPublic             []RSAKidPair
	ECDSAPublic           []ECDSAKidPair
	Secret                []SecretKidPair
	AcceptableSkewSeconds string
}

const pemOverflow = "jwt key [%s] only type PUBLIC KEY or CERTIFICATE allowed but found additional or invalid data, check your PEM block"
const pemTypeBad = "jwt key [%s] is not of type PUBLIC KEY or CERTIFICATE, check your PEM Block preamble"
const pemAsn1Bad = "jwt key [%s] asn data not valid, check your PEM Block"
const pemRsaNotFound = "jwt key [%s] RSA public key not found in your certificate, check your PEM Block"
const pemEcdsaNotFound = "jwt key [%s] ECDSA public key not found in your certificate, check your PEM Block"
const keyTypeInvalid = "unable to determine key type, not one of: [RS256, RS384, RS512, PS256, PS384, PS512, HS256, HS384, HS512, ES256, ES384, ES512, none]"
const ecdsaKeySizeBad = "jwt [%s] invalid key size for alg %s, parsed bitsize %d, check your configuration"

func (jwt *Jwt) validate() error {
	var err error
	alg := *new(jwa.SignatureAlgorithm)
	err = alg.Accept(jwt.Alg)

	if len(jwt.Name) == 0 {
		return errors.New("invalid jwt name not specified")
	}

	if len(jwt.Alg) > 0 && len(jwt.JwksUrl) > 0 {
		return errors.New(fmt.Sprintf("invalid jwt [%s] do not specify alg with jwksUrl which contains algorithm(s).", jwt.Name))
	}

	if alg == jwa.NoSignature && len(jwt.Key) > 0 {
		return errors.New(fmt.Sprintf("jwt [%s] none type signature does not allow key data, check your configuration", jwt.Name))
	}

	if alg != jwa.NoSignature && len(jwt.Key) == 0 && len(jwt.JwksUrl) == 0 {
		err = errors.New(fmt.Sprintf("unable to validate jwt [%s] must specify one of key or jwksUrl", jwt.Name))
	}

	if len(jwt.AcceptableSkewSeconds) > 0 {
		secs, nonnumeric := strconv.Atoi(jwt.AcceptableSkewSeconds)
		if nonnumeric != nil || secs < 0 {
			err = errors.New(fmt.Sprintf("invalid jwt [%s] acceptable skew seconds, must be 0 or greater, was %s", jwt.Name, jwt.AcceptableSkewSeconds))
			return err
		}
	} else {
		jwt.AcceptableSkewSeconds = "120"
	}

	if len(jwt.Key) > 0 {
		err = jwt.parseKey(alg)
	} else if len(jwt.JwksUrl) > 0 {
		err = jwt.loadJwks()
	}

	return err
}

func (jwt *Jwt) loadJwks() error {
	keyset, err := jwk.Fetch(jwt.JwksUrl)
	if err == nil {
		log.Debug().Msgf("fetched %d jwk keys from %s", keyset.Len(), jwt.JwksUrl)
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
				jwt.RSAPublic = []RSAKidPair{
					{
						Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
						Key: pub.(*rsa.PublicKey),
					},
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
				case *rsa.PublicKey:
					jwt.RSAPublic = []RSAKidPair{
						{
							Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
							Key: key.(*rsa.PublicKey),
						},
					}
				default:
					return errors.New(fmt.Sprintf(pemRsaNotFound, jwt.Name))
				}
			default:
				return errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
			}
		}

	case jwa.HS256, jwa.HS384, jwa.HS512:
		if len(jwt.Key) > 0 {
			jwt.Secret = []SecretKidPair{
				{
					Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
					Key: []byte(jwt.Key),
				},
			}
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
					jwt.ECDSAPublic = []ECDSAKidPair{
						{
							Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
							Key: parsed,
						},
					}
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
						jwt.ECDSAPublic = []ECDSAKidPair{
							{
								Kid: fmt.Sprintf("%s-%s", alg, uuid.New()),
								Key: parsed,
							},
						}
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
			return errors.New("none type signature does not allow key data, check your configuration")
		}

	default:
		return errors.New(keyTypeInvalid)
	}
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
