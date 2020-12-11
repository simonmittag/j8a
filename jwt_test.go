package j8a

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwt"
	"testing"
)

func TestLoadEs256WithJwks(t *testing.T) {
	cfg := Jwt{
		Name:                  "MyJwks",
		Alg:                   "ES256",
		Key:                   "",
		JwksUrl:               "http://localhost:60083/mse6/jwkses256",
		AcceptableSkewSeconds: "",
	}
	err := cfg.LoadJwks()

	if err != nil {
		t.Errorf("ECDSA key loading failed via JWKS, cause: %v", err)
	}

	if cfg.ECDSAPublic == nil {
		t.Errorf("ECDSA key not loaded")
	}
}

func TestFailJwksKeySetAlg(t *testing.T) {
	cfg := Jwt{
		Name:                  "MyJwks",
		Alg:                   "HS256",
		Key:                   "",
		JwksUrl:               "https://j8a.au.auth0.com/.well-known/jwks.json",
		AcceptableSkewSeconds: "",
	}
	err := cfg.LoadJwks()
	if err == nil {
		t.Errorf("expected remote RS256 key to fail HS256 JWKS config but received nil err")
	} else {
		t.Logf("normal. validation failed with msg %v", err)
	}
}

func TestFailJwksBadUrl(t *testing.T) {
	cfg := Jwt{
		Name:                  "MyJwks",
		Alg:                   "HS256",
		Key:                   "",
		JwksUrl:               "https://asldkjflkaj394028094832sjlflalsdkfjsd.com.blah/.well-known/jwks.json",
		AcceptableSkewSeconds: "",
	}
	err := cfg.LoadJwks()
	if err == nil {
		t.Errorf("expected remote bad Jwks URL to fail but received nil err")
	} else {
		t.Logf("normal. validation failed with msg %v", err)
	}
}


func TestFindJwtInKeySet(t *testing.T) {
	cfg := Jwt{
		Name:                  "MyJwks",
		Alg:                   "RS256",
		Key:                   "",
		JwksUrl:               "https://j8a.au.auth0.com/.well-known/jwks.json",
		AcceptableSkewSeconds: "",
	}
	cfg.LoadJwks()

	if cfg.RSAPublic == nil {
		t.Errorf("RSA key not loaded")
	}

	key := cfg.RSAPublic.Find("mFtH_0S8dpkKl-rr1T6VS")
	if key == nil {
		t.Errorf("RSA key not loaded for key id mFtH_0S8dpkKl-rr1T6VS")
	}

	key2 := cfg.RSAPublic.Find("ZJKp915ThUboHOq79JJsG")
	if key2 == nil {
		t.Errorf("RSA key 2 not loaded for key id ZJKp915ThUboHOq79JJsG")
	}
}

func jwtValErr(t *testing.T, jwt Jwt, want error) {
	got := jwt.validate()
	if want != nil {
		if got == nil || got.Error() != want.Error() {
			t.Errorf("%s key validation error, want [%v], got [%v]", jwt.Alg, want, got)
		}
	} else {
		if got != want {
			t.Errorf("%s key validation error, want [%v], got [%v]", jwt.Alg, want, got)
		}
	}
}

func jwtPass(t *testing.T, alg string, key string) {
	jwtValErr(t, Jwt{
		Name: "namer",
		Alg:  alg,
		Key:  key,
	}, nil)
}

func jwtBadAlg(t *testing.T, alg string, key string) {
	jwt := Jwt{
		Name: "namer",
		Alg:  alg,
		Key:  key,
	}
	var want = errors.New(fmt.Sprintf(unknownAlg, jwt.Name, alg, validAlg))
	jwtValErr(t, jwt, want)
}

func jwtBadKeyPreamble(t *testing.T, alg string, key string) {
	jwt := Jwt{
		Name: "namer",
		Alg:  alg,
		Key:  key,
	}
	var want = errors.New(fmt.Sprintf(pemOverflow, jwt.Name))
	jwtValErr(t, jwt, want)
}

func jwtBadKeyAsn1(t *testing.T, alg string, key string) {
	jwt := Jwt{
		Name: "namer",
		Alg:  alg,
		Key:  key,
	}
	var want = errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
	jwtValErr(t, jwt, want)
}

func jwtBadKeySize(t *testing.T, alg string, key string, badKeySize int) {
	jwt := Jwt{
		Name: "namer",
		Alg:  alg,
		Key:  key,
	}
	var want = errors.New(fmt.Sprintf(ecdsaKeySizeBad, jwt.Name, jwt.Alg, badKeySize))
	jwtValErr(t, jwt, want)
}

func TestJwtNonePass(t *testing.T) {
	jwtPass(t, "none", "")
}

func TestJwtNoneFailWithKeyProvided(t *testing.T) {
	jwt := Jwt{
		Name: "noner",
		Alg:  "none",
		Key:  "keydata",
	}
	jwtValErr(t, jwt, errors.New("jwt [noner] none type signature does not allow key data, check your configuration"))
}

func TestJwtFailWithoutNameProvided(t *testing.T) {
	jwt := Jwt{
		Alg: "HS256",
		Key: "keydata",
	}
	jwtValErr(t, jwt, errors.New("invalid jwt name not specified"))
}

