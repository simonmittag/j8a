package j8a

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/itchyny/gojq"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/semaphore"
	"strconv"
	"strings"
	"time"
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
	Claims                []string
	claimsVal             []*gojq.Code
	lock                  *semaphore.Weighted
	updateCount           int
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
const jwksRefreshSlowwait = time.Second * 10

func NewJwt(name string, alg string, key string, jwksUrl string, acceptableSkewSeconds string, claims ...string) *Jwt {
	jwt := Jwt{
		Name:                  name,
		Alg:                   alg,
		Key:                   key,
		JwksUrl:               jwksUrl,
		AcceptableSkewSeconds: acceptableSkewSeconds,
		Claims:                claims,
		updateCount:           0,
	}

	jwt.Init()
	return &jwt
}

// we need this separate because the JSON unmarshaller creates this object without asking us.
func (jwt *Jwt) Init() {
	jwt.RSAPublic = make([]KidPair, 0)
	jwt.ECDSAPublic = make([]KidPair, 0)
	jwt.Secret = make([]KidPair, 0)
	jwt.lock = semaphore.NewWeighted(1)
	jwt.claimsVal = make([]*gojq.Code, 0)
}

func (j *Jwt) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case map[string]interface{}:
		if v["acceptableSkewSeconds"] != nil {
			j.AcceptableSkewSeconds = fmt.Sprintf("%v", v["acceptableSkewSeconds"])
		}
		if v["alg"] != nil {
			j.Alg = fmt.Sprintf("%v", v["alg"])
		}
		if v["key"] != nil {
			j.Key = fmt.Sprintf("%v", v["key"])
		}
		if v["jwksUrl"] != nil {
			j.JwksUrl = fmt.Sprintf("%v", v["jwksUrl"])
		}
		if v["claims"] != nil {
			vc, ok := v["claims"].([]interface{})
			if !ok {
				return fmt.Errorf("unexpected JSON value type: %T", value)
			}
			for _, v1 := range vc {
				s, ok := v1.(string)
				if ok {
					j.Claims = append(j.Claims, s)
				} else {
					return fmt.Errorf("unexpected JSON value type: %T", value)
				}
			}
		}

	default:
		return fmt.Errorf("unexpected JSON value type: %T", value)
	}

	return nil
}

func (jwt *Jwt) Validate() error {
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
		return errors.New(fmt.Sprintf(missingKeyOrJwks, jwt.Name, alg))
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

	if len(jwt.Claims) > 0 {
		jwt.claimsVal = make([]*gojq.Code, len(jwt.Claims))
		for i, claim := range jwt.Claims {

			//poor mans jq query conversion
			if len(claim) > 0 &&
				!strings.Contains(claim, " ") &&
				string(claim[0]) != "." {
				claim = "." + claim
				jwt.Claims[i] = claim
			}

			if len(claim) > 0 {
				q, e := gojq.Parse(claim)
				if e != nil {
					err = e
					break
				} else {
					var c *gojq.Code
					c, err = gojq.Compile(q)
					if err == nil {
						jwt.claimsVal[i] = c
					} else {
						break
					}
				}
			}
		}
	}

	if len(jwt.Key) > 0 {
		err = jwt.parseKey(alg)
	} else if len(jwt.JwksUrl) > 0 {
		err = jwt.LoadJwks()
	}

	return err
}

// TODO this method needs a refactor and has high cognitive complexity
func (jwt *Jwt) LoadJwks() error {
	var err error

	//acquires the lock with true else skips
	if jwt.lock.TryAcquire(1) {
		var keyset jwk.Set
		keyset, err = jwk.Fetch(context.Background(), jwt.JwksUrl)
		if err == nil {
			log.Info().Msgf("jwt [%s] fetched %d jwk from jwks URL %s", jwt.Name, keyset.Len(), jwt.JwksUrl)
		} else {
			log.Warn().Msgf("jwt [%s] unable to fetch jwk from jwks URL %s, cause: %v", jwt.Name, jwt.JwksUrl, err)
		}

		if keyset == nil || keyset.Len() == 0 {
			err = errors.New(fmt.Sprintf("jwt [%s] unable to parse keys in keyset", jwt.Name))
		} else {
			keys := keyset.Iterate(context.Background())
		Keyrange:
			for keys.Next(context.Background()) {
				key := keys.Pair().Value.(jwk.Key)
				alg := *new(jwa.SignatureAlgorithm)
				err = alg.Accept(key.Algorithm())

				//check alg conforms to what's configured. J8a does not support rotating key algos for security.
				if jwt.Alg != key.Algorithm() {
					msg := "jwt [%s] key algorithm [%s] in jwks keyset does not match configured alg [%s]."
					err = errors.New(fmt.Sprintf(msg, jwt.Name, key.Algorithm(), jwt.Alg))
					log.Warn().
						Str("jwt", jwt.Name).
						Str("jwtAlg", jwt.Alg).
						Str("keyAlg", key.Algorithm()).
						Msgf(msg, jwt.Name, key.Algorithm(), jwt.Alg)
				}

				if err == nil {
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
					log.Info().Msgf("jwt [%s] successfully parsed %s key from remote jwk", jwt.Name, alg)
				} else {
					break Keyrange
				}
			}
		}

		//slow down JWKS updates to once every 10 seconds per route to prevent DOS attacks
		if jwt.updateCount > 0 {
			time.Sleep(jwksRefreshSlowwait)
		}
		jwt.updateCount++
		//release here, don't use defer
		jwt.lock.Release(1)
	} else {
		log.Info().
			Str("jwt", jwt.Name).
			Msgf("jwt [%s] already updating within 10s, skipping attempt.", jwt.Name)
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

	log.Info().Msgf("jwt [%s] successfully parsed %s key", jwt.Name, alg)

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

func (jwt *Jwt) hasMandatoryClaims() bool {
	return len(jwt.Claims) > 0 && len(jwt.Claims[0]) > 0
}
