package j8a

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
)

type Jwt struct {
	Name        string
	Alg         string
	Key         string
	RSAPublic   *rsa.PublicKey
	ECDSAPublic *ecdsa.PublicKey
	Secret      string
}

const pemOverflow = "jwt key [%s] only type PUBLIC KEY allowed but found additional or invalid data, check your PEM block"
const pemTypeBad = "jwt key [%s] is not of type PUBLIC KEY, check your PEM Block preamble"
const pemAsn1Bad = "jwt key [%s] is not of type RSA PUBLIC KEY, check your PEM Block"
const keyTypeInvalid = "unable to determine key type, not one of: [RS256, RS384, RS512, PS256, PS384, PS512, HS256, HS384, HS512, ES256, ES384, ES512, none]"
const ecdsaKeySizeBad = "jwt [%s] invalid key size for alg %s, parsed bitsize %d, check your configuration"

func (jwt *Jwt) validate() error {
	var err error
	var p *pem.Block
	var p1 []byte
	alg := *new(jwa.SignatureAlgorithm)
	err = alg.Accept(jwt.Alg)

	switch alg {
	case jwa.RS256, jwa.RS384, jwa.RS512, jwa.PS256, jwa.PS384, jwa.PS512:
		p, p1 = pem.Decode([]byte(jwt.Key))
		if len(p1) > 0 {
			err = errors.New(fmt.Sprintf(pemOverflow, jwt.Name))
			return err
		}
		if p.Type != "PUBLIC KEY" && p.Type != "RSA PUBLIC KEY" {
			err = errors.New(fmt.Sprintf(pemTypeBad, jwt.Name))
			return err
		}
		var pub interface{}
		pub, err = x509.ParsePKIXPublicKey(p.Bytes)
		switch pub.(type) {
		case *rsa.PublicKey:
			jwt.RSAPublic = pub.(*rsa.PublicKey)
		default:
			err = errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
		}

	case jwa.HS256, jwa.HS384, jwa.HS512:
		if len(jwt.Key) > 0 {
			jwt.Secret = jwt.Key
		} else {
			err = errors.New("jwt secret not found, check your configuration")
		}

	case jwa.ES256, jwa.ES384, jwa.ES512:
		p, p1 = pem.Decode([]byte(jwt.Key))
		if len(p1) > 0 {
			err = errors.New(fmt.Sprintf(pemOverflow, jwt.Name))
			return err
		}
		if p.Type != "PUBLIC KEY" {
			err = errors.New(fmt.Sprintf(pemTypeBad, jwt.Name))
			return err
		}
		var pub interface{}
		pub, err = x509.ParsePKIXPublicKey(p.Bytes)
		switch pub.(type) {
		case *ecdsa.PublicKey:
			parsed := pub.(*ecdsa.PublicKey)
			bitsize := parsed.Curve.Params().BitSize
			if alg == jwa.ES256 && (bitsize != 256) {
				err = errors.New(fmt.Sprintf(ecdsaKeySizeBad, jwt.Name, alg, bitsize))
			} else if alg == jwa.ES384 && (bitsize != 384) {
				err = errors.New(fmt.Sprintf(ecdsaKeySizeBad, jwt.Name, alg, bitsize))
			} else if alg == jwa.ES512 && (bitsize != 521) {
				err = errors.New(fmt.Sprintf(ecdsaKeySizeBad, jwt.Name, alg, bitsize))
			} else {
				jwt.ECDSAPublic = parsed
			}
		default:
			err = errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
		}

	case jwa.NoSignature:
		if len(jwt.Key) > 0 {
			err = errors.New("none type signature does not allow key data, check your configuration")
		}

	default:
		err = errors.New(keyTypeInvalid)
	}

	return err
}
