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

func (jwt Jwt) validate() error {
	var err error
	var p *pem.Block
	var p1 []byte
	alg := *new(jwa.SignatureAlgorithm)
	err = alg.Accept(jwt.Alg)

	switch alg {
	case jwa.RS256, jwa.RS384, jwa.RS512, jwa.PS256, jwa.PS384, jwa.PS512:
		p, p1 = pem.Decode([]byte(jwt.Key))
		if len(p1) > 0 {
			err = errors.New(fmt.Sprintf("jwt key [%s] only type PUBLIC KEY allowed but found more data, check your PEM block", jwt.Name))
			return err
		}
		if p.Type != "PUBLIC KEY" && p.Type != "RSA PUBLIC KEY" {
			err = errors.New(fmt.Sprintf("jwt key [%s] is not of type PUBLIC KEY || RSA PUBLIC KEY, check your PEM Block preamble", jwt.Name))
			return err
		}
		var pub interface{}
		pub, err = x509.ParsePKIXPublicKey(p.Bytes)
		switch pub.(type) {
		case *rsa.PublicKey:
			jwt.RSAPublic = pub.(*rsa.PublicKey)
		default:
			err = errors.New(fmt.Sprintf("jwt key [%s] is not of type RSA PUBLIC KEY, check your PEM Block", jwt.Name))
		}

	case jwa.HS256, jwa.HS384, jwa.HS512:
	case jwa.ES256, jwa.ES384, jwa.ES512:
	case jwa.NoSignature:
		if len(jwt.Key) > 0 {
			jwt.Secret = jwt.Key
		}
	default:
		err = errors.New("unable to determine key type, not one of: [RS256, RS384, RS512, PS256, PS384, PS512, HS256, HS384, HS512, ES256, ES384, ES512, none]")
	}

	return err
}