func TestJwtRS256Pass(t *testing.T) {
	jwtPass(t, "RS256", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS256PassFromX509(t *testing.T) {
	jwtPass(t, "RS256", "-----BEGIN CERTIFICATE-----\nMIIEajCCAlICAhAFMA0GCSqGSIb3DQEBCwUAMHYxCzAJBgNVBAYTAkFVMQwwCgYD\nVQQIDANOU1cxDTALBgNVBAoMBG15Y2ExGjAYBgNVBAsMEW15Y2EgaW50ZXJtZWRp\nYXRlMS4wLAYDVQQDDCVteSBjZXJ0aWZpY2F0ZSBhdXRob3JpdHkgaW50ZXJtZWRp\nYXRlMCAXDTIwMTEyMTIzMDgyOVoYDzMwNDcwODEwMjMwODI5WjB9MQswCQYDVQQG\nEwJBVTEMMAoGA1UECAwDTlNXMQ8wDQYDVQQHDAZTeWRuZXkxEDAOBgNVBAoMB3Jz\nYXRlc3QxCzAJBgNVBAsMAm91MRAwDgYDVQQDDAdyc2F0ZXN0MR4wHAYJKoZIhvcN\nAQkBFg9tYWlsQHJzYXRlc3QuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\nAoIBAQC4WvALvGp6TeUDtOREwGvuShqZMITkMXRTFU93QVuaWgZOSQstdbdHp87G\n5uLw7Y9eT+035N8lmI5izBQT/6eBHuARct5uiwu1V2ZsatpMbC2gaalJ7cVGv1mZ\nC7vilmwDPCplUyrPoIg1zu6/FMgON33w5FSue7GqAcT1jP+6jaq+do6BYhcqGQOQ\nyoDw7QWT+RoJ955lAw3vQ6cLAhL4s8UTy5gYh1toQiSNl8zGb+I1EQ05Xzj2/A6I\n65au4LKn5gypNW1xbj1DWZxl/IZPsHkYdZ7sy3GpA2F/WbY2+jWb8u9otqn8Qzfr\nzjE5rpF7EHVd0Hyh6CkxHKA/HuivAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAI2f\ns0I2it+M5AHPmqs2HJcW8uL7utGf+X+uDyX14BIkJNeNCxmoWDlahvoS+cviyA6Q\nAihg1ECgPAjNweKIVhOKV5462INMbcmqwQqaL4gHd6IVAAfXOLO6Eg2GlhP/NQj3\n8KH6rAUXLOe2pAboR6+j6Jfz53mYUJu2sYnmcduwrjXva8BPqHwH70aFQ81koRIz\nyqvYfxfg2QLO1IyEIAC93D6Pi4fYHGJCM9HMMeIVbbcnUjWTeSL4jfaIyWhbyry6\nLpKXSCAS+VPqE8qI88J6NxkjQ9Q7B5QOL4K+gserqzOryUjH0Bl6l2K2g3xQcB63\nB8/vFbqQTXkzNcLJch/KIPvw5N+WFi3DBxkM07K+Hettn7vdR0tHmty1VG5k5zl+\nL1rp9vBAXjhlpiOao7qZBHfReO7FIJjeQ+TqXx9jJIUFyKqoEFVjse/2gMtZG4Z9\nbp1y31IZUHPyBfeZDQvzEgQeuAjMPQpikPqr1KupA5cDQ5K5icWVDwFLfHj143Yn\nswZG+5zzuhsLs+9unyfJhkhzbxZg6rznnssUILcWRpgEzsgnMx0Hzk2ydJZLTz9z\nUF0YH+65vYC/x84tb1TiZhN9gkP9rbNwr12AIWmMRvxhVjIGm1S6MbKWVotNvUOq\nuBNo8nqxAmGBsH/sytKm9cT0v/FBF2edE6LThwVg\n-----END CERTIFICATE-----\n")
}

func TestJwtRS256FailFromBadX509(t *testing.T) {
	jwtBadKeyPreamble(t, "RS256", "-----BXXXEGIN CERTIFICATE-----\nMIIEajCCAlICAhAFMA0GCSqGSIb3DQEBCwUAMHYxCzAJBgNVBAYTAkFVMQwwCgYD\nVQQIDANOU1cxDTALBgNVBAoMBG15Y2ExGjAYBgNVBAsMEW15Y2EgaW50ZXJtZWRp\nYXRlMS4wLAYDVQQDDCVteSBjZXJ0aWZpY2F0ZSBhdXRob3JpdHkgaW50ZXJtZWRp\nYXRlMCAXDTIwMTEyMTIzMDgyOVoYDzMwNDcwODEwMjMwODI5WjB9MQswCQYDVQQG\nEwJBVTEMMAoGA1UECAwDTlNXMQ8wDQYDVQQHDAZTeWRuZXkxEDAOBgNVBAoMB3Jz\nYXRlc3QxCzAJBgNVBAsMAm91MRAwDgYDVQQDDAdyc2F0ZXN0MR4wHAYJKoZIhvcN\nAQkBFg9tYWlsQHJzYXRlc3QuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\nAoIBAQC4WvALvGp6TeUDtOREwGvuShqZMITkMXRTFU93QVuaWgZOSQstdbdHp87G\n5uLw7Y9eT+035N8lmI5izBQT/6eBHuARct5uiwu1V2ZsatpMbC2gaalJ7cVGv1mZ\nC7vilmwDPCplUyrPoIg1zu6/FMgON33w5FSue7GqAcT1jP+6jaq+do6BYhcqGQOQ\nyoDw7QWT+RoJ955lAw3vQ6cLAhL4s8UTy5gYh1toQiSNl8zGb+I1EQ05Xzj2/A6I\n65au4LKn5gypNW1xbj1DWZxl/IZPsHkYdZ7sy3GpA2F/WbY2+jWb8u9otqn8Qzfr\nzjE5rpF7EHVd0Hyh6CkxHKA/HuivAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAI2f\ns0I2it+M5AHPmqs2HJcW8uL7utGf+X+uDyX14BIkJNeNCxmoWDlahvoS+cviyA6Q\nAihg1ECgPAjNweKIVhOKV5462INMbcmqwQqaL4gHd6IVAAfXOLO6Eg2GlhP/NQj3\n8KH6rAUXLOe2pAboR6+j6Jfz53mYUJu2sYnmcduwrjXva8BPqHwH70aFQ81koRIz\nyqvYfxfg2QLO1IyEIAC93D6Pi4fYHGJCM9HMMeIVbbcnUjWTeSL4jfaIyWhbyry6\nLpKXSCAS+VPqE8qI88J6NxkjQ9Q7B5QOL4K+gserqzOryUjH0Bl6l2K2g3xQcB63\nB8/vFbqQTXkzNcLJch/KIPvw5N+WFi3DBxkM07K+Hettn7vdR0tHmty1VG5k5zl+\nL1rp9vBAXjhlpiOao7qZBHfReO7FIJjeQ+TqXx9jJIUFyKqoEFVjse/2gMtZG4Z9\nbp1y31IZUHPyBfeZDQvzEgQeuAjMPQpikPqr1KupA5cDQ5K5icWVDwFLfHj143Yn\nswZG+5zzuhsLs+9unyfJhkhzbxZg6rznnssUILcWRpgEzsgnMx0Hzk2ydJZLTz9z\nUF0YH+65vYC/x84tb1TiZhN9gkP9rbNwr12AIWmMRvxhVjIGm1S6MbKWVotNvUOq\nuBNo8nqxAmGBsH/sytKm9cT0v/FBF2edE6LThwVg\n-----END CERTIFICATE-----\n")
}

func TestJwtRS256BadAlg(t *testing.T) {
	jwtBadAlg(t, "RS257", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS256BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "RS256", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS256BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "RS256", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS384Pass(t *testing.T) {
	jwtPass(t, "RS384", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS384BadAlg(t *testing.T) {
	jwtBadAlg(t, "RS385", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS384BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "RS384", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS384BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "RS384", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS512Pass(t *testing.T) {
	jwtPass(t, "RS512", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS512BadAlg(t *testing.T) {
	jwtBadAlg(t, "RS513", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS512BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "RS512", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS512BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "RS512", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS256Pass(t *testing.T) {
	jwtPass(t, "PS256", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS256BadAlg(t *testing.T) {
	jwtBadAlg(t, "PS257", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS256BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "PS256", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS256BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "PS256", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS384Pass(t *testing.T) {
	jwtPass(t, "PS384", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS384BadAlg(t *testing.T) {
	jwtBadAlg(t, "PS385", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS384BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "PS384", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS384BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "PS384", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS512Pass(t *testing.T) {
	jwtPass(t, "PS512", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS512BadAlg(t *testing.T) {
	jwtBadAlg(t, "PS513", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS512BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "PS512", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS512BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "PS512", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES256Pass(t *testing.T) {
	jwtPass(t, "ES256", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES256PassFromX509(t *testing.T) {
	jwtPass(t, "ES256", "-----BEGIN CERTIFICATE-----\nMIIDkzCCAXsCAhAEMA0GCSqGSIb3DQEBCwUAMHYxCzAJBgNVBAYTAkFVMQwwCgYD\nVQQIDANOU1cxDTALBgNVBAoMBG15Y2ExGjAYBgNVBAsMEW15Y2EgaW50ZXJtZWRp\nYXRlMS4wLAYDVQQDDCVteSBjZXJ0aWZpY2F0ZSBhdXRob3JpdHkgaW50ZXJtZWRp\nYXRlMCAXDTIwMTEyMTIyNDUzNVoYDzMwNDcwODEwMjI0NTM1WjBxMQswCQYDVQQG\nEwJBVTEMMAoGA1UECAwDTlNXMQ8wDQYDVQQHDAZTeWRuZXkxDDAKBgNVBAoMA2o4\nYTELMAkGA1UECwwCb3UxDDAKBgNVBAMMA2o4YTEaMBgGCSqGSIb3DQEJARYLbWFp\nbEBqOGEuaW8wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARBDCsRRZCV7vmbP9P7\nC3h9egCff0yCYJ1ndljd9ciSpAM46rYknfHpFrLLUZOnueU3zwmrFfxQLvUt2Zsj\nIqLqMA0GCSqGSIb3DQEBCwUAA4ICAQASQ+VgKa2tUpDCL8hXCEjQT+LAVO3jxsDQ\nEj14R/W1hfxJ5UrpA39leCjWIa3ZE5rLJjbxg+/K5PvjbZhIXRVp49XoFdkcxD8Y\nMyBnBPSYuXht/jgzvVVHEvyQJ9xW3Yx2jo53UwMs85+fgZA/lUdcoCZ/j7iaDHkl\nGhE8OhPeHjzye98VZrs5IcTaZ3aBpE3atvvgzzb5xj+CwEdd07aRFdJ8Jez4u57x\nXH36omDUE8zHn2Lfr44e9AsszAprhqLdyRw3y1q7KMysxxLBmLibe7Km4rfYE0wN\nLCnSjaiXdZovOQNjNV+yPKhykWNw5mlWo9XcPcWx7hrl3ik380yK1cD6edrTev9R\nGGmbVgYfVXunvFC4w+xEgQzLA9Dc/hus/9GO+trwbVjRSoWEi4HUE2rF0eYpLk0m\nzvlvL0+JjVXDuWC3FBgAJwBLbEdrtD4dQeU8N88MZ9WWTBR6TwOvFvO7dVwJCDk0\nThlD/f2RD14vsjHstAU87OL+NTHBefIBXT6J1Ev+HIhpKjyt7zhkHcfxTVzxP2cJ\nk4n8HSUsx52ggLafcK5V+zkGa2uKLNWru5w/pVqrjWV4goWe6k6ZPahzp5riNIlP\nnhD2IZz+y8r7ognXs6+EjYqEJknm/4cMbWudBFz87w36w6gNwiHvW1Q8bkVJ65We\nAWKuTQnosA==\n-----END CERTIFICATE-----\n")
}

func TestJwtES256FailFromX509(t *testing.T) {
	jwtBadKeyPreamble(t, "ES256", "-----BEGXXXXIN CERTIFICATE-----\nMIIDkzCCAXsCAhAEMA0GCSqGSIb3DQEBCwUAMHYxCzAJBgNVBAYTAkFVMQwwCgYD\nVQQIDANOU1cxDTALBgNVBAoMBG15Y2ExGjAYBgNVBAsMEW15Y2EgaW50ZXJtZWRp\nYXRlMS4wLAYDVQQDDCVteSBjZXJ0aWZpY2F0ZSBhdXRob3JpdHkgaW50ZXJtZWRp\nYXRlMCAXDTIwMTEyMTIyNDUzNVoYDzMwNDcwODEwMjI0NTM1WjBxMQswCQYDVQQG\nEwJBVTEMMAoGA1UECAwDTlNXMQ8wDQYDVQQHDAZTeWRuZXkxDDAKBgNVBAoMA2o4\nYTELMAkGA1UECwwCb3UxDDAKBgNVBAMMA2o4YTEaMBgGCSqGSIb3DQEJARYLbWFp\nbEBqOGEuaW8wWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAARBDCsRRZCV7vmbP9P7\nC3h9egCff0yCYJ1ndljd9ciSpAM46rYknfHpFrLLUZOnueU3zwmrFfxQLvUt2Zsj\nIqLqMA0GCSqGSIb3DQEBCwUAA4ICAQASQ+VgKa2tUpDCL8hXCEjQT+LAVO3jxsDQ\nEj14R/W1hfxJ5UrpA39leCjWIa3ZE5rLJjbxg+/K5PvjbZhIXRVp49XoFdkcxD8Y\nMyBnBPSYuXht/jgzvVVHEvyQJ9xW3Yx2jo53UwMs85+fgZA/lUdcoCZ/j7iaDHkl\nGhE8OhPeHjzye98VZrs5IcTaZ3aBpE3atvvgzzb5xj+CwEdd07aRFdJ8Jez4u57x\nXH36omDUE8zHn2Lfr44e9AsszAprhqLdyRw3y1q7KMysxxLBmLibe7Km4rfYE0wN\nLCnSjaiXdZovOQNjNV+yPKhykWNw5mlWo9XcPcWx7hrl3ik380yK1cD6edrTev9R\nGGmbVgYfVXunvFC4w+xEgQzLA9Dc/hus/9GO+trwbVjRSoWEi4HUE2rF0eYpLk0m\nzvlvL0+JjVXDuWC3FBgAJwBLbEdrtD4dQeU8N88MZ9WWTBR6TwOvFvO7dVwJCDk0\nThlD/f2RD14vsjHstAU87OL+NTHBefIBXT6J1Ev+HIhpKjyt7zhkHcfxTVzxP2cJ\nk4n8HSUsx52ggLafcK5V+zkGa2uKLNWru5w/pVqrjWV4goWe6k6ZPahzp5riNIlP\nnhD2IZz+y8r7ognXs6+EjYqEJknm/4cMbWudBFz87w36w6gNwiHvW1Q8bkVJ65We\nAWKuTQnosA==\n-----END CERTIFICATE-----\n")
}

func TestJwtES256BadAlg(t *testing.T) {
	jwtBadAlg(t, "ES257", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES256BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "ES256", "-----BEGIN NOT SO PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES256BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "ES256", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES256BadKeySize(t *testing.T) {
	jwtBadKeySize(t, "ES256", "-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEbzpBvrm4YDC8/AWuFNylOeApmVi0NpiR\nI4tQOE0y5bm7zjwjr/nvwKreCF0mT0NivYOzQfu6D5Wy1Jpgm0ebMDljSKEGkZwC\nP2/jogu9YzmaAnV/Re16ffln/n+tsATG\n-----END PUBLIC KEY-----\n", 384)
}

func TestJwtES384Pass(t *testing.T) {
	jwtPass(t, "ES384", "-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEbzpBvrm4YDC8/AWuFNylOeApmVi0NpiR\nI4tQOE0y5bm7zjwjr/nvwKreCF0mT0NivYOzQfu6D5Wy1Jpgm0ebMDljSKEGkZwC\nP2/jogu9YzmaAnV/Re16ffln/n+tsATG\n-----END PUBLIC KEY-----")
}

func TestJwtES384BadAlg(t *testing.T) {
	jwtBadAlg(t, "ES385", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES384BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "ES384", "-----BEGIN NOT SO PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES384BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "ES384", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES384BadKeySize(t *testing.T) {
	jwtBadKeySize(t, "ES384", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n", 256)
}

func TestJwtES512Pass(t *testing.T) {
	jwtPass(t, "ES512", "-----BEGIN PUBLIC KEY-----\nMIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQAuObw6RyorEJfB+UUgkJcmfdIlLuK\nwO6KFjHF8u5+OPpHNZuq203djMAp07caPZsuEg2M0WIjdm10LefJhnepq/8B1AZT\nI/Kr/XahLw9Wn9/TDWXgLwaPnPk6jnzlkVFpts4mcvS2/ec/yBoPViHZ98sMwOoS\njTrSpJkw0BIo5qQSuGk=\n-----END PUBLIC KEY-----\n")
}

func TestJwtES512BadAlg(t *testing.T) {
	jwtBadAlg(t, "ES513", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES512BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t, "ES512", "-----BEGIN NOT SO PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES512BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "ES512", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtES512BadKeySize(t *testing.T) {
	jwtBadKeySize(t, "ES512", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n", 256)
}

func TestJwtHS256Pass(t *testing.T) {
	jwtPass(t, "HS256", "key")
}

func BenchmarkJwtHS256_256(b *testing.B) {
	DoBenchForAlgBearerAndSecret(b, jwa.HS256, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.OzPJr1hU0ffvHP6OH646IHbLw277gkswa2U-3McxAOM",
		"0123456789abcdef0123456789abcdef")
}

func BenchmarkJwtHS256_512(b *testing.B) {
	DoBenchForAlgBearerAndSecret(b, jwa.HS256, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.wC3uGUE3p1WdS9Yw_NfjZn8F0B-NmB5DT4h8duOE5Gg",
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
}

func BenchmarkJwtHS256_1024(b *testing.B) {
	DoBenchForAlgBearerAndSecret(b, jwa.HS256, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.o2XgUCY5i9BshAWQ4PNjaqYuGMbdvoJ6YD3Vt4fA0ZQ",
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
}

func BenchmarkJwtHS256_2048(b *testing.B) {
	DoBenchForAlgBearerAndSecret(b, jwa.HS256, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.7t8Bbbo520dN4LdsUvL3gEtHrWLSPYPVzS8cAll8S7E",
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
}

func TestJwtHS256BadAlg(t *testing.T) {
	jwtBadAlg(t, "HS257", "key")
}

func TestJwtHS256NoKey(t *testing.T) {
	jwt := Jwt{
		Name: "namer",
		Alg:  "HS256",
		Key:  "",
	}
	jwtValErr(t, jwt, errors.New(fmt.Sprintf(missingKeyOrJwks, "namer", "HS256")))
}

func TestJwtHS384Pass(t *testing.T) {
	jwtPass(t, "HS384", "key")
}

func TestJwtHS384BadAlg(t *testing.T) {
	jwtBadAlg(t, "HS385", "key")
}

func TestJwtHS384NoKey(t *testing.T) {
	jwt := Jwt{
		Name: "namer",
		Alg:  "HS384",
		Key:  "",
	}
	jwtValErr(t, jwt, errors.New(fmt.Sprintf(missingKeyOrJwks, "namer", "HS384")))
}

func TestJwtHS512Pass(t *testing.T) {
	jwtPass(t, "HS512", "key")
}

func TestJwtHS512BadAlg(t *testing.T) {
	jwtBadAlg(t, "HS513", "key")
}

func TestJwtHS512NoKey(t *testing.T) {
	jwt := Jwt{
		Name: "namer",
		Alg:  "HS512",
		Key:  "",
	}
	jwtValErr(t, jwt, errors.New(fmt.Sprintf(missingKeyOrJwks, "namer", "HS512")))
}

func BenchmarkJwtRS256(b *testing.B) {
	bearer := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.aHXyvYSVGeNAB4wig6Nv5--vo4ZWG6CSLlrIFA8Dm_ohaVLiT_2y_cJc_v0kc_bACgdsDJloooAllh75tc870MiqGXJ4enxoUUmIKory4PUooFlFENUiDMS4EQz7JXxxhbjlfTURo00kM9hSSJ8-yluGXaX0Xirc15H2oMBDUq-1nUCjECsI6T7P4rSYxrZOviK58PjKuyocvHm_3Q0nAY0NHJ9RVYatntJFai75EbLDzzOvTDUXFjVMIQ0-nR-55pKvnykoZoqdaGJInsHHl8-M_fwxeYzvBXd_saBfMI2-KeYLUXTN9_n2z6oey4wapalJavWLxSnH75uUhOlLkU0NQWeylAiEOKXoNLBJwf5QUnwMHgrEvzZSz7-2rYG6_yuAHQ0csHRMVCIG0YKKhvgjis6tXmb4vjiR-DdJ1D2zDXVgjeQCqYdS9ycbx_CgeZ6ignalz85Xu3fm2_UwsdrrSEg9sYzzhzHl9ksD8mxlfUwcywb_HVkz77pciIcx0D8DUFR8JuWxy2AIMSWoYlWag5hVkxhIDSlcNKOuxOitPwyLA2bejxuDhZ71Kh9i_PmTzVYCrnNoNY9CKjWmC2FS_QVoDpCKcdj3T6bclucidGbtdM50fYA47jnJULcqJuaRav7otR71RmxufnrBGI6yxxmEbQPeBBFt_Xj4bRg\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n"
	alg := jwa.RS256

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtRS256_2048(b *testing.B) {
	bearer := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.qZEo64DvgJLg5Pdg92C6uFM-D2MFBLV9wqkS8SMA9Tb3Bh-I1yrUEdTbfZtorkbL3e5ACMjMlgFAQtTmn0PWCpwOvelZ-INharlJCR_Qa3Fru5nEoEzmci63br-2Z7aU2AmiFs9bwDDpekvPJsFOHt2S-6qpIUfckaIUBcyufW09HrSUUpocY_m75VpI4SLc8rLcQC5_Gmkr3rcHbAhfECinDZvuWs_Dph_DV8nO_DYrcEOxnLgeaLPnONYKcao9th97bzpT9_firnSZochqRbzlJ6ORLt1rJWoCI7hEzi6ZUT1lUMedATmxjd4y7L8tJd-kGILpBl5GqTx6yiGPvQ\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuFrwC7xqek3lA7TkRMBr\n7koamTCE5DF0UxVPd0FbmloGTkkLLXW3R6fOxubi8O2PXk/tN+TfJZiOYswUE/+n\ngR7gEXLebosLtVdmbGraTGwtoGmpSe3FRr9ZmQu74pZsAzwqZVMqz6CINc7uvxTI\nDjd98ORUrnuxqgHE9Yz/uo2qvnaOgWIXKhkDkMqA8O0Fk/kaCfeeZQMN70OnCwIS\n+LPFE8uYGIdbaEIkjZfMxm/iNRENOV849vwOiOuWruCyp+YMqTVtcW49Q1mcZfyG\nT7B5GHWe7MtxqQNhf1m2Nvo1m/LvaLap/EM3684xOa6RexB1XdB8oegpMRygPx7o\nrwIDAQAB\n-----END PUBLIC KEY-----\n"
	alg := jwa.RS256

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtRS256_3072(b *testing.B) {
	bearer := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.UnmfhmjLnpZpZQGcdYsmjqB9iV-tMJHimaMDo49JG-CJvYYWoD2i_bixGqK6ZePXW2WuPWHjo60Sah76NQIEP3EWZNTgeR4D3mW_x5slqyq7pC86iwye7gDlHQ82MOz0lz2GgBTs5SIiEwCCuvMTCgweXUyJijxd3CsDSpBec7lbGo3Mho8h-qNwlRaiMa9Ja6K2udJJgquNHYPWM3dDWFjiBwZrwThg8N4CuGvNtnirBKKecs8ZaHK8B1iWo-wDBen9g85JQ7A53KtsBlXJe7DXVGbwsKhbphDMnOTmL0WEalMf7edKR3bOz6AesOVelHDGJ1ueYunbeZiCcEvK-8mMAYWB21g0_xaLs-j0Q0fC-gBx2Dkyhz0T8JekESFju-rA49D1CdbLjRuxO3uuXHFWA-4vHLL8zrDYQgH0GXiw_WG5IUegRZQ2cvNp5nPfqKOFYJbuY7A4HLdXWSjw6OQmq0E0RcuAC-6qYtTy6UWLnJ5rt3q_wpaAJhfHQVTE\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAqKrX9U4Vc79ejAz3TWVL\nSixaZzBLNQv5diPkUAsHd0XuZ3bKmaUfNnJ64nZJ/P5VZXZ6JWi+owbPX6qVAP49\nVskehq4SU1cdPUyjv0/Uf/Pi8YnNZQWS2K3aXRS00iDNuKCMpr/p3YE48mg2/I13\nx5RIZr/UzNVzSE6tLlGEGAJgx47C8RlfAN+EB5xoaozqHUqsABMOMpLmsda1Bt9Y\nhnvcfimp7YOnwDUTVChBr8uy7luCUfMbqz3TZCWgcVaSBk8Sxmv4SkpUr1cReq8o\nzw2cprewF3oCaw1oU+K84hOxiYj9pzLV3izkLO0pDNmKiav9H+OPeZNP0SvCdrsG\nZdfoWDC9oOtI3qRHC3XtDWm8U/b4v9K8LfXAIxVxZQSXvMtIXR0tsBg48yKvtk2A\nOzMv2lIBXGPAeKDnCPkq2jQVBBdrfdIO9bnSreOB98O+gDk7vKPx6ccqwN2NiYid\ngQ7I8gsowa5mV4PDfvZoBOfilUIUZ5owazxl7siQAmehAgMBAAE=\n-----END PUBLIC KEY-----\n"
	alg := jwa.RS256

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtRS256_4096(b *testing.B) {
	bearer := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.UmArmjxT8q1zFMBAXRRezXWE2okJOI8UtpaNvrou2VZsjuVX6KanZAmw5mXVF0wcCeXIysv5v-PlxSYgYJc2fRnOjU62MAEmQQQSQ_qa8Ueah5P9WBN7hYr6gJcLBgjNrq4iaWoKNqy7B1TOzr6ztJeO3ocy8nIvqFE4hF1l4baGOG-B63rW3e6C2MGAv6ucPu1MIUCTCC2b6MLv_C5Q2F0lMJ_4x7-2aJHbPxy77_F_smICK6j0nxutz1uUxxIOdV_uRCkEHfBJ-LsbgmmHQ-MwNUEZBiOouHj1trF6Lgs9w0quCMttRFfRLlBozEUjCrB3MHys1vyvJNWuFBXnvG0M9wIpRAOLsQ3Qw9_1ePHjQTH77YZSCwh5HNJNUuCEh2IpcCR9ZzBAQW3if04r9UPcnFrQwIaa4lk2E_s_anU6-PotLl0JktUkEAfiqTK7eRy9TRxMPqaslUNSS05Qw_qx_AkbVPLw9Jha34fHS4bz6iCzDprEGJKpI5AsQAfpE7d93HYzUwUbFVR5haefT0iC0XC0oC3AwrOOqWHyL0ehHzEktMNO6MtMdNNwMXxZqr2mgGNt-_w9Prgneu0NMlVjVX9t-XbO7_Jk7aQ2Vdx27FgTT10S50tHTdtmMk7Ihs-EQUoU8l5FPGykVV3Eft_rFV9HKsJKFVtEugfgRv8\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA4bwDrdKKCCRpDzMg1sNR\nl6vpoX4ouwWkTjBUzejEf8WNmFvCF4AqjXhrMEkU3nlewTihApeqcYE7yAQfbi0v\nSCWcMx4pYGPR4hhMDsFsQx5x3eNxpFnkT4A1+y0hPSwOLWN9Vc8V4pxWuplV9DD5\nBMMQ14tVhxSDq9A+tbpumFxGVw/NW3bcPCyzSdw8vHXhdT+lEuWjR6d+OaMfblmx\n0pywfdDE37j75cNXoK4N2U57U1mNRq/UfCpNig4PFgf9H8+IvrEvI319TnEqL5mG\n5vXdMUYblxtcxBTrrjgGhhHTAqXP03+Qw4uqjxbmHL9567iDELoEQHIYOcJS7jmK\nMyXfwl+CSRyra3iRlMCPFz/AyY9uL+wuxaMv7t5ayZ6dT2nDlTwispqkq7v05FyF\nSTlmWChM/2zY5owCmmk5/qsbrwbExDL5Gw3n24Qnt2MiaqtIv0XkbQl+pV0dSsGx\nasP4ueCzKfhmhimoY8rMh1mTzIbO0j88md/HecPksf1LVtWQ4YMGsVARHnhtKDKE\ntYKrljf51rAgH9RP68QAoOC21BnTxHhyKGkDSf9bbzIzReEvw8MG+QOuf4iMgkKA\nd0De94zddYgowoWMjXksgw+t+++4nBoSM3b0AZ/jT4iUhOFCmqvPyQ70tQGEiR2/\nOgLWvJGT8k86d0x9Eirsu4UCAwEAAQ==\n-----END PUBLIC KEY-----\n"
	alg := jwa.RS256

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtRS256_8192(b *testing.B) {
	bearer := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.ZyfPZ_KXCilUdLZHWU3X7gTI0pbYl3j5V8Rx1vFVGPbs-LQCDWBFNz_0owauLs57ezC_kg5NKWkAXSqxRUTJGJYDUt6o0vo6pItpIS-zEkzaJmszvOosvf1Eq_8yGyN-fzNYbx0Zj8yY4uNEALITZlFx7o6zzHJQNpTiTlmZjogUKStXzPeQuI3Ro1x_IZo4Emo5BoseEWj60As8XYNu975XTO31YT3EEFkuepB0rWtHZa7xa0fAQ6lMACOkCNRkfT-ffQfGqqCsEDAexHBoLdd2BGW0O7YTo0oI66T_-IdAmo5IyPM99sHh55wA2IJkKDuXw1XWbRSq70zqceMoxXDrLGw1AJ4zeMkW9C7Ej_ZfWEJMegiqvfcvuVdZEL23b_wXHD0yHsZ-p4zo-T6N13qAuEVEHmxTEYJug0k5SkAeChlq1BNixeUYp4k0mpXAyirMWKy0KINsfKCma4JcFUmYvxG_Lb0-CQrYVNpCEi91ayJlHXg96rhjrH_Z3iYtK8uIMk9dFW7v_HnzxbZa6wkzHST2F5oMlXNLcfiyDuozYMxkDavWgZw152LAxN2nfh1Zv4GplcpjMpxLWo_hOj2skquXYujArbMsPMlbk0HZCqaCrTxEt3XMe7gQrPwGg_DF1xVbG7jBE4KHCrVpeauI4GOc0XFnplDasETgDxipUnqPEnjdZeKdaw5nsifRi016f8-kIvB-kINC_AFUXmBzTkPFc1A0NxzodXQcfOgHesndE8trPE7F3XuukF2UDLzNNHZxmDNKhGEMmfJz-ul8Frn5aKPByBD8N7YoeHfl0sCUs8wfv4f-Rm8PiRMSZDsbpjMTr9ihia2Rhkq6rH0oXrE_cwltUnbEj9OFfq8qBWRFYewVdD9jw2ncNuE-WLBku6NPM9p4tFekJ4akUAXzbr0gfkW0HhtW6K3fkXS_PJpP90XN7f6qSzbYLUWoRXdwl62LZ3oxqPtDpuLGnWYObTIufTTAExV6JxvuiZsNO5gigc1pel0g6F88tifxdXxWA6ARi7cPK828Y-Fk8rLtm58ygwkiCt9DCHbCkhzDflyuRQWKC0cJuosGlsOmWe21p_I8thQQQ81aykIXIoSutiGQ2yEDoRD75i4KZTINND4R_UQjLSdfQ0C7LS8YabepaaL8UxDokN6gZCy_sG-AA6T_shMOPsSTBd-iC4Lht7BMXGaC0ASeSPALZjURN39W76XNxglVMBV_eh2_WPRktZIOqq_1IZlfQagvD9XFlKL6BoZnaFJTUfJnYOWHk7F5vfTvYjXA7n_FTnFI7kwQNLjbLrfSl9u6dWxtJiXF4mk0oWeFoei0Q8cRYU5dJNHHN3FFxFIfSNAIXVzb9Q\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIIEIjANBgkqhkiG9w0BAQEFAAOCBA8AMIIECgKCBAEAv+NZY2ssUeFgLB8lB8JX\nCshGXmAOR5BABKMWHZ+of3aKb70JjKCRL0RZByTNa5NRCNNnp/jILAzQpC5Cfswa\nRWhr2wQRdqZgUBK85fHt9zaBj35OP+LcPrWz8U9AbhoED9ZpuHdhgjZImWjBQBcf\nbm8jKlPcv73Atcdb/Tttw7fuodnmKFKWoKCU5D6b7PHVCRhziVfa+V7dZiL8q9Uc\nTzi+MNuoHP8AGe39nBU7E1zr1+L+A068ZBLGZDA/qkRQJOnm17ALbW/C9mSVwrD6\nRPVfNq8wSomTpEN6YTu1av15HJ8ZBjkRK8qX2Lu6KgBlL9nAbSTZb3V21+SSciU/\nW7T5XYntobVJyY5iRLThdetm8nkD1NP2J0SutozMF0ZB5zokRZUOblQFL/5aTIJ/\nGa9yq0kg4A8wIGtRhEZQzKZ7idByAg1YXt4U1xAgwEeuUp/DRnGzhsu9H4Dr1hwu\nhfpA80DcGtbu3ZyA3ghC5LnVrbXdiJkZOj85E1M6dfnd1VXN4790wk6DEUcbiBWR\nz8umyflde9g+aSw6qQEZm+rWrjI6I9ZSqMjnWhraAmlRjRJBaaOYXcKvKgGVvgcr\niJ+igwYVTc4jnWZE9OsQeok6yl/Ic7MU6QIcu/d4YgJi2ubk7unIzyWVM3rAQrzK\nDVPcv8MT+oLgtszI28D3APRjdRYfRG8MtsIQNLV9GPjDyRpJ1iHdah98VIbm+1Zh\nCF66EZaXT43yqy2agTSFPmmldf4OPSL6WCWZ5RVdEHGaUQyojmA0uTKg4PdvpqLA\n/lLB+Qa7dNiIdguMSSPlc/pEJv0iN3SlIpHSL6HwYnlsm00gpU61Gwo6432r1RT3\n5VFHdnH4HyRQKgkFgqUE7zIz43BY57o7ShlOb7O0zqe/JP/qh0wKBqDchVAgLObX\nkWiepf2L+Un3Pi3QZGVPvLQ5f/SAha4EVZyvKK49N0JseUOSyTjfZ95/X1jUlp5Z\nti9ihQUEkhfyvAO57AWo1CYQ26I8zSfP+T67TXIz0hgDfJ20op2oW8w0wkfWfteV\nPmDQWXt0iSOFF9jIEZPlkbas06oC+kHgcwEwdA1SHvEqJwoBGHRTA0JOCma494Nq\nLqXNft1ua3rDJnP3f+3ygEeHhM5HR6gnTASflPBi6GPGb1o2j5qFdHa99qMehZQz\neiNZFDM9NYxz6ooDHEaPWutiapBev0u378vZuN1xYCJW8/3WbimK/4cXSFZc8arI\n0PRpeY/mkVNaqbYCKgycup1NR7v8hXfLwxS/0vL/OnqEbr59/KFeJFlxPvpyC7Hp\nqlZcx0DvzMLAsxWgmvruL+SGtlwknJugqIPBHBhksjSv0l7yeU9SfOTNLfA1Ajfe\nJwIDAQAB\n-----END PUBLIC KEY-----\n"
	alg := jwa.RS256

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtRS256_16384(b *testing.B) {
	bearer := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.ttAo2joM4bBeyAVLbIO12GqCme2hD8iJIoDEucoP-EbckYHmDVnCftUeXtiFiqrnb97RSh16PPY-w71_tUM6wetGPe_DcPVxcJmQOJyLJSwT48Cm4vnDiOPj8RMbyBeubFodDpaNHPNXFYrsrZ5J3RbBssn0zAsthePOQ75qkbF9Ep1aa-UJqWvSOwgC886JxtE-6AM4OYAJ9zQF7HgPAFGqiBytLggxm1a9c9W1wIt2UAGd9krQegynRQIVrX4dvGBHJvEIE3LbZ6rzmpJooFOD0LBaAcMxrsr8I8kgvWsxfQjoLPvd6jq7T2nfi6qJODDJEhWgvcjE3lTfUH4Dc1peh0d_CQJbuLqfx8l6lZmk0bEqMnE8si532Q2BTxhWWd8pGmyrAZ-h7z1dJeYmh005KkcNvShNcBTorNrXeUiVUPEz28H0PeVq6GY4S00YGeBjGK7jEAoq-lEedr9mKyjeRTSuEqqGcqoVI5owQdZlS2wrMqQ3xi3bE8skVSxHBqG6QvZqotLkjHf6NR-f4p7JPLkKlSgtxyppX0NYFhPhQf8jdovFvbcfqTeoc7Tt2HzN9YZ3xYXkss1GJZSpI_7DvHyNR615zpq5CXdecPp5_juj4Wt00_fZKBEtLFGwB_UfwtsubliLLXwXcz0PxId73g5JhbY8g3BfrL2zUbsgIo94AT6ckb8gR3b2IwAHziukheOcFTFMGy4MB8yJDQ3vtwSotBQPAMODKettU3DBYc2B_8eC_QsOz3xArfnHtyAOf02e3wicdHCNkmoRyfCaJlcKnVhdyzZQESy3iHeG4nU45WSRMo4PnfMwqQPP6g-3Am2JPysE5j5Jlox0uMlH3FTwmVukrlWOO3UkScoOFeiO13O7_zGIAi9VRXnjAJwKNLKnbEuBjv4p9nzqE2x-xkk24_rTjT9YCRT2LHZYMH2Exz_-4fguMQV69eFECHAhOsELbqaAvOdmOc2IvqHazXO1yQkTZhpjcF2HCCfSwdxwSa059CHrmSyzd2DKEPnbQ7F_No9aM9qtguSkUzfCWXu-L-LllHxVANM5HeNm7acTjsZi34HnXPZKQdb4KoubFkHUWewXL8CXOiz8aO9WVXoV3k-pSP7mqUjUoJApI7yaHweLTFeb0MX1Sfz_N2J_8c_SLxSqwYfJA-DD3-5ilOwzZitSys4EayCPoLJuuIM9GOVaCqyPLo_V-MkKnX67xWpeRwLCV10QV_pfEiAVYGB4_DQoMWTGM5-dV7dvnihE0tcsR77zCeuLMgeYI-gyoUf5uvlIRY5B1b512mkeiFDUXYuxqTFrHfvxj2_AkK0pbi87sN_u639aH_Fjq5yuLTBSnwY1zCeB_H3XQ7am2zlfJ4nnnXwuIgsYknE1rtmdxNSxJIazQjUN4zjTnw3N4h8oQaM_XboZdJbyW9QceC_5lm_-BAB3KLRyDyBNxmbNgwww5EKKircOiccvnM-8pa1YF-nMNO0_wTfWJRqgaJ412RoMonXCxJJS4pwKgs8L9vqVhMJI8QWqmUhOaUSlU-5KkOJuTO4htea5tK-RI1esxZUwDvSA_zGz1wgVLU99tRkdsAwq1MKIdRj1y2Ura6ZtfCuClA9KYbZwhjGnyGHdssOS0VPxIaW6-FHjDx_COeoePvOsGKn2Osn_aBuLLBqHVigPzDNjXc5Zjn8pye7AH3tz5ZsbVNgWJtdw1OrQqPrWkawaD6WqApELrBPnACbOImRr9SKbEGCMjEqcWfOpHrypW2LYIvknQ_v-BWEh7RaoORWzdu47LywYm1EGnuxeRYP3Z5jnGJ0bK0ermEHmdPfkmlmqIiVgQykihDdLcs55-k7OVpMUzONB7phVfk_uZcXGYkfLeCe8hgG92NFgg1aEyibo56Ods_gBWo69AE5ELnkI_VYPLM4E4VVGFtLRnMefVk07aDwhiR2UyYEc7_PIpWNMQ5vYocjjtuwfW-KvhK3mgTQNeTOUDTCKrJcReZDt558PfMxnd5sdrUb5yE3_dIYama512xAts4cACMPP8FmNBalNGMOjcJvOBS393w5lTbABIma4InOjCVH-Y-pjPZLqO9qiJSzLnnVEtRFtLqNhGyRv2_GRsttI99RmT7vYHRwhz9n7Aow1QtA8AUNTsEvpMZH3o7oha4PD7WeRYzV7KuFZdxlkun3e0B0kvVP5cCZLcEmjMiLG4dvRfH01zVtCpDhJe_qO11eELg3Bn1QHp_wdY2jpUryEGE4SYQzwKRKcK1VHwkJVE0gdE8QaDLM6ptU9GqSHieZ6i9WHPYzkjhgqmXSLjq3VULgcKqJnD8EU2_JIBKaa6Qr0TRPT7qXSUbz9aBuuym-9zdUoL7zejycgB_Rnq7E5in_1vaLrN6BQ6FFEhLPqR_NRhKsQoRDN15Tg8DO5cxt0jv4nDIs36vhAR40NOtgMLPXslOuCHfQVnOSgK6bBylegImR3K3HWZ3OF_2LcGcDe5NW2uz0IusiK-YqTz3l8NzQEw83u1nQ3xODzPbnV9vp5yfVefJDomseIZzBsWQ6aEf-mj2rXt-l24tCBaegQ9eh_l_D3s-c2otUbXOt9UgqAj8q24ML2xw4g5e8MFoRbUTBIxYLMnfbguQvPVQQ5h35VCwWKPCyQnC0Ts5Sq7G_yjRhJgG4KhPRrTg7b1TucoDp4lOGz4ijtuEDNZG-8Q7byWhzjZlFTIJ18_NQRs8F1VYKgPGdyLmts86w\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIIIIjANBgkqhkiG9w0BAQEFAAOCCA8AMIIICgKCCAEA7PHss+wR/pW6VaB9dS1x\nYVnxwmq51C0NEldT6PJJBEHXzRj9y05o0NXishKcsfQQe8cOmsNlnIU0Ls4eEGck\nyFeS06gXL6Q5PCr0mMqY6XmkIarnRc7ZxDUIDMU4VFAuRH3UCwXAwKVvs0ZtbvQ9\nLiW4fqVboGi2qeyowIEc7b7/x3+8vQhLEPOoKEwMyiIT7MjAzFysh8iqzMVng9k2\nfFYe3zgxDlkZpSCXQNH5U+xH1udtdlOjvEENCi3fmgXZL2P5yWUyACb27V0w/GgP\nC2s+gbS/+YRzqwQUTjg2p1afb0R5DT9xHFJwneG8HJJdK4RchRV+f7wGkf3KESM+\n5nBOaQlFmQNZheJCagCLcyD65uy8C+rjTPTLdJR89SbapL39ScNhVAD7I5gfCOVQ\nGrU4k0MCUkAiReiKjRJ947WyhI2ULgpEjRtIFanZ9I7pABmU/vI5Nplgqu35QZCd\nSpRBSHUBHbsRIQrTdpQKCuc+M8zJIR6Rp27huV2BGmSwoVZ730u9IrZL7iHGniEr\nR33cXrgF5mRiZ7LAW6TA0FpVhIZZHuvTPpKXFxNtPgnkte30/5Q9wg3+glSJ1nlT\nSJo86hpACUZOZksBKlH785gZAmtwuKFYpwMD0/8RsRbwaEi9d4SrY7wVZSzmdSyU\nBOcZPjiP0d4VXmvgpl01k1c8xEXo68RZCUbvx1UlrC5sjDAFttHS9wg/G3jGHK0o\nt+3YTBynEdUTmed5ymYbFiPZMFQK8Fs2FfbqNN8sxUNvlGuWo+Lpkr2YaJli1tbm\nn58JUNPF7QtaeAYRthe2x4dyKi58MM7xFnyvEJWNCwocQ6IZU6oV2oBof+BBvmAc\n1dyKROV437EMLL3CNfWYNpBcT5iJfpmOvkNrRZloO98qGIF2WSnYYodEoHkw/RIr\n0gm3P3nPqbSccjpTu0Umpiex5F9bFD4SVwnd2r7IaxFQ5NvgyIqyjc1/1IMqgDbw\n5ggKvJgMsnMbiNwUWlS/kHzlY7nT8fQWBruy6vKBptb73iqKIJ3/J4lG/pxdOoVN\nmh70QLXA0XSTY8NgjiNFnhZT4S9fjwybyp5Mbmu2zonCnzsefGRlNF0WkOuPKczD\ngJB2K/oK5bGeZunwsM9EAPE86qHcgAR4lTp/c/Bm9K8336zlh1bymYtKfwEufJxv\n3ztXG6lrQWtlu062cVeVt7f88M7JM9wmqvSChGvyCmgpzNQ0WPlYG7DPEcq3KyNj\np+/hvlVYfg7SKxYYHmd4Rs4ycBtadhhuPkcetbuYpCJ8g4DZ+QGz8Pka46kVgHwK\nH+juEeaQ1hHzEbuunY+amopRMmdYT88qJep3IyzJ5FGMOfESJ/Iaj9gnYZg3mHcS\nNod3gG0UtjdvmibgwsA/E4N98XwbMXIRknGmCG2MsT+i8tISb73bOw+2jALmTKoB\n8J/HIV5lAzlT91eTo36ZXbAioEoMB3Oq+ZLRp4usUbMISOFBzUYHrqecclMPLG/e\nDaa2lVSs7UEyLFfLCB2o7/5xktq2QmdVhjNIqX01KM7akBYjjY8zSr5BqpeI5Xq7\n7agfUat9yfCpXpxlzbXphcITh1L/oU+X/XdZPLWmbXe45/z8zp9to0il+q8+LK6q\nS6PXin4zc6ep4s2cKBvO692bBiZIYurm9dMXPSl+u+lKOOAAlWDWAmnDj31uJysS\nbiyIqs7yTKpj+9pkB3yuvNn2RSAx9iOhIwdjTsVXMwsfHrufUtIYj5WwkyM/G92F\nKjIU04gH4mbfL3//kSiAGUDO7ZCmrjS6xC8XkY40B//RUlp6KCumhfzn1A7Mt0SJ\nDDZIZOKWM3Q5BNhJe2Q/tBaOiOBKUiCI1zkRP2SGWCgDjIVLbesucVvWuH3Sbb5V\nKrTuXdB6u1KxtxI+KtIv1Tkt06MbRk1myzP/uf6ryx0UQlJugdAal6+2WoHhQyi4\nN0T2pRmsbhcb7eOJ36mQW9LsRX5nbXzLk32RZXEjupKWAGAOn2LwoU+m155XUMCK\nSUNda6F3V1pSpzkW7SSVUue/laU3wQWw/bQab6LIJAnXL8KshlIhs7umdwS8RZC4\nYF8VI8hBlygMsmI55w/wExOZ93dy12Ah8hpgqpRv3a4ej60yfl7EseFE43FH3Ru7\nrT/kAN4ZUgaeK/FVbZr0yDUd283O8qNI7pVXdT6MA1Oe7cTik5ehZLJsBoy01Up4\n5XdlobV0sdiND7B3NRhI79BF3ls4tQStdOeXAkels3jkkSQ0LuHrio28NGGjccUI\nDwTswOzh6pjyIpnevdDIFivq0jIPEcvSKbdYcxOm8rkxTq1sof7TRr/8Tk7CEMTk\niyS55TcgJrQBPRQ54BUmRbB+qIsaEv9KTemiwlP1TJqHjFFZvCSsPk7vqzrCegoO\nvN0+XMKqK7nZan6wEbQjP9jLzgeJ+M5hOBpuwJVojXC/PjbG/m1k7Y78OXrXuZ8m\nOzUKsx6ZwHFJscvX/sURCJ7+56wRLBLWEmNAV4612oYvRERggzNpU6AIKdE9ci9k\n0R/6aWSswQDM9e57qvizoO+pBPDjGgVHQ0Gdt2D8YBOKV7IoZZr39CZ2M99zD9UN\n/x7CGOCjt3WYuv7z5J+0FEDApIMTNOAmxKVSDLHNB3a0CxDYD3YRnCyymuB2Qshp\n/bLoKtEeEVOZY2vytmYpBsOt7vJIdZ1dOGdsOfO4c5GtvV+DfrvxxaEfn89a1A8E\nCI/F/F88ff8HZXSmKnav6zMCAwEAAQ==\n-----END PUBLIC KEY-----\n"
	alg := jwa.RS256

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtRS384(b *testing.B) {
	bearer := "eyJhbGciOiJSUzM4NCIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.XhsxQEnWwzoduyU6KMx2ers0SlzW2v5rnDw-cmI8ZLgrhIIAZFY1T1C09lls002rP5_1gSsJcshEnc0e5RIUXYDzmejyZ5Ll8A3Z3aQmtEhVXjgeQDm9Ke6l2Ou2b2x1UUlUIz6zZXAolzlgRk02jbpsbRXjtKEZVZpq7mZA5u9-Dl9pPR3UR8ndW6NxSKuP1PbeDnW-Ky6qEgH_FL_laMbMNYQNoHTR_zVoEZyz6pdlTVDj4s2REN8Yq6ntR9QqE2kh-9Mjo1MrrBCh-9T1LB2QzPzQ7YZaHhCurZ3twEz9x9IRsIMGCDOVnd8_fNo1qlq_aGx7l6-GQaw_QURyZKlVMo91fY9hdbkhT_wCmnrSUqdFTVggPdqYl1uSE2mH3aBk5ht_eSKzrvh3D6w6tWWVLOr0aPhIVN9KXXZSlrqCijvhxCXyXpZZxcae1kJmIcRkbqoi2HkNAiDqgp-YfSQOfgm9Mrm7x7HnUNJerHzEAkpFym2xDdzxil9tzYlBVOPOHHauIWtKraqP97Tf1HZQdbl0jQS-zYl9M8SVl6fDP4DJ4iYCJSEu8BRUiP0hA9T7coFdPAgrCzHoAiKtm0fbjY1iKZvA9UgRg6Ngq84BcG3KXsLUJZpBrFd1bb8Ya3cv7xQNDpwinjXn_aHYIFAjOLenv3fnn-YlExYCsB8\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n"
	alg := jwa.RS384

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtRS512(b *testing.B) {
	bearer := "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.ZBWwH08G8XgKZSirRCn_y66VCubpWmvPd9LmyPu126Cp0VH-ujqgNsih8jRqdE_iK_XpzPIiKteFfzfpZgemtQvM1w7yGGg_yDlgYrAeyosIgTXL_6aj5xsiS0ybE1Da5PhEp8jxIP9vjgQmEoJPYR_8B4enGjLap5spkag4ixQfvOosighATAmtTegEzv5Rqjk7DrswfGyu2AwiXkAXBSSeNckvTwXslNkyAxE82uOfB8iASqfarEdc_d61tKIMhiDYsWZhjdJXeczCKw6-0aG-lqnPe9lWAYj_R5pHC7rxXL582exyUEQoHeriwNog1P2LBUmnSQEia1sKv1XK9F33QbIijCCCSkwv01V0HKfTQA4mQ8wD2WZwjUj6seUH0ezdSC8Z7fJeQ7GbVM-VoKLMSdrSjf988ICwbNy2XBuBUHjpDluTV0BphJnWLvL-X4CBi0UV0HJMqBE3zLr6VEuPsc3N7NjGW-FTbh-bzlfCNUCb7fln_xMGSILWRKOpqj_xjIQrKbi3Kus41vFP2ChX-GHAJiJvuZVfOjIytpE19npnUC_CZzAhD9w5s-ToIR95FpA3wE1HUgle4p7_cbYMwDvvzi1Y7hf0h0bouUS3FSmycBTx7dUgKJeNwKtetNFHloM9CUZfdK9QGY8bSXh4r5h4gOsOeM47ic-W8Uk\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n"
	alg := jwa.RS512

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtPS256(b *testing.B) {
	bearer := "eyJhbGciOiJQUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.TbloDjGto2e90mE6yKFs-Ox9n2NCdtpm490c-NHVXCvWZrtbICO02q74tiHxIvwKMSpULlfR5YhjcO2GN4tvcimWJAbaBamG5ydfEry3sGCCmuNkPhuCTQprGFkP7ZCLJLsXwbrWCYJtkYXg23rX3MlxAP1TcebxU6x3nRRpGbjGpOUvKBgoLH5Wc0aLqShZZ12m5ck4ipHqrjLeBiqbcQD-5gQQH2ohSoDrXFgRQVDS5lJFILlSFHQYY3z_9Yxz1Y7GYVHHrY6HgNJ1INhf1SurW_8w8gzWAhkfJg0zrR6T5AtM9Z85PJSEuK--_FW4zJkvbgB7pyYD3r4Q9qQp1cpq12vQq-nTxIhHLp-zNwjdrKH03NWfnvx4_e5LrZKEKAvEpxkXEXDANrOHwFAAGcZ4AsnjOLI3UcGqXozs2OHAmXfk6byxzjXugy1j3DkgzkB2-nr0cSXqkWOd66mFVw8ZN1NdU6p6ioD05glKHZdp653YC6ySv5UO0KOGwHirP6iIzIqQXNiSzQNLq1Q0XwlYUIHYnM01h3po3zDSgGsEP1UoyhbWdBbpyw0FN26SJvAdqkWmV9okK92e62pgP49P1eD9NVOoxQx7_fqe-CAjtLBOoYJ1B-Rtbm78mFiDvSzH-uDxwcLjVrnzylwoKGfJgq3k9EN-HkxJ_IOnmjQ\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n"
	alg := jwa.PS256

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtPS384(b *testing.B) {
	bearer := "eyJhbGciOiJQUzM4NCIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.sSwCUL4suyCEGFiRt-OY87eYXvFtJOnQwZAoHcSfI1wH8EW09Evbnrduwx4-DpMlOGbw5VQsdArcjPeojJ3IeqnI-SP9TNjPOciWQlB0MbAD32ANuyq3dQt7Z5xsKy-jObESB-c0iDb1CdwdCamLzBoFajo06hxxLDgvFsV0c8er77u85Wrvlh5tr_e6Wqmg14xKADzaZU7QDBBsh9L6XcPrhCif7NzKuxJXVWI5MwNBNgU6bgeuTJiz6OZJmNqMNOe1KeR9k--xLpsCmAf4vW2o7lPysY84Q4dXGLCOX_Ox_Mmtqyu6qJZXI7KlLMDYbyFIEJqkw8tEAhrwbeIlG2k5yXg7TA3t2E-YmKC2hapz4qvs9eFkWtNoQIcBHPXh7BqIE905tbLEdosddTmZkHDTd504puDZQOMm9_oIRWLhE890BT-7rL_jte93uJShJeRsHP4UrahdBL0Dh2KF0aIH5GkgBadkpdgbC88v5PMQi_Iw79wixBmIw2qGMG3HWpUksp597eyg8G0jGBxnY1hzd08FSkxCx82CUQUkNkLf7zVJsWhup2b0YkSFBpUnSziDE1gbS6zTI3rOPqlGaZYAgh1kYiOkC2rhP769fyTI_T62vP8uM6BVr4ehAkZ9aDbD-ojINym-a7vCgcLfyk1Oo6xO_gzy31CPuzoij9I"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n"
	alg := jwa.PS384

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtPS512(b *testing.B) {
	bearer := "eyJhbGciOiJQUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiR29sYW5nIFVzZXJzIl0sImlhdCI6MjMzNDMxMjAwLCJzdWIiOiJodHRwczovL2dpdGh1Yi5jb20vbGVzdHJyYXQtZ28vand4L2p3dCIsInByaXZhdGVDbGFpbUtleSI6IkhlbGxvLCBXb3JsZCEifQ.XLUNSwV3wsLfxlkNQo48dBtSk3NEsJQNppd3EqnqMZ2WL6OWY22bIR2FXfsp-PIY85pAhJlfdI9qMaYrHSrcwnKWl8gY11tMIOh9mMfRc-hUcurnczPAcdn6eIoxvS4bLTfKuQy5VXb5O1hax8xtaH4Hvj9BeinNhkPG4JDDCVKcZa_Cn-Zcd-WFv-0Sht_dv2560F0zRwHCE7h1JYQzAQj0ZhfFVc9PomUJjHbrgtFvQ0EuRsqVhtQvl1reKydZuytiNR5K5lH-fBVG06IWsIst-XKf11uBpjIRgtrGpU2h8fZopDDEGPKJdISGIVUrjJfTKZShm8D6-REx3AR7M5htytpKUf8l1tW4NRBzJeG15x1tPw-0sJwh1IF63IB4bJzDHFM70KkJc5x6yK5tjTY_HREYx8oTy7z42MmviQaMTMwpq1wIGgviXPRnCDtkCiurKgMgRwPDvRabKfvKpzCKekBzF5cFt1ttgM8nkNfgpQUkv40XxkxPaQWZr3GBWBhojncYG9aw91qZUhNeskTMCl-LxVhWWnqtdA9w4hRuBUj20I-ZrAcSohvRBC2NAMa3IvpWwuKz9p38-Z9xZ7VqrEgNJjScyZet-t6w-cz_PJvaJ7f8RxzaRGaWdlfKY3hQEI2qQJbfouzw8GAYPTV62U6_GVKk5AHTigImAXM"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n"
	alg := jwa.PS512

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtHS256(b *testing.B) {
	DoBenchForAlgBearerAndSecret(b, jwa.HS256, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.sHm-91nvgMEqZelFK6sQhWBGtshLy-A1FYFkZQ62sE4", "your-real-secret")
}

func BenchmarkJwtHS384(b *testing.B) {
	DoBenchForAlgBearerAndSecret(b, jwa.HS384, "eyJhbGciOiJIUzM4NCIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.YZ3u3q6lqhr2aZJb0sHIiv86Sc-osMy4_IxMVMh8VOJtWPxr37nw9x4mL7RMn1FE\n", "your-real-secret")
}

func BenchmarkJwtHS512(b *testing.B) {
	DoBenchForAlgBearerAndSecret(b, jwa.HS512, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.S14tgmkZES1f6l8G3qXOPnn9DEA60o45tZyw6l3QHCdwvc5PnSx3iOnJWZb3pXqyS3WOJK55BgPm43KVreqhAg\n", "your-real-secret")
}

func BenchmarkJwtES256(b *testing.B) {
	bearer := "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.PVhXCdrAglq2TuRBuF6QFjTV2Lo33F44NQS0kwFLr0d9IjlGIknEiRSZ0UCQxktVX_bUcEA3f5vhWxAUmAAZcg\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n"
	alg := jwa.ES256

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtES384(b *testing.B) {
	bearer := "eyJhbGciOiJFUzM4NCIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.TTt9MBH2XKNL-1IxoiLmFQyR3O-t-kPkUQbCFxqFQfDDnZDWDeG06Ea-s1_syBrlsmd0gJq3g1PvVS6KVswOtqpF28R2s5LVwAYKbmsziMRGc5jHAPjABzT3Aigwh3j9\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEbzpBvrm4YDC8/AWuFNylOeApmVi0NpiR\nI4tQOE0y5bm7zjwjr/nvwKreCF0mT0NivYOzQfu6D5Wy1Jpgm0ebMDljSKEGkZwC\nP2/jogu9YzmaAnV/Re16ffln/n+tsATG\n-----END PUBLIC KEY-----\n"
	alg := jwa.ES384

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

func BenchmarkJwtES512(b *testing.B) {
	bearer := "eyJhbGciOiJFUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.AYEE2U3R1D40NxA7NcRApQ6jHExwLijMOxW527gN1usYLIu_Qg-X9vqm5XrDXFEyn4VpeQ83nE_az8QqRBZPkq5gAEIeJoyO3Mfj1G9uFyAyQmy2W8YYyqUUyjZ5lFu03gMmR_vdTQfWfFyAw6NbD0k78KF4h2qfXeybN-Ljjc7e7K7y\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIGbMBAGByqGSM49AgEGBSuBBAAjA4GGAAQAuObw6RyorEJfB+UUgkJcmfdIlLuK\nwO6KFjHF8u5+OPpHNZuq203djMAp07caPZsuEg2M0WIjdm10LefJhnepq/8B1AZT\nI/Kr/XahLw9Wn9/TDWXgLwaPnPk6jnzlkVFpts4mcvS2/ec/yBoPViHZ98sMwOoS\njTrSpJkw0BIo5qQSuGk=\n-----END PUBLIC KEY-----\n"
	alg := jwa.ES512

	DoBenchForAlgBearerAndKey(b, alg, bearer, pemb)
}

//12μs
func BenchmarkSimpleParse(b *testing.B) {
	bearer := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ims0MiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiOGUzMzM4NzAtMGVmNC00NTNlLWFkYzktNjk4NTc5ODNmMDM0IiwiaWF0IjoxNjA3NTEzNjU4fQ.MPDlM3h1A9a_Yy0ENRAh-E2rTJdVnb0AVQx7eTksm0kdq1CtIW7OAQQghNWwYhc-X1f1t36EG4j6fEViAvCvB-0M89ZJAPrNqLeyVrtU-O9nmaEgSCf5eQFJcEea41L42wbL4T4-HtwUGhuEnaH0UrR5hc9BSXKQwiXVIi0mtWcJKK6tCT4CNlCYH73igPmBXIxJeRl3ZLCqxDEZ5hdC1HEgOls_HGRSUfNKlXtc3-U4kz32HlsmDwwPO2-7xuj3vfeWCv6UgIwkSaiCkcBWJ-MsWzOX_VDflrgQwWe7yV2pnvHmJxido_NwuSd6JJ55da6xY6yBAAbbVo6pRIK6og"
	DoBenchForBearerTokenParse(b, bearer)
}

//10μs
func BenchmarkJwtEs512Parse(b *testing.B) {
	bearer := "eyJhbGciOiJFUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.AYEE2U3R1D40NxA7NcRApQ6jHExwLijMOxW527gN1usYLIu_Qg-X9vqm5XrDXFEyn4VpeQ83nE_az8QqRBZPkq5gAEIeJoyO3Mfj1G9uFyAyQmy2W8YYyqUUyjZ5lFu03gMmR_vdTQfWfFyAw6NbD0k78KF4h2qfXeybN-Ljjc7e7K7y\n"
	DoBenchForBearerTokenParse(b, bearer)
}

func DoBenchForBearerTokenParse(b *testing.B, token string) {
	for i := 0; i < b.N; i++ {
		_, err := jwt.Parse(bytes.NewReader([]byte(token)))
		if err != nil {
			b.Errorf("token parsing failed, cause: %v", err)
		}
	}
}

func DoBenchForAlgBearerAndKey(b *testing.B, alg jwa.SignatureAlgorithm, bearer string, pemb string) {
	pem, _ := pem.Decode([]byte(pemb))
	pub, _ := x509.ParsePKIXPublicKey(pem.Bytes)

	for i := 0; i < b.N; i++ {
		_, err := jwt.ParseVerify(bytes.NewReader([]byte(bearer)), alg, pub)
		if err != nil {
			b.Errorf("key verification failed, cause: %v", err)
		}
	}
}

func DoBenchForAlgBearerAndSecret(b *testing.B, alg jwa.SignatureAlgorithm, bearer string, secret string) {
	for i := 0; i < b.N; i++ {
		_, err := jwt.ParseVerify(bytes.NewReader([]byte(bearer)), alg, []byte(secret))
		if err != nil {
			b.Errorf("key verification failed, cause: %v", err)
		}
	}
}
