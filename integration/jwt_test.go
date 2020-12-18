package integration

import (
	"fmt"
	"github.com/simonmittag/j8a"
	"net/http"
	"testing"
	"time"
)

func TestKeyRotationSuccess(t *testing.T) {
	//use token with kid k1
	k1tok := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImsxIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiOTMyOTBmZWEtOGViNS00NTE1LWE1OTgtYTkzMzMwZDE3MTFlIiwiaWF0IjoxNjA4MzI1MTc1fQ.jOHFb5NCId24hgI482fh-oPL-PaDJOWi8R9IBGg-MnpchR1oHsCNC4Atpuik4CtO-182qkjuIV-Dj85vdrsUMwe6nFhX_JvUITBbNO3bgE0REVncGfyeOZTOmLd8wnZrqnr5JGp2zaDOufEU-KMhpbWrBXZz5RxF_XgiMkXOwAdq88cYkU6DUprKirUk_HxE2Se91x5vJthfjb1sTMhHS7hvbPvmBad1OGoSBbh6MTFzg6f48-OK_Z2AhlVVLly4bp56DdJsGr1len5fJLksM9tqXtSARhk24xRPhUb2VWJwUu4PdV1yHXhT50ESQ4UM8leulvzal5KSOcSnltwBkw"
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 200, k1tok)

	//k2 should work after initial failure and background update
	k2tok := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImsyIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiNWZlNTUzOTAtMmIxMy00YmNjLWJmOGMtMTk0OGZkYmI2YjI5IiwiaWF0IjoxNjA4MjM4Nzg2fQ.DEh1rme5jg1PVka-hkRrA92kqtaQZiu8PkBduztssJrK5rKEPZOKk3EOBoq5CwiLbUh1ZF77EszYtBXUHaThi05HsUk4bIF6Qj9plY-nPtxgkSihC6m-d_FXu6qONGwNjmpgt9o-FCOmuvtzpOMh6LYRTIf_mley_w7tN-QwEIViEGGK54j6g-DPxPlxA_2MwfwiHjwtndI3JzfFWBnyOhvPoGJo9SSE4JP33neh7YOw6UZu7anZHWOSuRRun2Vb9rgr-6_NaUKWCRfd3IxcQWVH6rk_2m4AfqyWc9EJ438q_uxg5Md9sgw9qPQvkJyaeM0D6UcmEFMew-RBgiaqYg"
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 401, k2tok)
	time.Sleep(time.Second * 10)
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 200, k2tok)

	//now test we can still use k1
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 200, k1tok)

	//we cannot initally use k3 because it's not in the config until but it will cause rotating to k2b which is bad
	k3tok := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImszIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiOWI5ODk5NDctMWNhNi00Njk4LThiMmYtMjBhODE4ZjdlZDcxIiwiaWF0IjoxNjA4MzMwNDIxfQ.dCK241wxJrwAcv-eL4E_LgzeBXutfikWao5WBsJJFZDBlAf3yAHGe7SLz4P9EjqMPcn0BYiCTXGLKvkb7CEgMwKnl2GL7KImMxJZTnxHe2_45hV3KhFtfyuLMoO005eJfcD5U-27tEK6niXku8g5XRclSaVaphRJja1G0AleOFfplhhxaaMKokdpoj9E_Tr4TejUeW1LS_OvujE0idHxothGmtD3w21W9iujuVfRMCn37SxI59YcA4HlExd7zwXvb5iv9k40NSd8AfqILwVGZgDqb-uRUq3z5gOIOmNNRULNaiPrpVzhdMVdGXOUS7Gelk487RgsiUDIV619IvKpUw"
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 401, k3tok)

	//after 10s we still cannot use k3tok because k1, k2 are in config but k2b has caused the updater to exit in background
	//however this will trigger another search for k3
	time.Sleep(time.Second * 10)
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 401, k3tok)

	//and after 10s max k3 will now return a 200 and k1 and k2 still need to work from cache.
	//this assures the updater continues to work after intermittent failure and no keys are forgotten
	time.Sleep(time.Second * 10)
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 200, k3tok)
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 200, k2tok)
	DoJwtTest(t, "/mse6jwtjwksbadrotate2", 200, k1tok)
}

