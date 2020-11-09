package j8a

import (
	"errors"
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

func TestJwtRS256Pass(t *testing.T) {
	jwtPass(t, "RS256", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS256BadAlg(t *testing.T) {
	jwtBadAlg(t, "RS257", "-----BEGIN PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtRS256BadKeyPreamble(t *testing.T) {
	jwtBadKeyPreamble(t,"RS256", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
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
	jwtBadKeyPreamble(t,"RS384", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
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
	jwtBadKeyPreamble(t,"RS512", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
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
	jwtBadKeyPreamble(t,"PS256", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
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
	jwtBadKeyPreamble(t,"PS384", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
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
	jwtBadKeyPreamble(t,"PS512", "-----NOT PUBLIC KEY-----\nMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
}

func TestJwtPS512BadAsn1(t *testing.T) {
	jwtBadKeyAsn1(t, "PS512", "-----BEGIN PUBLIC KEY-----\nCCCCCCQEFAAOCAg81111111111111111111AMIICCgKCAgEAtXhyIjACJ9I/1RLe6ewu\nBIzZ1275BUssbeUdE87qSNpkJHsn6lNKPUQVix/Hk8MDME6Et1zmyK7a2XoTovME\nLgaHFSpH3i+Eqdl1jG9c0/vkHlwC6Ba+MLxvSCn6HVrcSMMGpOdVHUU4cuqDRpVO\n4owby8e1ZSS1hdhaqs5t464BID7e907oe7hE8deqD9MXmGEimcXXEJTF84wH2xcB\nqUO35dcc5SBJfPAibZ6U2AaNIEZJouUYMJOqwVttTBvKYwhuEwcxsPrYfkufbmGb\n9dnTfKMJamujAwFf+YUwifYfpY763cQ4Ex7eHWVp4LlBB9zYYBBGp2ueLuhJSMWh\nk0yP4KBk8ZDcIgLZKsTzYDdnvbecii7qAxRYMaSEkdjSj2JTmV/GtDBLmkejVNqo\n9s/BvgEIDiPipTWesPKsaNigyhs6p6POJvOHkAAc3+88cfShLuDpobWmNEO6eOAG\nGvACbWs+EOepMrvWuL53QWgJzJaKsxgGejQ1jVCIRZeaVsWiPrJFSUk87lWwxGpR\ncSdvOATlGgjz28jL/CqtuAySGTb4S0LsBFgdpykrGChjbajxeMMjnV3khI4c/KXl\nSmOsxHfJ5vzfbicw1Inn/4RoVxw72p4t1NN3va1W6jZt/FZ5R8xgV5T5zgeAEkSm\nHJa/PXCQoBYwK7cuMJhjRaMCAwEAAQ==\n-----END PUBLIC KEY-----\n")
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
	var want = errors.New("unable to determine key type, not one of: [RS256, RS384, RS512, PS256, PS384, PS512, HS256, HS384, HS512, ES256, ES384, ES512, none]")
	jwtValErr(t, jwt, want)
}

func jwtBadKeyPreamble(t *testing.T, alg string, key string) {
	jwt := Jwt{
		Alg: alg,
		Key: key,
	}
	var want = errors.New("jwt key [] only type PUBLIC KEY allowed but found more data, check your PEM block")
	jwtValErr(t, jwt, want)
}

func jwtBadKeyAsn1(t *testing.T, alg string, key string) {
	jwt := Jwt{
		Alg: alg,
		Key: key,
	}
	var want = errors.New("jwt key [] is not of type RSA PUBLIC KEY, check your PEM Block")
	jwtValErr(t, jwt, want)
}
