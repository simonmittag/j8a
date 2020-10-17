package j8a

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestServerBootStrap(t *testing.T) {
	setupJ8a()
	resp, err := http.Get("http://localhost:8080/about")
	if resp != nil && resp.StatusCode != 200 {
		t.Errorf("server does not return ok status response after starting, want 200, got %v", resp.StatusCode)
	}
	if err != nil {
		t.Errorf("GET req to /about returned error, %v", err)
	}
}

func TestServerTlsConfig(t *testing.T) {
	tlsConfig := mockTlsConfig()
	if tlsConfig.Certificates[0].Leaf != nil && tlsConfig.Certificates[0].Leaf.DNSNames[0] != "*.jabbatest.com" {
		t.Errorf("failed to read TLS certificate")
	}
}

func TestZerologAdapter_Write(t *testing.T) {
	zla := zerologAdapter{}
	//this shouldn't crash
	zla.Write([]byte("i'm a log line from 127.0.0.1 there's no place like it"))
}

func mockTlsConfig() *tls.Config {
	certPem := "-----BEGIN CERTIFICATE-----\nMIIEkzCCAvugAwIBAgIRANiwkh9AuRgrvYh7Y5DtWIUwDQYJKoZIhvcNAQELBQAw\ngYExHjAcBgNVBAoTFW1rY2VydCBkZXZlbG9wbWVudCBDQTErMCkGA1UECwwic2lt\nb25taXR0YWdAdHJvb3BlciAoU2ltb24gTWl0dGFnKTEyMDAGA1UEAwwpbWtjZXJ0\nIHNpbW9ubWl0dGFnQHRyb29wZXIgKFNpbW9uIE1pdHRhZykwHhcNMTkwNjAxMDAw\nMDAwWhcNMzAwNzMwMDExNDU5WjBjMScwJQYDVQQKEx5ta2NlcnQgZGV2ZWxvcG1l\nbnQgY2VydGlmaWNhdGUxODA2BgNVBAsML3NpbW9ubWl0dGFnQE1hY0Jvb2stUHJv\nLTE2LmxvY2FsIChTaW1vbiBNaXR0YWcpMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A\nMIIBCgKCAQEAsCTQ9rLTQYjIlGF7EOrTJux8E514TUoAuQ0xo1NOSssptjmDyGhb\n8K7+A/TgdU/xlPMcJf22nNDQ2MpqpgHGlDcuXt3SmVrcsTeby1Pa81gxKp23a51B\n8xAoHoHwXVSWdiMWk3H/Jjv/dtYL1L180neewcWvK26ANUwlzWG6BW1QVUXXNdRo\ndmxQ1eg2S/qMBASFj6QjCsWWJiEfmz4PQpsP8q5IqCcX85BUqGO919JlE/eXEAgk\n9Yuh61/50n39B/sPC0mU5s6vH0SPCBvz1g8SiXa8jj3jCXxa/0ZsYtAVqPe5BoRP\nvK2q1sbKbJVr7EpmiOdKxKPHonRHasweGwIDAQABo4GiMIGfMA4GA1UdDwEB/wQE\nAwIFoDATBgNVHSUEDDAKBggrBgEFBQcDATAMBgNVHRMBAf8EAjAAMB8GA1UdIwQY\nMBaAFMNEcloV4jg+eonB5omuJvQXiqiRMEkGA1UdEQRCMECCDyouamFiYmF0ZXN0\nLmNvbYIKamFiYmEudGVzdIIJbG9jYWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAA\nAAABMA0GCSqGSIb3DQEBCwUAA4IBgQDGI3EUWPKsEOqLCpnwSlFihu8n9+g4pV3/\njItYhUqMBz1v8TqV2zykkJUtlfNoxrp5OAg4CG0Xr1zhqjub3teKbsNKlRpV+h04\n4ncltpe66u4gg9RW+ww/f+J3C2yZRIX+brhDcTpdEMyfVoCV/5jeCxWf29MdFcLU\nBfgFdEp1oe3bK/dyZc8SbUlmizyumaDOaZACihz/DKsJ+lzRdy6c3UPQgC3r72oN\nLx/ccpnwdeumWFs+qYOjYfrCGFXaabokdtyit4XURFngxpnPUB9jHDvkI5+/eTaB\nSpdjJxE6x4mciyZSvshhu1v8j52+d9zUANs9+Y/v6EoCZ6byaaS4NAmTXdAWlnYb\nhIuRRsI4gIDhJWLrACBu1Osh7ZknaLNVMt5xo3TemCkVKud3NHGbycHTUoFBuHz/\nJOTQJ/Z1Ym3enpTAESZVcZTzS9gL62wfIfLcFvq+tVjoJZVJCcolP2fYn3U5lEiN\nvZvs72xp4sYEOa9zhvEs/yte9c6rkU0=\n-----END CERTIFICATE-----\n"
	keyPem := "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCwJND2stNBiMiU\nYXsQ6tMm7HwTnXhNSgC5DTGjU05Kyym2OYPIaFvwrv4D9OB1T/GU8xwl/bac0NDY\nymqmAcaUNy5e3dKZWtyxN5vLU9rzWDEqnbdrnUHzECgegfBdVJZ2IxaTcf8mO/92\n1gvUvXzSd57Bxa8rboA1TCXNYboFbVBVRdc11Gh2bFDV6DZL+owEBIWPpCMKxZYm\nIR+bPg9Cmw/yrkioJxfzkFSoY73X0mUT95cQCCT1i6HrX/nSff0H+w8LSZTmzq8f\nRI8IG/PWDxKJdryOPeMJfFr/Rmxi0BWo97kGhE+8rarWxspslWvsSmaI50rEo8ei\ndEdqzB4bAgMBAAECggEAc9nDFn7HM3MjeXQj3RyVhCRF9yC63xqtHwjufN1twQOe\ni5uIcWcyETsHFtMYThAmdDDxcotMcBdnRS7cthK06QbiGMMMoJCCVoyciz674xE+\nRSk2WjE0DwmxWV9dGAVqcIjjcFap2hvcCez+Gw4F6ueCIzBB5e7npCZRNqPwFWCY\n/og06ypz/4LHXFNatvRJC3qhWFwFo1bVC1ycZmc5RQ4IHeQHzi6oCJhSRdCbM7Q9\n7fQhmjtcw0pxvJTVV+XP7tTf9iDDwgi/Le2iEqNQ1D6c4+nYAGYj2D3919oUYnyv\ntnznZ2GTibIyP6kl4L79ChRz0JGBzraKH7aJh9H1AQKBgQDS0xd8RNLJ5hkwPitt\nx/4RNlobGGZqqvQiCKkaDvnc1E1eKR3rVlxH17/ccA/qs0vcoFDPbGyRa/OgD7p5\nRo3R+EPDFoFq1KMP4SRoGcDgrNKHQ3o06sUngmtUUP28G+DUQm46xW7cnQrvvNiK\nf9MMfeNH+tAPtssZh0HNKJa9fwKBgQDV40peDqh9b4Ag+mfiHIJGuwTN8LMvKRqr\nN5dVirl2BDYMwF6JflIwIwjBZq7ah3NsT5/Yd+nuY+ux/pO4iU5jMbTtoAOv2dc7\nVKpqNTaQdJhta5OlOdBSP5iXj4siVCMIFL1jz8JtWuXX4hUtbliG6ICZTH2/5ivG\nfaPiOhAlZQKBgQCp941jnojiRSPhhP22UBpA/jS+y3kmXhTcq2bJn3FJ289ULon0\nhXd4ZDRGIAJ1EYADqyv7TkppI0MStBt+UqdbtG/NBIPqAOxFjRmw47JgcHR6oKgR\nqYSxSbAGFhW6Zi9ocPY1Y57xNZrvlKxvXIZl98gY69h6EsDDIAyoviRpOQKBgC0g\nRjlv+EZ2tt6+VhqTjzzjClF03ikuD+1dzjUDDrwCiXDJSWjS2P5E9fzv8CY0+7o3\nVm8yZY2hUUH9hycg+QPeoeCcqQp5+HoRE99SmM+DegFj+AOdHgGsX0Jiy6UTgUyc\nK5UaaVfvHJ0emv85z72u4ir1w3YwVr4LFf+N5ogtAoGAHODQpVC7sg+nlbeSKsPf\nRbULfOG4YD5pHszNM+nCjNWs00ofJoZOFA64qXwTIc4Vrh8JLiwAkXiTGYM2guv5\nQnp+HbFi/tAc+rQu4SGBaVIglnIj7jFNdgJOb68Vw/L9v2jW1Y8VoAC0eCRWpHud\nGsMkN4GFOfQKoBI/aCXn4DM=\n-----END PRIVATE KEY-----\n"
	cAPem := "-----BEGIN CERTIFICATE-----\nMIIE0zCCAzugAwIBAgIQB2bsiI7SUtxu+HwBxuNtpDANBgkqhkiG9w0BAQsFADCB\ngTEeMBwGA1UEChMVbWtjZXJ0IGRldmVsb3BtZW50IENBMSswKQYDVQQLDCJzaW1v\nbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBNaXR0YWcpMTIwMAYDVQQDDClta2NlcnQg\nc2ltb25taXR0YWdAdHJvb3BlciAoU2ltb24gTWl0dGFnKTAeFw0yMDA1MDEyMTE2\nNDNaFw0zMDA1MDEyMTE2NDNaMIGBMR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1l\nbnQgQ0ExKzApBgNVBAsMInNpbW9ubWl0dGFnQHRyb29wZXIgKFNpbW9uIE1pdHRh\nZykxMjAwBgNVBAMMKW1rY2VydCBzaW1vbm1pdHRhZ0B0cm9vcGVyIChTaW1vbiBN\naXR0YWcpMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAzivKfp5OiWpT\n362cVgbw9DBqwMP0pO32aP79Y4UYeAxCfaWQDdqQEatBdraShtZcvUX8vZ9jvgHE\noGMGSJb/DIVRxIDfhdvhh4qGQgbbSLwDkfLJTkpGMdONa/5yDC54fNZjF095YZn7\niPmsFbvYUfTwpM8qrP+jZzobByrTO4rG3Ps080gIR08RCA0E+uLg58rTpnsdBKZ0\nK2uuE4B4lVAs2AeS4KPMrH/rnCjSZz4KRwnaGqh+wiAjO0PHAfrbrhNsFB6P1/Zk\nCqzclj3TXdkMDaXhSvt0qJPEpNIPQMkvj9GROom7hExZUT7t7LPOZwODtiR2VjM3\nDDehfLqpNPRrxU3aOR7b4lFVtEL1+9NXKc3rnR5T2xPVVvBxx8FqYAxFmQtkGqpA\nYlRxImBONBreIr5/fdkr5xqd/S0s1pb8ubuK7x5COfqf0Mv++j+UjMptBQ3kYvOh\ntNrbnEI1q/7kvHNB8ETtJ4hqXikl9EHMYWdOo4nyGd4P8jo9jmGVAgMBAAGjRTBD\nMA4GA1UdDwEB/wQEAwICBDASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTD\nRHJaFeI4PnqJweaJrib0F4qokTANBgkqhkiG9w0BAQsFAAOCAYEAb+K3HO2AlDed\nS2yT7GnxD75Hcjnv1tMvMIlh1EOmRMHrzbsi7jv3Z7SDe2R5s1qRku3nxbVWj8i8\noRBi5GeRE+q/HkVloi4WPmgFGxUUbkWszAFSSGN5TAs72e5sCG/wMyEa0Gj8cOO1\ndK5SH3thP8+OjSpgQXToYfOimILlk7Hj7EgKE5Y8YX8UV+41LhGkzeK2UX9dBZn1\nof9qBc0dAQVlAA/O3dOgXorgiDbNT38cjignWEwVYzjeuJCYB91Ixf0CfHJZKHZR\nZCdIAHTJqW1tx7vsbrcl0PVAMgm+rkHLL0Dh9cp4fvONXWygVSjbqKM1s8UI9bFA\nbWU5Z3MhEn25wZCXLQDIq0uC+FwCxyS9e/exL4wmYpCLmRKVCp2gUa78Rlr/FJNa\nH9kfvP41Ya+fLzDWNKAlYQgizpZJmZuhPZu7O6n0UusaI+0WTKblCFUQJkx4aKEv\nio8QmLzoedmvVpO9Zp44Lyabmc7VnjoYTOcZczx4ECwEdKH/jswc\n-----END CERTIFICATE-----\n"

	r := mockRuntime()
	r.Connection.Downstream.Cert = certPem + cAPem
	r.Connection.Downstream.Key = keyPem
	r.Connection.Downstream.Mode = "TLS"
	r.Connection.Downstream.Port = 8443

	tlsConfig := r.tlsConfig()
	return tlsConfig
}

func setupJ8a() {
	ConfigFile = "./j8acfg.yml"
	Boot.Add(1)
	go BootStrap()
	Boot.Wait()
}
