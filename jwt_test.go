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

func jwtValErr(t *testing.T, jwt Jwt, want error) {
	got := jwt.validate()
	if want != nil {
		if got == nil || got.Error() != want.Error() {
			t.Errorf("%s key validation error, want %v, got %v", jwt.Alg, want, got)
		}
	} else {
		if got != want {
			t.Errorf("%s key validation error, want %v, got %v", jwt.Alg, want, got)
		}
	}
}

func jwtPass(t *testing.T, alg string, key string) {
	jwtValErr(t, Jwt{
		Alg: alg,
		Key: key,
	}, nil)
}

func jwtBadAlg(t *testing.T, alg string, key string) {
	jwt := Jwt{
		Alg: alg,
		Key: key,
	}
	var want = errors.New(keyTypeInvalid)
	jwtValErr(t, jwt, want)
}

func jwtBadKeyPreamble(t *testing.T, alg string, key string) {
	jwt := Jwt{
		Alg: alg,
		Key: key,
	}
	var want = errors.New(fmt.Sprintf(pemOverflow, jwt.Name))
	jwtValErr(t, jwt, want)
}

func jwtBadKeyAsn1(t *testing.T, alg string, key string) {
	jwt := Jwt{
		Alg: alg,
		Key: key,
	}
	var want = errors.New(fmt.Sprintf(pemAsn1Bad, jwt.Name))
	jwtValErr(t, jwt, want)
}

func jwtBadKeySize(t *testing.T, alg string, key string, badKeySize int) {
	jwt := Jwt{
		Alg: alg,
		Key: key,
	}
	var want = errors.New(fmt.Sprintf(ecdsaKeySizeBad, jwt.Name, jwt.Alg, badKeySize))
	jwtValErr(t, jwt, want)
}

func TestJwtNonePass(t *testing.T) {
	jwtPass(t, "none", "")
}

func TestJwtNoneFailWithKeyProvided(t *testing.T) {
	jwt := Jwt{
		Alg: "none",
		Key: "keydata",
	}
	jwtValErr(t, jwt, errors.New("none type signature does not allow key data, check your configuration"))
}

func TestJwtRS256Pass(t *testing.T) {
	jwtPass(t, "RS256", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
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

func TestJwtHS256BadAlg(t *testing.T) {
	jwtBadAlg(t, "HS257", "key")
}

func TestJwtHS256NoKey(t *testing.T) {
	jwt := Jwt{
		Alg:         "HS256",
		Key:         "",
	}
	jwtValErr(t, jwt, errors.New("jwt secret not found, check your configuration"))
}

func TestJwtHS384Pass(t *testing.T) {
	jwtPass(t, "HS384", "key")
}

func TestJwtHS384BadAlg(t *testing.T) {
	jwtBadAlg(t, "HS385", "key")
}

func TestJwtHS384NoKey(t *testing.T) {
	jwt := Jwt{
		Alg:         "HS384",
		Key:         "",
	}
	jwtValErr(t, jwt, errors.New("jwt secret not found, check your configuration"))
}

func TestJwtHS512Pass(t *testing.T) {
	jwtPass(t, "HS512", "key")
}

func TestJwtHS512BadAlg(t *testing.T) {
	jwtBadAlg(t, "HS513", "key")
}

func TestJwtHS512NoKey(t *testing.T) {
	jwt := Jwt{
		Alg:         "HS512",
		Key:         "",
	}
	jwtValErr(t, jwt, errors.New("jwt secret not found, check your configuration"))
}

func BenchmarkJwtRS256(b *testing.B) {
	bearer := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.aHXyvYSVGeNAB4wig6Nv5--vo4ZWG6CSLlrIFA8Dm_ohaVLiT_2y_cJc_v0kc_bACgdsDJloooAllh75tc870MiqGXJ4enxoUUmIKory4PUooFlFENUiDMS4EQz7JXxxhbjlfTURo00kM9hSSJ8-yluGXaX0Xirc15H2oMBDUq-1nUCjECsI6T7P4rSYxrZOviK58PjKuyocvHm_3Q0nAY0NHJ9RVYatntJFai75EbLDzzOvTDUXFjVMIQ0-nR-55pKvnykoZoqdaGJInsHHl8-M_fwxeYzvBXd_saBfMI2-KeYLUXTN9_n2z6oey4wapalJavWLxSnH75uUhOlLkU0NQWeylAiEOKXoNLBJwf5QUnwMHgrEvzZSz7-2rYG6_yuAHQ0csHRMVCIG0YKKhvgjis6tXmb4vjiR-DdJ1D2zDXVgjeQCqYdS9ycbx_CgeZ6ignalz85Xu3fm2_UwsdrrSEg9sYzzhzHl9ksD8mxlfUwcywb_HVkz77pciIcx0D8DUFR8JuWxy2AIMSWoYlWag5hVkxhIDSlcNKOuxOitPwyLA2bejxuDhZ71Kh9i_PmTzVYCrnNoNY9CKjWmC2FS_QVoDpCKcdj3T6bclucidGbtdM50fYA47jnJULcqJuaRav7otR71RmxufnrBGI6yxxmEbQPeBBFt_Xj4bRg\n"
	pemb := "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n"
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

func DoBenchForAlgBearerAndKey(b *testing.B, alg jwa.SignatureAlgorithm, bearer string, pemb string) {
	pem, _ := pem.Decode([]byte(pemb))
	pub, _ := x509.ParsePKIXPublicKey(pem.Bytes)

	for i := 0; i < b.N; i++ {
		_, err :=jwt.ParseVerify(bytes.NewReader([]byte(bearer)), alg, pub)
		if err!=nil {
			b.Errorf("key verification failed, cause: %v", err)
		}
	}
}


