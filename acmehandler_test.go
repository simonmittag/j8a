package j8a

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCertCacheAndReloadFromCacheSucceedsForValidCert(t *testing.T) {
	certPem := "-----BEGIN CERTIFICATE-----\nMIIFtDCCA5ygAwIBAgICEAIwDQYJKoZIhvcNAQELBQAwdjELMAkGA1UEBhMCQVUx\nDDAKBgNVBAgMA05TVzENMAsGA1UECgwEbXljYTEaMBgGA1UECwwRbXljYSBpbnRl\ncm1lZGlhdGUxLjAsBgNVBAMMJW15IGNlcnRpZmljYXRlIGF1dGhvcml0eSBpbnRl\ncm1lZGlhdGUwIBcNMjAwOTE5MDQxNjAwWhgPMjEwMjExMDkwNDE2MDBaMGIxCzAJ\nBgNVBAYTAkFVMQwwCgYDVQQIDANOU1cxDzANBgNVBAcMBlN5ZG5leTEOMAwGA1UE\nCgwFY2VydDIxETAPBgNVBAsMCGNlcnQyIG91MREwDwYDVQQDDAhjZXJ0MiBjbjCC\nASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAN3FFHDc3fWIyxukMDRriEbY\ntVA41EeiQiwf7RLdDxh+N2VAazUbbxUJ06nKAslX2+6ZmJrMlS+ionX1BvPhPy3s\nnuZI1movXcvH6ZV5yUGZyJDocjOTHHqNwPSDOAQX87tLjQbCa8Rw//B488GoPbaZ\nlWYDvZQ0Mw5rasiu0B+OI6PL8+Vnc2jXdPlc3tiNoIVXRZ14TNei7bUDA3O1y593\nift2tQ/TZxlY7fylZWhTV4sUm/9yk/zob+dyzro795Jy8vThlePAN//tZGLWFzG7\na8o9Mx36BPncSZ0v+EfEvP24ZffIDFRtysBewu2+33IVpISlbaHgj6nsuv8GFM0C\nAwEAAaOCAVwwggFYMAkGA1UdEwQCMAAwEQYJYIZIAYb4QgEBBAQDAgZAMDMGCWCG\nSAGG+EIBDQQmFiRPcGVuU1NMIEdlbmVyYXRlZCBTZXJ2ZXIgQ2VydGlmaWNhdGUw\nHQYDVR0OBBYEFEEFnOmrROOjNNQrLRoXPXsJLPkeMIGdBgNVHSMEgZUwgZKAFOlV\np+B1WShwNQuSAVuQAvCe/teZoXakdDByMQswCQYDVQQGEwJBVTEMMAoGA1UECAwD\nTlNXMQ8wDQYDVQQHDAZTeWRuZXkxDTALBgNVBAoMBG15Y2ExEjAQBgNVBAsMCW15\nY2Egcm9vdDEhMB8GA1UEAwwYbXkgY2VydGlmaWNhdGUgYXV0aG9yaXR5ggIQADAO\nBgNVHQ8BAf8EBAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwHwYDVR0RBBgwFoII\nY2VydC5jb22CCiouY2VydC5jb20wDQYJKoZIhvcNAQELBQADggIBADSL8AegMDhJ\nUgRfP6CQeAcLgbHAb9cS7vo0ju9E38pSVDBKA1VachXgwf6630XJ4/YrHzCNgbGO\neX3GcwwcD8oWopnPX4bnGdwZaQ52qd4yUNgErNFpsZU02+ohgJew1Wx+caGNQ5F3\nMqsIy8X86a5FOFCGa0CUx4Iv4JieD6kKFWzJwvXwbWS6tFUxUOlpxYZRpZj4ZPb/\nyz65PBHeH9K+A0q+upwvVdK3Gp0qbcl7ZEE3rVR1GB5VSGnyG4YG0Y59Ys0JlsgR\n2jY0zdC2DTAGQdPL6u1HsNgCDz2nzUDaYGOMb1NVRTsRZ/25irkAsOJFHP/CkuSy\nW/xogRCbX5WhHwxIzucpj+tnB7Hi9TBJLcsl7MNHuhUz5vtkl3d3dUEipIESwKmC\nn/avv2+6/8tm3UV2ji2N10246nPHZX8IddAAMwdfNriwPsz5XfXaF7czgaWYvBsu\nxkd5b2mbGH3BVJiEwRDeQRo8WGBfs0vqAF3abqjIrTiJikZpcI7GbqzhYzLcJnlX\nbOc2Xo8PXj9mE7dQ0Tfkd2wAovQ2xnuKBQgu14adFJjLFhSk5xRuu3274Kn3CmrK\nlo4FaSoSIw1vHs3J9VH6z/VRpu6dwC1U3XMSSlNSLcVO7UI/FJdacypAz6NDgpfP\nlv2z+ne8UwT6KGgPPTvCs0T5kbIXjOPY\n-----END CERTIFICATE-----\n"
	keyPem := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA3cUUcNzd9YjLG6QwNGuIRti1UDjUR6JCLB/tEt0PGH43ZUBr\nNRtvFQnTqcoCyVfb7pmYmsyVL6KidfUG8+E/Leye5kjWai9dy8fplXnJQZnIkOhy\nM5Mceo3A9IM4BBfzu0uNBsJrxHD/8Hjzwag9tpmVZgO9lDQzDmtqyK7QH44jo8vz\n5WdzaNd0+Vze2I2ghVdFnXhM16LttQMDc7XLn3eJ+3a1D9NnGVjt/KVlaFNXixSb\n/3KT/Ohv53LOujv3knLy9OGV48A3/+1kYtYXMbtryj0zHfoE+dxJnS/4R8S8/bhl\n98gMVG3KwF7C7b7fchWkhKVtoeCPqey6/wYUzQIDAQABAoIBACPJuhK8kdUdzikX\nxe+vqr5EGn5nrVoiBSu5uzhgFB+PvsDINITNeI+cllvADdMQKp3Gi6nveePGCxGe\nCREyOE/g74OaHX/lRO2txTQqAyBjAMrhuAw6oU3lsk3DHzcJ5ntDJe8BUQLSeXsF\nCdEmpU7iWgmscNuJ0PNywjjAfTWaHXNgXbcragVT/El53/fAnO36aDmd5SP4BiiQ\n984Hig9Z+B9AuqYzKour8o96+IC8eD6EzSVbyvE7WnUZiVV2Opf4mJ8qUEw1NQlg\nGScrcF5RSCJTmB1lt9/mLE1PFS2SZpt2u3iCyKPAqWLa3oAzWMqD9X45+UV2UFlV\nnrfkrsECgYEA/rEE64qKiR5dgjvZVJls6dVu2WYy+EXCSqY2mYFbzHP+rw/xs7oZ\nk39/c0QghZJXDzzxXFUgKa5oeKrkYefPBWFquUfZx/OltbWfjdk8L/z8kfpYJetB\nySELnZiq9mb0JcDPGT5TJVR/udTlCtz89VPeYVt7dOypsAF0uvSrrUcCgYEA3ujC\nvvlughdm7oqhIgaRsIZKQedXLQVb8B1X1HnrbDgnuvBXEKioxIZT6Aw73scl5IFU\n7VBA+tasm9MdwtM18wJ62XCKuN3EgAA0/XpiuageWxSMfwm4Gy2t6FnV5CM+3in/\nmEPDG4NiUqyhk8eDuuuPLWtnXpRN+HQKM5xHp0sCgYEA3dZb/bkXP6WGNxhgDRLx\nzZ6MxakBvkQsng62QfBtf+CMtfjCQxRWkKWd4k01soIreGdRp2Wx9PwnnOrkr+5T\n4FDgv2843rN245XF2qybgwTtDU0rmCOYklJJJsTCLIqyH2wYNtmVXE+ETN2FfnfL\nkPezG8Ot/cLhbh9miCzyl6MCgYEAgYU9oznLvEtcw75JYjvu62McQq7pOH+krCBg\nqFUvNfJrI3QDIurdJVPn7S0unIOawOtlLX80Qov6P5Cr+kg/ULRgLXf3IvO4+acl\nIyO5uaa1/LYz7Jz5HNGt+xQ39BeGsBA3M4IsHBB7UQ591CBZqoK07u85YPtLUtIa\nG2LzP4ECgYBHOPg3ndFMe5EBql/92nSH+RILE6ADUCa+oQUOKa5p/cdWMt6ClT0m\n6cMOJN8lMmtVzwRG/aLPhN2L/vCbtBFDBDIm8PM5gg0340uFv5Mo4p1Sf8iRZG4B\nmzl86a1/OBk4MrtJqoqKrR9yg5/BXlvwuXBJRHaLjGERxhzyhk/WaQ==\n-----END RSA PRIVATE KEY-----\n"
	interCAPem := "-----BEGIN CERTIFICATE-----\nMIIFzDCCA7SgAwIBAgICEAAwDQYJKoZIhvcNAQELBQAwcjELMAkGA1UEBhMCQVUx\nDDAKBgNVBAgMA05TVzEPMA0GA1UEBwwGU3lkbmV5MQ0wCwYDVQQKDARteWNhMRIw\nEAYDVQQLDAlteWNhIHJvb3QxITAfBgNVBAMMGG15IGNlcnRpZmljYXRlIGF1dGhv\ncml0eTAgFw0yMDA5MTcyMTE3MjdaGA8yMjM5MDkzMDIxMTcyN1owdjELMAkGA1UE\nBhMCQVUxDDAKBgNVBAgMA05TVzENMAsGA1UECgwEbXljYTEaMBgGA1UECwwRbXlj\nYSBpbnRlcm1lZGlhdGUxLjAsBgNVBAMMJW15IGNlcnRpZmljYXRlIGF1dGhvcml0\neSBpbnRlcm1lZGlhdGUwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCm\n1qRLG0twxwbxBdt/nDeUr0Ia8LtLDtvPjdVVUDTpCp4gnlkEEzHu8JXPPsIami2C\nt5vo35JW1AI223FD5eef54wZG2rXlJbzwlB+yyE5+/V/6WSKe42rePvZDCD+Ym/Q\nyYeqzObViXGnmIvta2aEYZzLeTJPzppvQws/bM+d5IhRa43JuJOVYmjPdp1cjaOm\ntmW3zQSj/00a3i/97SHoyqaJX+y2bPQIJ+yScdBSn9W+Ke3o7/WnuP0HO/ST1fZM\nyzorGbso6aGnTswFbOdWMDUpauE97SL1M6ztoaI4a0HHD8Z8dPhtAmXWbs/5hmQr\njZqBj5W4oUik9iIjUhC2l1aYUf934Om62JjMn9if/mIIA5UTorddj/wKtIsd0n4X\nq5nhJ+X4yVXi3YjqW8iegenaq6UGuvNsm6m/JRAf+5n3FuspHH4WrCgAaIrYg0ZY\nDDu5ro6zHxTcHF6j01CXlJTDEJlStoZ6N9cIKVT94pUPM+EZBq3DGlhBDKipZWk7\n+sEu7sZoQ51WoV4haMY+4Wd7ea8o4sE50eoW+DN2o9lIPHMyxY5uFD7CluUt/b37\ntCcOYAV86JWBN5htTPYAH3wXsDBU/KFSJPLRPF96cuHL6Dq++Gvlqw0rKDKQ/gKh\nDma8lZ9SjVTskqk3l5wzHyNjy7nYFSIRItGIhVbp0wIDAQABo2YwZDAdBgNVHQ4E\nFgQU6VWn4HVZKHA1C5IBW5AC8J7+15kwHwYDVR0jBBgwFoAUD+ANepMk0O9Poxxx\nMpCnxxVyHNMwEgYDVR0TAQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAYYwDQYJ\nKoZIhvcNAQELBQADggIBAAQKj6FLBiy23kqHB7iUrl3dSXjJEsPm03zApRhWhr3e\nuxGVYO/YM6RlcJlc7RiKrQAO7XMuOfGbV/TedKPYz+SAeoHCdAVmT21o9HqgwRJ2\nkbJulqIF7oRmmqFOUDIUNg+ZC68QvR9cfuhzcLsEdmfEhXvI5j6CvrhOUN3UHw8A\nO7b4kiymBVT88uXUC0i3bGeEI3h6Fz/RZLbShcvTz2BwcuqoWdInyKi+8mKNfc1O\n+HGBMjnPahNAiovaEuUGErloETdjhmSOkbPBG8h9KpkndCwclEhsBN1+skKiDzKa\nMk53cXXKjqPvPEG9dfQQu0NEnOeY3ZtyVpMqnbo+G0MtyzkozvAB5WjWlpaWZYV2\nnw/wnyCi57ruYI7UjUp+NvFDiIRlOysLC7K6xia+8m7mP8MaFJibQh0tA2UDmdXs\nwy/Z87c6KUCyDB8Hl//rLWbWg6JpHTcH+81yDkVeq2TvJkB6P8jThv51Pz1z4b6U\ndHWAMK5kLmHv+P6sw0JkE5fwszoFOaqSxABq02Pkt5+Hv2EvwxpJZvySkdp7s+Xn\nGUwXhduMscVL/Yd62ES5dYSQ+vbmZIEK3PIttcIyleif6DLFZijJnywf5etYxvrK\nY9wgX6D9PwShl32sf3nzHXh3npLdbio3XwJQUcO6c/lm49rKD7L9L5RM6FNShl8R\n-----END CERTIFICATE-----\n"
	rootCAPem := "-----BEGIN CERTIFICATE-----\nMIIFzDCCA7SgAwIBAgIJAKdYQFPloO6RMA0GCSqGSIb3DQEBCwUAMHIxCzAJBgNV\nBAYTAkFVMQwwCgYDVQQIDANOU1cxDzANBgNVBAcMBlN5ZG5leTENMAsGA1UECgwE\nbXljYTESMBAGA1UECwwJbXljYSByb290MSEwHwYDVQQDDBhteSBjZXJ0aWZpY2F0\nZSBhdXRob3JpdHkwIBcNMjAwOTE3MjA1MTM1WhgPMjI5NDA3MDMyMDUxMzVaMHIx\nCzAJBgNVBAYTAkFVMQwwCgYDVQQIDANOU1cxDzANBgNVBAcMBlN5ZG5leTENMAsG\nA1UECgwEbXljYTESMBAGA1UECwwJbXljYSByb290MSEwHwYDVQQDDBhteSBjZXJ0\naWZpY2F0ZSBhdXRob3JpdHkwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoIC\nAQDgvRdI24Rv9XBnirlB1LwS32MYyVM2mksTF52E0qrg1OKcMs1D2737BrgaUD6C\nB1I2lMAKR25Q3+x9fSutyww8KZ7yQkFcX2lhwsyYll0j1rvkjek0M1K4787ZFrXS\ncRihE6BSvP5886O+v7a30HxtKbI9oFHdbzgpLpTzvVAn53tokRgAJNtQZWpyJ5Qq\nIG7c96dG9zsXE5+tYT0E0p3ec1z/Ucdx6SKOFjCR8bVLX+Y97mxypOMaPEhGJ4D3\nBlxlCvwDo5sF46e/ntie3Fqghk3jRZTUXedB0IjN8iJCKODPMO1j1cESqVg21xGZ\nyZxIn/ra1iqx9VDCP8egfUOmmMF8flGV08qOGDLGEc/dpVe/yHvG3lmld3MBsW+3\nu6O2l7GIKdLHKibe3uGHhmuPbHq2vlc6IIlRtpsZtK3IXt+bpvlKdI3rxbl4MbT7\n8Z09IUpTsT5jDPEVRnX0zV78Gs4TyKqJKxJJaINx9n0AuXJ8b3jmth/Bb6OkoPgv\nsbFS2QER2Yp8whE1W2PMwtJ06u20YX0RSwuKD+CsnTVmtQwWLBXescCNRH372HwS\nLHO8dvyFWfekLaB2LfciJWYBd8thO5Y4O65FnKLGDvEUh6Ew2OOnhOpy4flWAng6\n39r5uuDQqmWrPFjDNR5HvQjQu1Bv0j81cFY4qZqSIskR9wIDAQABo2MwYTAdBgNV\nHQ4EFgQUD+ANepMk0O9PoxxxMpCnxxVyHNMwHwYDVR0jBBgwFoAUD+ANepMk0O9P\noxxxMpCnxxVyHNMwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMCAYYwDQYJ\nKoZIhvcNAQELBQADggIBAFdMzocjv6RojMXft1TnwYKb8H0ce6qcsBHZmd/M7IXf\nhyedRkcm7RuN7ayNjFA+44pwAr4jMMNklBQDpGD5yYt1jsltiYoYX5bwZdn2I/rr\nwNQ/FfNSp8rJWqtBhaEt0VI+snHuy0Gdx1eQGf4bJNzvsDLjjJuQ32VUjaCOzsd1\n7d8jR/yjR3Sq20oFEu3HqFSC9OCH2QORTqf6i2IkaUeJbkVTa8+uVceDDbRs3CwY\nVgk/4WcOzcrz0F2BJPpFQ4knrSuHgUbElPHPVuZcn3XZ0n1KBXZdNVCIyLVRowdr\nI+gNEgWE3670Osx55QWg7depP7hU30nQlC1cm2ej2MxM48ddbAL4Zqs8/W1gm+Xb\nDkTsfh81QZQaw6qFVGHJNRIyfMT68ekFB8AgqntulIFR2RJTr/3QJBMhGHKQkmcT\nsa0z0ZrmS/ieurRUjaCsud10Y5VbY5Y8ll5kPsuRWuyijftjcPFqHBzLSSdLacO9\nlVIGkTA3ARCGgym3v5+ZZJ4DeLOJRz9c9OCIASlCkNFFEm1aJ8oagynh2tYqe5TK\nCva1MX8QW5OjHbrm1xvQ8uZOSj55yuBQWKH47GF4QxiojzKikLv4Cpv2Tk5SR9qv\nq3C4t8B26KurNb4z99eo5XhW5XXvQdKZTQC9BqZDN7xhQlwm5lbRSuhZMBJJaQOS\n-----END CERTIFICATE-----\n"
	mockRunner()
	Runner.Connection.Downstream.Tls.Acme = Acme{Provider: "testacme"}
	Runner.Connection.Downstream.Tls.Key = keyPem
	Runner.Connection.Downstream.Tls.Cert = certPem + interCAPem + rootCAPem

	//cleanup
	os.RemoveAll(filepath.FromSlash(Runner.cacheDir + "/testacme"))

	if e := Runner.initCacheDir().cacheAcmeCertAndKey("testacme"); e != nil {
		t.Errorf("unable to cache cert, cause %v", e)
	}

	//nuke these so we can test they reload
	Runner.Connection.Downstream.Tls.Key = ""
	Runner.Connection.Downstream.Tls.Cert = ""

	if e1 := Runner.loadAcmeCertAndKeyFromCache("testacme"); e1 != nil {
		t.Errorf("unable to load cert from cache, cause: %v", e1)
	}

	if len(Runner.Connection.Downstream.Tls.Key) == 0 || len(Runner.Connection.Downstream.Tls.Cert) == 0 {
		t.Errorf("keys or cert not loaded from cache")
	}

	if e2 := Runner.ReloadableCert.triggerInit(); e2 != nil {
		t.Errorf("could not init cert loaded from cache, cause: %v", e2)
	}

	//cleanup
	if e3 := os.RemoveAll(filepath.FromSlash(Runner.cacheDir + "/testacme")); e3 != nil {
		t.Errorf("unable to clean up cache after unit test, cause: %v", e3)
	}
}