func TestLoadJwksBadRotation(t *testing.T) {
	cfg := j8a.NewJwt("MyJwks", "RS256", "", "http://localhost:60083/mse6/jwksbadrotate?rc=0", "120")
	err := cfg.LoadJwks()
	if err != nil {
		t.Errorf("expected remote RS256 key to pass on first load but received %v", err)
	} else {
		t.Logf("normal. validation passed on jwksbadrotate with rc=0")
	}

	time.Sleep(time.Duration(time.Second * 10))

	cfg.JwksUrl = "http://localhost:60083/mse6/jwksbadrotate"
	err = cfg.LoadJwks()
	if err != nil {
		t.Errorf("expected remote RS256 key to pass on second load but received %v", err)
	} else {
		t.Logf("normal. validation passed on jwksbadrotate with rc=1")
	}

	time.Sleep(time.Duration(time.Second * 10))

	err = cfg.LoadJwks()
	if err == nil {
		t.Errorf("expected remote HS256 key to fail on third load but received nil error")
	} else {
		t.Logf("normal. validation failed for jwksbadrotate with rc=2, cause: %v", err)
	}

	//reset
	http.Get("http://localhost:60083/mse6/jwksbadrotate?rc=0")
}

func TestLoadJwksWithBadSkew(t *testing.T) {
	cfg := j8a.NewJwt("MyJwks", "ES256", "", "http://localhost:60083/mse6/jwkses256", "notnumeric")
	err := cfg.Validate()

	if err == nil {
		t.Error("should refuse to load non numeric skew seconds but failed to provide error")
	} else {
		t.Logf("normal. skew not numeric, msg: %v", err)
	}
}

func TestLoadEs256WithJwks(t *testing.T) {
	cfg := j8a.NewJwt("MyJwks", "ES256", "", "http://localhost:60083/mse6/jwkses256", "")
	err := cfg.LoadJwks()

	if err != nil {
		t.Errorf("ECDSA key loading failed via JWKS, cause: %v", err)
	}

	if cfg.ECDSAPublic == nil {
		t.Errorf("ECDSA key not loaded")
	}
}

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
	jwt := j8a.NewJwt("Testy", "RS256", "", "http://localhost:60083/mse6/jwks", "120")
	err := jwt.LoadJwks()
	if err != nil {
		t.Errorf("error loading RS256 JWKS config, cause %v", err)
	}
}

func TestJwtLoadJWKSFailNoAlg(t *testing.T) {
	jwt := j8a.NewJwt("Testy", "", "http://localhost:60083/mse6/jwks", "", "")
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func TestJwtLoadJWKSFailBadAlg(t *testing.T) {
	jwt := j8a.NewJwt("Testy", "RS257", "", "http://localhost:60083/mse6/jwks", "")
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func TestJwtLoadJWKSFailIllegalJwks(t *testing.T) {
	jwt := j8a.NewJwt("Testy", "RS256", "", "http://localhost:60083/mse6/jwksbad", "")
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func TestJwtLoadJWKSFailMixedAlgorithmsInKeyset(t *testing.T) {
	jwt := j8a.NewJwt("Testy", "RS256", "", "http://localhost:60083/mse6/jwksmix", "")
	err := jwt.LoadJwks()
	if err == nil {
		t.Error("should have failed loading JWKS config with alg")
	} else {
		t.Logf("normal. alg verification failed with %s", err)
	}
}

func TestJwtLoadJWKSFailBadUrl(t *testing.T) {
	jwt := j8a.NewJwt("Testy", "RS256", "", "http://localhost:60083/mse6/get", "")
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
