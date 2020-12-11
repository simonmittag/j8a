package integration

import (
	"fmt"
	"github.com/simonmittag/j8a"
	"net/http"
	"testing"
)

func TestJwtInvalidKidStillPassesSecondLayerValidationWithKeyMatching(t *testing.T) {
	//this jwt token has a kid of k42 which isn't in the keyset it should still
	//pass by matching the signature once the kid wasn't found.
	DoJwtTest(t, "/mse6jwt", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Ims0MiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiMmQ1NWVhMmItMGE0NS00OTIyLTgzZTUtNDExYzgxYmM1MzkwIiwiaWF0IjoxNjA2OTQ2MTE0fQ.LqIDVbxE-k4X6EbH5g5WVzUMpBh4WS-wjGUrPpDZ75d0mTWncCVwDfgjcf9IgqAnaKmPaTa2f_DrqLLtIaZTQAfz0OZ2cakeKc7lxyju9l-1QCgcpLun1G63QLXfBxei9y1p5iRJ0WlUjgwa7jaVpBu6HkUpFBHIdSdZgnavwaj7LVHhiGiJefZlVB__z0sgIshHTellqGEAJlD6ANiB0IeR_BHT3LMfHgkL3xBzr7XXKTDw67hT0jU1JJpHox5Q2CTHVRVDsEzUH1ZBTmsjSQuoqnzql8XCqL1QD0sw0BWBxZrB4doXLjpKSPTRv1Pa2lRDOfl4Xt0asD0He5wcLg")
}

func TestJwtNoneNotExpired(t *testing.T) {
	DoJwtTest(t, "/jwtnone", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiNDNkODEzOTYtOWZlMi00ZTRkLTgxMDYtOWQyOTFlY2FmMjVkIiwiaWF0IjoxNjA1NTY2MDU0LCJleHAiOjk0NzU4NzgzNTd9.")
}

func TestJwtNoneNoExpiry(t *testing.T) {
	DoJwtTest(t, "/jwtnone", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiYjJjOTBiZDYtZmQ2Yy00MjI0LWI4MjItOGQ1ZWNlZWIzMmE0IiwiaWF0IjoxNjA1NzM5ODExfQ.")
}

func TestJwtNoneExpired(t *testing.T) {
	DoJwtTest(t, "/jwtnone", 401, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiZTFjNDhhMmQtODZiYS00NDBiLTk0YWMtMWVjNjQzYzAzYzMwIiwiaWF0IjoxNjA1NTY2MTQ4LCJleHAiOjE0NzU4NzgzNTd9.")
}

func TestJwtNoneCannotBeSmuggledAsRS256(t *testing.T) {
	DoJwtTest(t, "/jwtrs256", 401, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiYTg4ZjU3MmItNDE2ZS00OTVlLTk0NWMtNGMwMmRjOWRhYjI5IiwiaWF0IjoxNjA1NjQ2MTk3LCJleHAiOjE2MDU2NDk3OTd9.")
}

func TestJwtNoneCannotBeSmuggledAsHS256(t *testing.T) {
	DoJwtTest(t, "/jwths256", 401, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiYTg4ZjU3MmItNDE2ZS00OTVlLTk0NWMtNGMwMmRjOWRhYjI5IiwiaWF0IjoxNjA1NjQ2MTk3LCJleHAiOjE2MDU2NDk3OTd9.")
}

func TestJwtNoneCannotBeSmuggledAsES256(t *testing.T) {
	DoJwtTest(t, "/jwtes256", 401, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiYTg4ZjU3MmItNDE2ZS00OTVlLTk0NWMtNGMwMmRjOWRhYjI5IiwiaWF0IjoxNjA1NjQ2MTk3LCJleHAiOjE2MDU2NDk3OTd9.")
}

func TestJwtNoneCannotBeSmuggledAsPS256(t *testing.T) {
	DoJwtTest(t, "/jwtps256", 401, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiYTg4ZjU3MmItNDE2ZS00OTVlLTk0NWMtNGMwMmRjOWRhYjI5IiwiaWF0IjoxNjA1NjQ2MTk3LCJleHAiOjE2MDU2NDk3OTd9.")
}

func TestJwtES256Expired(t *testing.T) {
	DoJwtTest(t, "/jwtes256", 401, "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiMjU4YzQ2NWMtYjg2OS00YjdlLWEyMzItMjFlMzEzNmY4YWJmIiwiaWF0IjoxNjA1NjQ3OTk2LCJleHAiOjE0NzU4NzgzNTd9.vWcnJ1pKN9bXUhFv34I8ILT2FpDKWVVDxKgQFMaMjrduWNvXHRBvvfGy0Nax_rgLSv3BFP15xKole3eofkcbVw")
}

func TestJwtES256NotExpired(t *testing.T) {
	DoJwtTest(t, "/jwtes256", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiZjUyYTIyMzItMzY4OS00OTM0LTg5OGMtYjhjMDkwNDNlNzM4IiwiaWF0IjoxNjA1NzMxMzE5LCJleHAiOjk0NzU4NzgzNTd9.mQMZXjhrpA_0PMF-0aO56Zj9G6pON-2DZ1R0dUn1xjECkolzBNx7J8JoQ9Ik8LfFivU4YPY7C7pkob6ygA0LCQ")
}

func TestJwtES256NotBeforeFails(t *testing.T) {
	DoJwtTest(t, "/jwtes256", 401, "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiNmIzMjlmYjgtNTM5Zi00NzgyLTlmNDEtMDIzZjM2M2EwNjFmIiwiaWF0IjoxNjA1NzMxNTAxLCJleHAiOjE2MDU3MzUxMDEsIm5iZiI6OTQ3NTg3ODM1N30.5QIBnoc_XtCnukv9Yisohk8IqEyW2KPX5Ixnw2h-wwDga8v00dsV4adFuuYAz0JDyZntDeR_SQFEbU69hh3yjw")
}

func TestJwtES256NotBeforeFailsAndExpired(t *testing.T) {
	DoJwtTest(t, "/jwtes256", 401, "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiYWFkNjMwNjctYzg4NS00NWViLWE1MGEtOTFiMThhOWFiYzJkIiwiaWF0IjoxNjA1NzM0MDE5LCJleHAiOjE1NzU4NzgzNTcsIm5iZiI6OTQ3NTg3ODM1N30.FDmGI21h2aOpng2cEx6dqiAUyynnOe_kilT9raTocYcVX2xGIcneu7LHZML9mh1MOHr8e4htxe5zhiZ90Pv9rg")
}

func TestJwtHS256NotExpired(t *testing.T) {
	DoJwtTest(t, "/jwths256", 200, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.sHm-91nvgMEqZelFK6sQhWBGtshLy-A1FYFkZQ62sE4")
}

func TestJwtRS256NotExpired(t *testing.T) {
	DoJwtTest(t, "/jwtrs256", 200, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.aHXyvYSVGeNAB4wig6Nv5--vo4ZWG6CSLlrIFA8Dm_ohaVLiT_2y_cJc_v0kc_bACgdsDJloooAllh75tc870MiqGXJ4enxoUUmIKory4PUooFlFENUiDMS4EQz7JXxxhbjlfTURo00kM9hSSJ8-yluGXaX0Xirc15H2oMBDUq-1nUCjECsI6T7P4rSYxrZOviK58PjKuyocvHm_3Q0nAY0NHJ9RVYatntJFai75EbLDzzOvTDUXFjVMIQ0-nR-55pKvnykoZoqdaGJInsHHl8-M_fwxeYzvBXd_saBfMI2-KeYLUXTN9_n2z6oey4wapalJavWLxSnH75uUhOlLkU0NQWeylAiEOKXoNLBJwf5QUnwMHgrEvzZSz7-2rYG6_yuAHQ0csHRMVCIG0YKKhvgjis6tXmb4vjiR-DdJ1D2zDXVgjeQCqYdS9ycbx_CgeZ6ignalz85Xu3fm2_UwsdrrSEg9sYzzhzHl9ksD8mxlfUwcywb_HVkz77pciIcx0D8DUFR8JuWxy2AIMSWoYlWag5hVkxhIDSlcNKOuxOitPwyLA2bejxuDhZ71Kh9i_PmTzVYCrnNoNY9CKjWmC2FS_QVoDpCKcdj3T6bclucidGbtdM50fYA47jnJULcqJuaRav7otR71RmxufnrBGI6yxxmEbQPeBBFt_Xj4bRg")
}

func TestJwtPS256NotExpired(t *testing.T) {
	DoJwtTest(t, "/jwtps256", 200, "eyJhbGciOiJQUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.TbloDjGto2e90mE6yKFs-Ox9n2NCdtpm490c-NHVXCvWZrtbICO02q74tiHxIvwKMSpULlfR5YhjcO2GN4tvcimWJAbaBamG5ydfEry3sGCCmuNkPhuCTQprGFkP7ZCLJLsXwbrWCYJtkYXg23rX3MlxAP1TcebxU6x3nRRpGbjGpOUvKBgoLH5Wc0aLqShZZ12m5ck4ipHqrjLeBiqbcQD-5gQQH2ohSoDrXFgRQVDS5lJFILlSFHQYY3z_9Yxz1Y7GYVHHrY6HgNJ1INhf1SurW_8w8gzWAhkfJg0zrR6T5AtM9Z85PJSEuK--_FW4zJkvbgB7pyYD3r4Q9qQp1cpq12vQq-nTxIhHLp-zNwjdrKH03NWfnvx4_e5LrZKEKAvEpxkXEXDANrOHwFAAGcZ4AsnjOLI3UcGqXozs2OHAmXfk6byxzjXugy1j3DkgzkB2-nr0cSXqkWOd66mFVw8ZN1NdU6p6ioD05glKHZdp653YC6ySv5UO0KOGwHirP6iIzIqQXNiSzQNLq1Q0XwlYUIHYnM01h3po3zDSgGsEP1UoyhbWdBbpyw0FN26SJvAdqkWmV9okK92e62pgP49P1eD9NVOoxQx7_fqe-CAjtLBOoYJ1B-Rtbm78mFiDvSzH-uDxwcLjVrnzylwoKGfJgq3k9EN-HkxJ_IOnmjQ")
}

func TestJwtPS256CannotPassAsRS256(t *testing.T) {
	DoJwtTest(t, "/jwtrs256", 401, "eyJhbGciOiJQUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.TbloDjGto2e90mE6yKFs-Ox9n2NCdtpm490c-NHVXCvWZrtbICO02q74tiHxIvwKMSpULlfR5YhjcO2GN4tvcimWJAbaBamG5ydfEry3sGCCmuNkPhuCTQprGFkP7ZCLJLsXwbrWCYJtkYXg23rX3MlxAP1TcebxU6x3nRRpGbjGpOUvKBgoLH5Wc0aLqShZZ12m5ck4ipHqrjLeBiqbcQD-5gQQH2ohSoDrXFgRQVDS5lJFILlSFHQYY3z_9Yxz1Y7GYVHHrY6HgNJ1INhf1SurW_8w8gzWAhkfJg0zrR6T5AtM9Z85PJSEuK--_FW4zJkvbgB7pyYD3r4Q9qQp1cpq12vQq-nTxIhHLp-zNwjdrKH03NWfnvx4_e5LrZKEKAvEpxkXEXDANrOHwFAAGcZ4AsnjOLI3UcGqXozs2OHAmXfk6byxzjXugy1j3DkgzkB2-nr0cSXqkWOd66mFVw8ZN1NdU6p6ioD05glKHZdp653YC6ySv5UO0KOGwHirP6iIzIqQXNiSzQNLq1Q0XwlYUIHYnM01h3po3zDSgGsEP1UoyhbWdBbpyw0FN26SJvAdqkWmV9okK92e62pgP49P1eD9NVOoxQx7_fqe-CAjtLBOoYJ1B-Rtbm78mFiDvSzH-uDxwcLjVrnzylwoKGfJgq3k9EN-HkxJ_IOnmjQ")
}

func TestJwtES256LoadedFromX509(t *testing.T) {
	DoJwtTest(t, "/x509es256", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJFUzI1NiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiNmJlYjdmYjItYzhiOC00ODc5LTg5ODItYWMxZGQ0NjFhZDllIiwiaWF0IjoxNjA1OTk5NDgyfQ.jyKZhJp_DYxjoaL3xt_-aP0HLNPwi0A18OfhyB60ByNFCIghN4eed-N0gAualQgPzMk5z76nyhLsswOc2FBDFQ")
}

func TestJwtRS256LoadedFromX509(t *testing.T) {
	DoJwtTest(t, "/x509rs256", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiNmExMTUyNGUtZGNjYS00YmY4LTkyZTgtMTFhOTEzOTZkNTRjIiwiaWF0IjoxNjA2MDAwNTE1fQ.jngGY_yVkrtTuLaFhoHgDQI7Htqseeg7UWxmICsrkBZVN7NNjZEnLs4MLs32b_PsRRS5V6DQzMRQWpBiGnJrfKeQeKPOkNiO9Eq5rNYbYUWpzldk0uIxNPyM6aJViuKFON7ixiEU_GEy9hU5Q1JbVfMOPRXpMkX37-weY4sINCjCOvqNf8NUe9daKfv63qOjn6UOjvaS7hXbiH1f6B39cYzFx1uKcyOIR8k4ACHwA05suCSRV4kBIV3zmhrvFdB2h3lN8TWM3QNVRiNHrqc7gKP1HtX_XOSlyTTLNhGQMRX8v9rAhLIaqXnlLqakvCHNi71HdQebUhSa64fLYId0lQ")
}

func TestJwtLoadJWKS(t *testing.T) {
	jwt := j8a.Jwt{
		Name:                  "Testy",
		Alg:                   "RS256",
		Key:                   "",
		JwksUrl:               "http://localhost:60083/mse6/jwks",
		RSAPublic:             nil,
		ECDSAPublic:           nil,
		Secret:                nil,
		AcceptableSkewSeconds: "",
	}
	err := jwt.LoadJwks()
	if err != nil {
		t.Errorf("error loading RS256 JWKS config, cause %v", err)
	}
}

func TestJwtLoadJWKSFailNoAlg(t *testing.T) {
	jwt := j8a.Jwt{
		Name:                  "Testy",
		Key:                   "",
		JwksUrl:               "http://localhost:60083/mse6/jwks",
		RSAPublic:             nil,
		ECDSAPublic:           nil,
		Secret:                nil,
		AcceptableSkewSeconds: "",
	}
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func TestJwtLoadJWKSFailBadAlg(t *testing.T) {
	jwt := j8a.Jwt{
		Name:                  "Testy",
		Alg:                   "RS257",
		Key:                   "",
		JwksUrl:               "http://localhost:60083/mse6/jwks",
		RSAPublic:             nil,
		ECDSAPublic:           nil,
		Secret:                nil,
		AcceptableSkewSeconds: "",
	}
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func TestJwtLoadJWKSFailIllegalJwks(t *testing.T) {
	jwt := j8a.Jwt{
		Name:                  "Testy",
		Alg:                   "RS256",
		Key:                   "",
		JwksUrl:               "http://localhost:60083/mse6/jwksbad",
		RSAPublic:             nil,
		ECDSAPublic:           nil,
		Secret:                nil,
		AcceptableSkewSeconds: "",
	}
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func TestJwtLoadJWKSFailMixedAlgorithmsInKeyset(t *testing.T) {
	jwt := j8a.Jwt{
		Name:                  "Testy",
		Alg:                   "RS256",
		Key:                   "",
		JwksUrl:               "http://localhost:60083/mse6/jwksmix",
		RSAPublic:             nil,
		ECDSAPublic:           nil,
		Secret:                nil,
		AcceptableSkewSeconds: "",
	}
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func TestJwtLoadJWKSFailBadUrl(t *testing.T) {
	jwt := j8a.Jwt{
		Name:                  "Testy",
		Alg:                   "RS256",
		Key:                   "",
		JwksUrl:               "http://localhost:60083/mse6/get",
		RSAPublic:             nil,
		ECDSAPublic:           nil,
		Secret:                nil,
		AcceptableSkewSeconds: "",
	}
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func DoJwtTest(t *testing.T, slug string, want int, forjwt string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080%s/get", slug), nil)
	req.Header.Add("Authorization", "Bearer "+forjwt)
	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error connecting to server, cause: %s", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	//must be case insensitive
	got := resp.StatusCode
	if got != want {
		t.Errorf("server did not return expected status code, want %d, got %d", want, got)
	}
}