func TestCertCacheAndReloadFromCacheFailsForBadCert(t *testing.T) {
	certPem := "xxxx\n"
	keyPem := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA3cUUcNzd9YjLG6QwNGuIRti1UDjUR6JCLB/tEt0PGH43ZUBr\nNRtvFQnTqcoCyVfb7pmYmsyVL6KidfUG8+E/Leye5kjWai9dy8fplXnJQZnIkOhy\nM5Mceo3A9IM4BBfzu0uNBsJrxHD/8Hjzwag9tpmVZgO9lDQzDmtqyK7QH44jo8vz\n5WdzaNd0+Vze2I2ghVdFnXhM16LttQMDc7XLn3eJ+3a1D9NnGVjt/KVlaFNXixSb\n/3KT/Ohv53LOujv3knLy9OGV48A3/+1kYtYXMbtryj0zHfoE+dxJnS/4R8S8/bhl\n98gMVG3KwF7C7b7fchWkhKVtoeCPqey6/wYUzQIDAQABAoIBACPJuhK8kdUdzikX\nxe+vqr5EGn5nrVoiBSu5uzhgFB+PvsDINITNeI+cllvADdMQKp3Gi6nveePGCxGe\nCREyOE/g74OaHX/lRO2txTQqAyBjAMrhuAw6oU3lsk3DHzcJ5ntDJe8BUQLSeXsF\nCdEmpU7iWgmscNuJ0PNywjjAfTWaHXNgXbcragVT/El53/fAnO36aDmd5SP4BiiQ\n984Hig9Z+B9AuqYzKour8o96+IC8eD6EzSVbyvE7WnUZiVV2Opf4mJ8qUEw1NQlg\nGScrcF5RSCJTmB1lt9/mLE1PFS2SZpt2u3iCyKPAqWLa3oAzWMqD9X45+UV2UFlV\nnrfkrsECgYEA/rEE64qKiR5dgjvZVJls6dVu2WYy+EXCSqY2mYFbzHP+rw/xs7oZ\nk39/c0QghZJXDzzxXFUgKa5oeKrkYefPBWFquUfZx/OltbWfjdk8L/z8kfpYJetB\nySELnZiq9mb0JcDPGT5TJVR/udTlCtz89VPeYVt7dOypsAF0uvSrrUcCgYEA3ujC\nvvlughdm7oqhIgaRsIZKQedXLQVb8B1X1HnrbDgnuvBXEKioxIZT6Aw73scl5IFU\n7VBA+tasm9MdwtM18wJ62XCKuN3EgAA0/XpiuageWxSMfwm4Gy2t6FnV5CM+3in/\nmEPDG4NiUqyhk8eDuuuPLWtnXpRN+HQKM5xHp0sCgYEA3dZb/bkXP6WGNxhgDRLx\nzZ6MxakBvkQsng62QfBtf+CMtfjCQxRWkKWd4k01soIreGdRp2Wx9PwnnOrkr+5T\n4FDgv2843rN245XF2qybgwTtDU0rmCOYklJJJsTCLIqyH2wYNtmVXE+ETN2FfnfL\nkPezG8Ot/cLhbh9miCzyl6MCgYEAgYU9oznLvEtcw75JYjvu62McQq7pOH+krCBg\nqFUvNfJrI3QDIurdJVPn7S0unIOawOtlLX80Qov6P5Cr+kg/ULRgLXf3IvO4+acl\nIyO5uaa1/LYz7Jz5HNGt+xQ39BeGsBA3M4IsHBB7UQ591CBZqoK07u85YPtLUtIa\nG2LzP4ECgYBHOPg3ndFMe5EBql/92nSH+RILE6ADUCa+oQUOKa5p/cdWMt6ClT0m\n6cMOJN8lMmtVzwRG/aLPhN2L/vCbtBFDBDIm8PM5gg0340uFv5Mo4p1Sf8iRZG4B\nmzl86a1/OBk4MrtJqoqKrR9yg5/BXlvwuXBJRHaLjGERxhzyhk/WaQ==\n-----END RSA PRIVATE KEY-----\n"
	interCAPem := "-----BEGIN CERTIFICATE-----\nMIIFzDCCA7SgAwIBAgICEAAwDQYJKoZIhvcNAQELBQAwcjELMAkGA1UEBhMCQVUx\nDDAKBgNVBAgMA05TVzEPMA0GA1UEBwwGU3lkbmV5MQ0wCwYDVQQKDARteWNhMRIw\nEAYDVQQLDAlteWNhIHJvb3QxITAfBgNVBAMMGG15IGNlcnRpZmljYXRlIGF1dGhv\ncml0eTAgFw0yMDA5MTcyMTE3MjdaGA8yMjM5MDkzMDIxMTcyN1owdjELMAkGA1UE\nBhMCQVUxDDAKBgNVBAgMA05TVzENMAsGA1UECgwEbXljYTEaMBgGA1UECwwRbXlj\nYSBpbnRlcm1lZGlhdGUxLjAsBgNVBAMMJW15IGNlcnRpZmljYXRlIGF1dGhvcml0\neSBpbnRlcm1lZGlhdGUwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQCm\n1qRLG0twxwbxBdt/nDeUr0Ia8LtLDtvPjdVVUDTpCp4gnlkEEzHu8JXPPsIami2C\nt5vo35JW1AI223FD5eef54wZG2rXlJbzwlB+yyE5+/V/6WSKe42rePvZDCD+Ym/Q\nyYeqzObViXGnmIvta2aEYZzLeTJPzppvQws/bM+d5IhRa43JuJOVYmjPdp1cjaOm\ntmW3zQSj/00a3i/97SHoyqaJX+y2bPQIJ+yScdBSn9W+Ke3o7/WnuP0HO/ST1fZM\nyzorGbso6aGnTswFbOdWMDUpauE97SL1M6ztoaI4a0HHD8Z8dPhtAmXWbs/5hmQr\njZqBj5W4oUik9iIjUhC2l1aYUf934Om62JjMn9if/mIIA5UTorddj/wKtIsd0n4X\nq5nhJ+X4yVXi3YjqW8iegenaq6UGuvNsm6m/JRAf+5n3FuspHH4WrCgAaIrYg0ZY\nDDu5ro6zHxTcHF6j01CXlJTDEJlStoZ6N9cIKVT94pUPM+EZBq3DGlhBDKipZWk7\n+sEu7sZoQ51WoV4haMY+4Wd7ea8o4sE50eoW+DN2o9lIPHMyxY5uFD7CluUt/b37\ntCcOYAV86JWBN5htTPYAH3wXsDBU/KFSJPLRPF96cuHL6Dq++Gvlqw0rKDKQ/gKh\nDma8lZ9SjVTskqk3l5wzHyNjy7nYFSIRItGIhVbp0wIDAQABo2YwZDAdBgNVHQ4E\nFgQU6VWn4HVZKHA1C5IBW5AC8J7+15kwHwYDVR0jBBgwFoAUD+ANepMk0O9Poxxx\nMpCnxxVyHNMwEgYDVR0TAQH/BAgwBgEB/wIBADAOBgNVHQ8BAf8EBAMCAYYwDQYJ\nKoZIhvcNAQELBQADggIBAAQKj6FLBiy23kqHB7iUrl3dSXjJEsPm03zApRhWhr3e\nuxGVYO/YM6RlcJlc7RiKrQAO7XMuOfGbV/TedKPYz+SAeoHCdAVmT21o9HqgwRJ2\nkbJulqIF7oRmmqFOUDIUNg+ZC68QvR9cfuhzcLsEdmfEhXvI5j6CvrhOUN3UHw8A\nO7b4kiymBVT88uXUC0i3bGeEI3h6Fz/RZLbShcvTz2BwcuqoWdInyKi+8mKNfc1O\n+HGBMjnPahNAiovaEuUGErloETdjhmSOkbPBG8h9KpkndCwclEhsBN1+skKiDzKa\nMk53cXXKjqPvPEG9dfQQu0NEnOeY3ZtyVpMqnbo+G0MtyzkozvAB5WjWlpaWZYV2\nnw/wnyCi57ruYI7UjUp+NvFDiIRlOysLC7K6xia+8m7mP8MaFJibQh0tA2UDmdXs\nwy/Z87c6KUCyDB8Hl//rLWbWg6JpHTcH+81yDkVeq2TvJkB6P8jThv51Pz1z4b6U\ndHWAMK5kLmHv+P6sw0JkE5fwszoFOaqSxABq02Pkt5+Hv2EvwxpJZvySkdp7s+Xn\nGUwXhduMscVL/Yd62ES5dYSQ+vbmZIEK3PIttcIyleif6DLFZijJnywf5etYxvrK\nY9wgX6D9PwShl32sf3nzHXh3npLdbio3XwJQUcO6c/lm49rKD7L9L5RM6FNShl8R\n-----END CERTIFICATE-----\n"
	rootCAPem := "-----BEGIN CERTIFICATE-----\nMIIFzDCCA7SgAwIBAgIJAKdYQFPloO6RMA0GCSqGSIb3DQEBCwUAMHIxCzAJBgNV\nBAYTAkFVMQwwCgYDVQQIDANOU1cxDzANBgNVBAcMBlN5ZG5leTENMAsGA1UECgwE\nbXljYTESMBAGA1UECwwJbXljYSByb290MSEwHwYDVQQDDBhteSBjZXJ0aWZpY2F0\nZSBhdXRob3JpdHkwIBcNMjAwOTE3MjA1MTM1WhgPMjI5NDA3MDMyMDUxMzVaMHIx\nCzAJBgNVBAYTAkFVMQwwCgYDVQQIDANOU1cxDzANBgNVBAcMBlN5ZG5leTENMAsG\nA1UECgwEbXljYTESMBAGA1UECwwJbXljYSByb290MSEwHwYDVQQDDBhteSBjZXJ0\naWZpY2F0ZSBhdXRob3JpdHkwggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoIC\nAQDgvRdI24Rv9XBnirlB1LwS32MYyVM2mksTF52E0qrg1OKcMs1D2737BrgaUD6C\nB1I2lMAKR25Q3+x9fSutyww8KZ7yQkFcX2lhwsyYll0j1rvkjek0M1K4787ZFrXS\ncRihE6BSvP5886O+v7a30HxtKbI9oFHdbzgpLpTzvVAn53tokRgAJNtQZWpyJ5Qq\nIG7c96dG9zsXE5+tYT0E0p3ec1z/Ucdx6SKOFjCR8bVLX+Y97mxypOMaPEhGJ4D3\nBlxlCvwDo5sF46e/ntie3Fqghk3jRZTUXedB0IjN8iJCKODPMO1j1cESqVg21xGZ\nyZxIn/ra1iqx9VDCP8egfUOmmMF8flGV08qOGDLGEc/dpVe/yHvG3lmld3MBsW+3\nu6O2l7GIKdLHKibe3uGHhmuPbHq2vlc6IIlRtpsZtK3IXt+bpvlKdI3rxbl4MbT7\n8Z09IUpTsT5jDPEVRnX0zV78Gs4TyKqJKxJJaINx9n0AuXJ8b3jmth/Bb6OkoPgv\nsbFS2QER2Yp8whE1W2PMwtJ06u20YX0RSwuKD+CsnTVmtQwWLBXescCNRH372HwS\nLHO8dvyFWfekLaB2LfciJWYBd8thO5Y4O65FnKLGDvEUh6Ew2OOnhOpy4flWAng6\n39r5uuDQqmWrPFjDNR5HvQjQu1Bv0j81cFY4qZqSIskR9wIDAQABo2MwYTAdBgNV\nHQ4EFgQUD+ANepMk0O9PoxxxMpCnxxVyHNMwHwYDVR0jBBgwFoAUD+ANepMk0O9P\noxxxMpCnxxVyHNMwDwYDVR0TAQH/BAUwAwEB/zAOBgNVHQ8BAf8EBAMCAYYwDQYJ\nKoZIhvcNAQELBQADggIBAFdMzocjv6RojMXft1TnwYKb8H0ce6qcsBHZmd/M7IXf\nhyedRkcm7RuN7ayNjFA+44pwAr4jMMNklBQDpGD5yYt1jsltiYoYX5bwZdn2I/rr\nwNQ/FfNSp8rJWqtBhaEt0VI+snHuy0Gdx1eQGf4bJNzvsDLjjJuQ32VUjaCOzsd1\n7d8jR/yjR3Sq20oFEu3HqFSC9OCH2QORTqf6i2IkaUeJbkVTa8+uVceDDbRs3CwY\nVgk/4WcOzcrz0F2BJPpFQ4knrSuHgUbElPHPVuZcn3XZ0n1KBXZdNVCIyLVRowdr\nI+gNEgWE3670Osx55QWg7depP7hU30nQlC1cm2ej2MxM48ddbAL4Zqs8/W1gm+Xb\nDkTsfh81QZQaw6qFVGHJNRIyfMT68ekFB8AgqntulIFR2RJTr/3QJBMhGHKQkmcT\nsa0z0ZrmS/ieurRUjaCsud10Y5VbY5Y8ll5kPsuRWuyijftjcPFqHBzLSSdLacO9\nlVIGkTA3ARCGgym3v5+ZZJ4DeLOJRz9c9OCIASlCkNFFEm1aJ8oagynh2tYqe5TK\nCva1MX8QW5OjHbrm1xvQ8uZOSj55yuBQWKH47GF4QxiojzKikLv4Cpv2Tk5SR9qv\nq3C4t8B26KurNb4z99eo5XhW5XXvQdKZTQC9BqZDN7xhQlwm5lbRSuhZMBJJaQOS\n-----END CERTIFICATE-----\n"
	mockRunner()
	Runner.Connection.Downstream.Tls.Acme = Acme{Provider: "testacme2"}
	Runner.Connection.Downstream.Tls.Key = keyPem
	Runner.Connection.Downstream.Tls.Cert = certPem + interCAPem + rootCAPem

	//cleanup
	os.RemoveAll(filepath.FromSlash(Runner.cacheDir + "/testacme2"))

	if e := Runner.initCacheDir().cacheAcmeCertAndKey("testacme2"); e != nil {
		t.Errorf("unable to cache cert, cause %v", e)
	}

	//nuke these so we can test they reload
	Runner.Connection.Downstream.Tls.Key = ""
	Runner.Connection.Downstream.Tls.Cert = ""

	if e1 := Runner.loadAcmeCertAndKeyFromCache("testacme2"); e1 == nil {
		t.Error("bad cache cert read wanted error but got nil")
	} else {
		t.Logf("normal. expected error during reading bad cert from cache: %v", e1)
	}

	if len(Runner.Connection.Downstream.Tls.Key) > 0 || len(Runner.Connection.Downstream.Tls.Cert) > 0 {
		t.Errorf("keys or cert should not be loaded from cache but got key: %s, cert: %s", Runner.Connection.Downstream.Tls.Key, Runner.Connection.Downstream.Tls.Cert)
	}

	if e2 := Runner.ReloadableCert.triggerInit(); e2 == nil {
		t.Error("triggerInit should error out for bad cert but returned nil")
	} else {
		t.Logf("normal. triggerInit expected error: %v", e2)
	}

	//cleanup
	if e3 := os.RemoveAll(filepath.FromSlash(Runner.cacheDir + "/testacme2")); e3 != nil {
		t.Errorf("unable to clean up cache after unit test, cause: %v", e3)
	}
}

func TestAsSha256Struct(t *testing.T) {
	acme1 := Acme{
		Provider: "letsencrypt",
		Domains:  []string{"api.domain.com"},
		Email:    "noreply@domain.com",
	}

	acme2 := Acme{
		Provider: "letsencrypt",
		Domains:  []string{"www.domain.com"},
		Email:    "noreply@domain.com",
	}

	//we need these to be different
	if asSha256(acme1) == asSha256(acme2) {
		t.Errorf("error. struct hashes should not be equal but were acme1 %v, acme2 %v", asSha256(acme1), asSha256(acme2))
	} else {
		t.Logf("normal. different struct hashes acme1 %v, acme2 %v", asSha256(acme1), asSha256(acme2))
	}

	acme11 := Acme{
		Provider: "letsencrypt",
		Domains:  []string{"api.domain.com"},
		Email:    "noreply@domain.com",
	}

	//we want these to be identical
	if asSha256(acme1) != asSha256(acme11) {
		t.Errorf("error. struct hashes should be be equal but were acme1 %v, acme11 %v", asSha256(acme1), asSha256(acme11))
	} else {
		t.Logf("normal. struct hashes acme1 %v, acme11 %v", asSha256(acme1), asSha256(acme11))
	}
}
