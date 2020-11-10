package j8a

import (
	"errors"
	"fmt"
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

func TestJwtES384Pass(t *testing.T) {
	jwtPass(t, "ES384", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
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

func TestJwtES512Pass(t *testing.T) {
	jwtPass(t, "ES512", "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQQwrEUWQle75mz/T+wt4fXoAn39M\ngmCdZ3ZY3fXIkqQDOOq2JJ3x6Rayy1GTp7nlN88JqxX8UC71LdmbIyKi6g==\n-----END PUBLIC KEY-----\n")
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


