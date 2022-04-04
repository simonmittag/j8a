package jwt

import (
	"fmt"
	"github.com/simonmittag/j8a"
	"net/http"
	"testing"
	"time"
)

func TestSubGarbageTokenJ8a1PassRouteWithNoClaims(t *testing.T) {
	//this route has sub: h@tzl which is known garbage but does not validate any claims on this route
	subgarbage := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwic3ViIjoiaEB0emwiLCJqdGkiOiI2NTQxNmRhZC00NGM2LTQzZWUtYjRmMi1iNmEzYTNiYmU2ZjMiLCJpYXQiOjE2MDg3NjA0NDR9.RR9FrL1zthYSkmTgUVzxHQFqx8qgxTr9oVhCR7Q9avQyAn2pyZteWLvo7uf473EMXolBfM1sQyQ31vVYLgGBPl82pV6OP06q2AMvfqiT-mxHD6JDmDWRJqfv4GTvOaWDLa2rWVOuPaV0lrTLLDZiIaz7ggI_CKJqPCrj9-siE4vI0hmhLEHQIA4gwbTHvngpqytz2Vcwvp8RrtMQt0c7Zm7HwU2WLgHjJwWJkLKZMV7m5OPeN14drnPhd_zI8ITNE8-AqZC8b6OOBFDI94lURPEACpBT2XwyR1XRZxbnh27v5eAWTqT9f6D1rdrNxKlhnWylkQysFCm1kEOwGFDqtbbExzY8ZtgDxUijUP7kNsr2VzxOISRmlCJ-45TnFV1Cmn8RNKQnpdjazqO_YAanxZ2SlYhE3_yZQXqZAz0vxhvBHkypVjVMxkx90ne1AS64DSVD_QhDswjNHbCYTDpSrASRk9jm_WMqVtRdkIn8mVbhE6pZ4DL0wZoO0F7bx766LTl2wZf1KoGBpIgOiQVcDKKfEt6yWTAzoLkXXG2NphLoG_NnokINUjtn6xUM4BKPUEDeOC48ByHy129GFSdHm4rrvgFOOl47YotTCvC7R0nPfEOqvS5pdnBahtlHnsTLxffbupA9NV9P09lmjsLgFV1-5nvvEF9f7yAKx2iQ1g4"

	DoJwtTest(t, "/jwtrs256n", 200, subgarbage)
}

func TestSubGarbageTokenJ8a1Fail(t *testing.T) {
	//sub: h@tzl needs to fail, garbage value
	subgarbage := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwic3ViIjoiaEB0emwiLCJqdGkiOiI2NTQxNmRhZC00NGM2LTQzZWUtYjRmMi1iNmEzYTNiYmU2ZjMiLCJpYXQiOjE2MDg3NjA0NDR9.RR9FrL1zthYSkmTgUVzxHQFqx8qgxTr9oVhCR7Q9avQyAn2pyZteWLvo7uf473EMXolBfM1sQyQ31vVYLgGBPl82pV6OP06q2AMvfqiT-mxHD6JDmDWRJqfv4GTvOaWDLa2rWVOuPaV0lrTLLDZiIaz7ggI_CKJqPCrj9-siE4vI0hmhLEHQIA4gwbTHvngpqytz2Vcwvp8RrtMQt0c7Zm7HwU2WLgHjJwWJkLKZMV7m5OPeN14drnPhd_zI8ITNE8-AqZC8b6OOBFDI94lURPEACpBT2XwyR1XRZxbnh27v5eAWTqT9f6D1rdrNxKlhnWylkQysFCm1kEOwGFDqtbbExzY8ZtgDxUijUP7kNsr2VzxOISRmlCJ-45TnFV1Cmn8RNKQnpdjazqO_YAanxZ2SlYhE3_yZQXqZAz0vxhvBHkypVjVMxkx90ne1AS64DSVD_QhDswjNHbCYTDpSrASRk9jm_WMqVtRdkIn8mVbhE6pZ4DL0wZoO0F7bx766LTl2wZf1KoGBpIgOiQVcDKKfEt6yWTAzoLkXXG2NphLoG_NnokINUjtn6xUM4BKPUEDeOC48ByHy129GFSdHm4rrvgFOOl47YotTCvC7R0nPfEOqvS5pdnBahtlHnsTLxffbupA9NV9P09lmjsLgFV1-5nvvEF9f7yAKx2iQ1g4"

	DoJwtTest(t, "/jwtrs256", 401, subgarbage)
}

func TestSubadminTokenJ8a1OrPass(t *testing.T) {
	//this token has a subscriber claim with value of "admin" it needs to pass validation only with the "admin" value
	subadmin := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwic3ViIjoiYWRtaW4iLCJqdGkiOiI2OGQ1NzlmYi0yNjNjLTQzNWQtYTcwNS1kZjM0ZmYxODk2NmQiLCJpYXQiOjE2MDg3NTc0OTl9.r9FbFyAxrqSNMco4q6Pm-Ksdi727PBPj3FCYPYazvP9FC_z3yRR4fZYk9IZbXBiP6so6WG_9W6XFuHDyy1vk0RdWnr1DelqPKWeEh-ixIgY4Lb-i_iPNspE54l3EvkvzJFD7-ZYr0mFq_MoLkDG5LdAZ-VYt-FSNzidDopeh2XSDPZf5bGWjR5lqyubpFIYdY-qyG5EGHHuIQH0d_XT9NyQtzqYSJZ_nLWECXaPtz102DeUaT-TGX3Qs7uJRKuHtZelES2quWGWTAs32OB5Bq93Cy3L1m9_HS_TP0QwyYBORbh51gmadtUY_PsciobbjTLMnciX8c76tnTns2HlaAcS4lI8DstqBm59rPEFvGv9Ptcf2NBQg8BTAazha22TUKh7gDWqbuME-LS_fP9ujXBoZHPbBRnei717butQ0r4_Egfcv9kY2PiV_9Kh8Qhz1GyeBPx2QCukbvDSzqU1YwEP6zkfp1sN-RZqHO3z1fIF0a2I2iULw2FxT98gH_zs-tyUu70PHicsooDCQHf6Sv4bJJBUQspZmjIBVDQx7LuB1BCgI-NwDkH1eklyRmf3BARlyGAzpmTauedwLH0wHM5Zq3Ew0OBmG-jALvNHbYFiNaY1IEFg1ZLk6UDxLLleabySDsktrgIp4mFQkUYqKp8-NyU8GkqxLy5ZR1-PbVG8"

	DoJwtTest(t, "/jwtrs256", 200, subadmin)
}

func TestNonceTokenWithoutValueJ8a1OrPass(t *testing.T) {
	//this token has nonce claim with empty string value. it needs to pass with any value
	nonce := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwibm9uY2UiOiIiLCJqdGkiOiIwMmQyNzczZS0zOWMxLTQ0MTktYTk2NS04MzEzMjg4YWQ0NTkiLCJpYXQiOjE2MDg3NjAyNDh9.PGYNznJfjTmwuLjploH_OJaEhilt_XvB2cFuPV6EcYzmfp5-8eqlJBCITJX8TZ4Zahe3Df0M4rR_P8DGVS5frkAhuGzl-_OMqTJhfLSlP19XZqk_0P8KG62s3ffb0esjVtFdTc3ZHQheo3Uhs7cr3QjemO7I8M0pdgekg25YMmaKn3EZ4jRsMFikTyiKcEx1mOJB8oE74DVGa1MiMufmr7qw7IrKTpRHxPwUy1IMGX9_wPs4EAUVp7nEvVlxgs5buIHp1sN3vK1Z0x1oOGbMBnxPJckP8OOSHxiQ4jDy6zCoRWRzvyvYNKFxf-F0rh2tBMebCAESN7o1dmQ6e8BeHltFaqTKurfEGLoLf7I61oWC2ZXpb3Tion1YEnO_EiKc-AghcXwt7MJW4Tf7IL08xxDwADeTjycO45zNKjqowJmC-MerXO6BObCEo-Q90pyCA9OI4KYkFXSRjwsuDib7yDlWLtBERIV6QdmsxqRUtRT1GtySeO_ypSSMhfAMRKuVCnBN1NpqwtyRm5BbUK2o3Qp8_6T_hhmii2qLaGShhfAZZO2W9OzZFux1EEZJucZvSPEZLGbJsAHCaCei992xk7JkNMueEbmx4Te_AZpXRooG2blwfK9jpuBMBg0aRtdFvlENzvWSe85l-PqvVjIjrMDmLVT4biP51uEkJ-U0SHk"

	DoJwtTest(t, "/jwtrs256", 200, nonce)
}

func TestNonceTokenJ8a1OrPass(t *testing.T) {
	//this token has nonce claim with value of "sessionID". it needs to pass with any value
	nonce := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwibm9uY2UiOiJzZXNzaW9uSUQiLCJqdGkiOiJlZmVhNDU5Yi1mNjE5LTQ5ZTAtYjc3ZC0zNGQ4ZTJkNTIxYTIiLCJpYXQiOjE2MDg3NTQ4MjF9.gMsebvl3wXpa9QJMGiEDyvzd2XgQBGaNo9H1Qq6Hj8igYZfki_gUtJ9efth6CHfQcFeSxPz4EKDKTtvwpmU843kXdY1eOQbaKWel9SbOcIesczQEeVqtRfDyw6CPdKFc4Qg0e7M4tsXn-GsLTnP7T5TKL2Zg2FzGvZ8uJTvRMSKhpKVpIIzreLff6mTTYvx8G6_vlChNtpvrc3bFaTUY1QbAHjzuySIFXltE67WdCW4os3nzYygzoo34RoqmTM1pb6VIlDCIXScefAyfVSMW7S_dldb0bCuEmcCXv8xIh_qyq5kLnGlS0VLKcDbtFfk3gLhOVEAjcKnrLvg-E1q_PpO7ZcGVjDVp21GVU3KQWNmkCjQiUL9-ayYQhHnoHCdT_61A5ec8HE5FNbkLSNVuX_sY7eNRvovJtdk5XpVSKy_vrtFIoR6m6uw-JRzQQppTauemJeb_Sb8dKo2Wo2P9umTPa6WTZFbZw4AiArEplklakzjAuJLnnT6yCXnYFAgFcoi9-JBs16eOUeN8cDb6fFIWINl89pnudjOy6SyDMX2evXEBi9YqAjjkspWe0dbNAh-141PiD1E-GjkdbKqR7IteVtC2RjZyBwQ265RAFq0bxA5Qos4uudEFRtOfocBivSDMF2zLWX9u5xetnkftIXK8moq8N_nP0_7yDcGV9nk"

	DoJwtTest(t, "/jwtrs256", 200, nonce)
}

func TestNoClaimsTokenJ8a1FailClaimsValidation(t *testing.T) {
	//this token neither a nonce claims nor a sub claim"
	no := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiZTVjYTc2YzEtNTZmOC00OWRkLTk1ZDUtNTgxMWFkMDM0YmNkIiwiaWF0IjoxNjA4NzU1NzMzfQ.FJRfBYGZOe7gHjc_J-GSQISByuMs_UzFZW20PVKPfO9T4C1V_1inlMO3q8oPOrp1eLYlgJ9FAd910Ct2XzlnFAIljCU5rNTA8HlDgwE35iK7ashptYWTIT7XbXl3PmQLfQncrnDjC99deRF3w4FGi4KrpwfCPmQB1m4zlh5EikEMIOj_EHQeNsuoTM-Vp1SAmvAsZJAOjp2704x21TbMbSRmMqxT9HqfKwze95tSDPZaC7TQQYuEOGflx3QXmq2J4OyEBqc2HinEtpYc-yPMiJHYEJXzXbO97gR6Geuvhy5N5e4kFy03624pQvprLFDYr_U_TxehZG1J9Id1rj8IfAVBVJB_LPM1NyFQ6zBimfLAUKWWiFhjZSrKsy0BfuZyxIhMh0C3b1AA9PuonOlwfeBKfaX3DumqiR8cj30gqI_LE-6gteP3QYUOlT02DoJyRY6hTUB57QKXLunUvTbDNX44u9Tu3GL5Jco8IErCPRom7WJ5hXZQQF5F6MQSqifVq5Tldyvy4dNM5vynWib8WA_nMiDXEWZwfsc3YzVp7xp9ykyt0907jl4tzRFYb3TOziZkigC4u2BSZkYBirnbdPNWcUdRGPpzOT6-Z8NzF53nDK5I45GjmARyinYM6i9d1-6SV-kWRqIFE994x6adFmCnGdYFrde6vAC7z-2pOXs"

	DoJwtTest(t, "/jwtrs256", 401, no)
}

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

	//we cannot initially use k3 because it's not in the config until but it will cause rotating to k2b which is bad
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
	DoJwtTest(t, "/jwtrs256", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImszIn0.eyJpc3MiOiJpc3N1ZXIiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwic3ViIjoiYWRtaW4iLCJub25jZSI6InNlc3Npb25JRCIsImp0aSI6IjhjYWMxOWI1LTBiYzgtNGIwYi05NWQwLTBlNjcxNmMzMTlkYyIsImlhdCI6MTYwODY3ODc1N30.Zr-aeI9ndt1iOORuZ7_w00zp3QpbCu87Mcl6ixgvxp8b6eXJdLMisAAvf_eyTmWOt7s_7_I_Vl3UG7WwYFeUUpamrou-uiCues9npEOFWLoPUoB8QmA10aBosGNqW8hGZlQDL7TMcUIejpXs9qc1VJ5LD3hVv4Tm16al8FmL7UoQqGHAMpMRi6dDhUZ45d3bfKtwNpJYNWDQuc1Gy5Sss7xoJiVJ64YbcD5M-7O49MmXv1hUOjvZ36SLtLkFYKi68OhY8fJvywLihFzj1rTqf5Wt8sV0GMRXYopfNBmLvetMvSSHSeUYKvGeL5qzIoblH1pWeX_4IUEEAHTLJ3RkTw5xfgaPn1NBFxzO4Q9I5PwDGZ_BESNmbvSe4gfrNiEPVIHjCnTlcpFkfXrgzGPkehMP4pWLIG-7bvy2lX5-6vnACLccmUnh5i7xpeXEOhXCBG9HHal6OuTFko4Y6ObolNWX7KTjIhf8QTR3WORehNLrJQFYNHvzHnTJtZNpN5z02JnY6LTDK9MJGbdkx0HgRWdGveOsobAib_54o7bM40dmcZL6FMfEXY8BtxRGiwiudX5Q9_vl1Ib4gaGFxlgHWSy6u-MiL2te_XYRyB75iEbiIMEIprd8RJwoTKAbKNgR2D7j_Kh5gGXe3G3ymt_y9FOw0HmxJqRpkivgrEx7Mhc")
}

func TestJwtRS256FutureExpiryPass(t *testing.T) {
	futureExpiry := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJzdWIiOiJhZG1pbiIsImp0aSI6IjEzNGU3NDc5LTFmODktNDM2Ni04MGQ5LWNhYjYzZjdjN2ZhNSIsImlhdCI6MTYwODkzNjQ0OSwiZXhwIjo5OTQ3NTg3ODM1N30.KPdo9MWn23ew9IHJ9s25wz7y7aCPSodThKhiOr1WyRRRjTVQZKbsGnKNHPKUTRR6OpfPBNsuUxZKXEgGCrZJ37sP5l6WsESeoj6Fla9_GxBxzHZh852x8hNqzGJoKsoVegFlEKcprBgU1V1T0lyGyibfb6mOd_iHtvR6vXb9qAHDoCSo6y-hYMuHVqPxOx9nW1Bdl9xHOub9MYmTvhFb6yUc_FFId95zQ2WJUlhPZsdaMWnYxyAPCbRnEjeZJpYY3APEsv8CLrWPijB-MY0S4W0TyQTIT-hVB6k1lwqXeLzwUwDo7KtkrWnGzdcKdzx25aLoYObE53jhOxkCZoauUYhcOddepqS5Q7aP0vrXa4YAx8lcIJbQXlQzb9tMyCWw4yT7_KDsWDW_PNdV0ytnyHoIIUbaN_W7jFkO4riWReUmRzMFwgFgFsLPPtm8gO6Ui_pJDjq_WO61PhSFJ3QwGeXgJtXhzbETyMuRFDaDCk3gsAvg7qZBkKA3I7NsoP9A2fZpi93_yhdL-vV_XoVqMaeC7SxCVOswc16NiqWAo-jRvnmXIT7OnFcu027kw1euj7gcZo0IBf7-1aryk67FwE0SdJQ4ur_iEhCt6IJk5cJxG2LFKXiTpiPS8G8eVXlsEFiChNVi6n_IlGynTqeAJpvQqdVLnprBb41ExNyssLU"
	DoJwtTest(t, "/jwtrs256", 200, futureExpiry)
}

func TestJwtRS256NoExpiryPass(t *testing.T) {
	futureExpiry := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJzdWIiOiJhZG1pbiIsImp0aSI6ImI2NzU4MzM5LTdjOWYtNDI5NC04ZmQ3LTljNjM3NGQ3NGQyYSIsImlhdCI6MTYwODkzNjUyMH0.ZtPwFlZkFC-k_andUk0ZMJFP58Kn0BdbkbTIt2CbX9t1Oub2NTtQjbBceJQASwgXFmohGg_ob6g29cnkKjtWSdZa66rkG17VMA0XiFzG-iONXvtHaQHpza1TQfIc_I678CFDr-_0JUvHhQMuKXUVK4Kj--T2Jp4VdAP6a4Lei4rJxHxVeUt5NFIg5aUzxFTyxjD9m9dECDnY6kdrXbGDugssV8E7Hz9t4P8n2DKGiTmTobrfVbYMdnqMpb9sXIJ4M-_WgJVj0ew1wG0vVskqxVS6cNeXihacRo5jopGT9-ZiYOLjIUZa8bje26ZZR83fNmjja1FM7G1m8ODxSpnu1S0Xx_MUPUyRfiYqd1uGNPVEw76FTWO_RzKTjtjbfkySBOIGrYPrm73d4kXyNJDHEeNavSUi1suYT6ZeeIyZNYG_PlrRi2KgExPvR5vZH-b9f3rWanoarVGTlrAdC4mvZxErvaXGkPNsZkPsFbrZ2a8geJ1vx0Re4ESqWEiUlO92kGG9Elty7R8aF1HqGItBh-XRg8Krmx70lNQeC99doJhEHzXYShc8dA7k86lBUFNzMz90dudbw2viuTOFZ6YOn0I74rz1gD8xjsommPcraXTMNb08PlvWhquOpo8ILfJrQXDLOQE4hJNxSNXefXqdSKfdyjj95e90R5Fyi8N2o3c"
	DoJwtTest(t, "/jwtrs256", 200, futureExpiry)
}

func TestJwtRS256ExpiredFails(t *testing.T) {
	futureExpiry := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJzdWIiOiJhZG1pbiIsImp0aSI6ImQ3N2FiNGZiLThmN2UtNGE3Zi04ZjgwLWM1ZGE0YjJjMjNhNCIsImlhdCI6MTYwODkzNjYwMywiZXhwIjoxNDc1ODc4MzU3fQ.paHKeUIvPjJNqLZU-o9TWhTcctBQ8_tcMg5orz8cbANKVUTdpt4vyiVnHiseZ0tmZGzoekjwhUMHCl0ejN4LOSx_0A6FSw-OYqIljDu08_mGUW23_WwelUSHPUBbhCw64TLvQ_cz18N0MwwMqn3c0C0_6o476TEYsP6CgpwnvjFSJZXUU6xx8tpUJNfZEuCSBOxKEKpWiMLpZqvDxjIdtqW2Ia01Fgz3zfH-cz_rNauRuOSRVcFOMXHNmvBxezEIwBTyUdM-U7dgmcOV485cQPxvDSkpBHIqE4baHGAKKFD8phpU5SvFOmBF6JSp1gnenZFae6fkrpvvcPgqG4aK2l137j6_Nm_7LS4OPMlqe9KzftCZjTrwGQGOKd0SbGEjYNBl9AM3lIwIBinGsI-ScM-O4Qf7AxcLw6SQixyJM3yxuKTKr5SuZbdjy1k6s110EtHdLzzpEzj3xBbQKWztfD48mu1lNfaElzBaF4ltfdyJm0JhspQobIQIMjxpjHRoFRctF6TnS4d8TKz9rv3VXq29gy9wBvUYDDOJlZ_Lta6PZY2q-n8h2YARfw1TdFBJXmch3QfahPObg7ajF2t1C6aRNwRRzoCi7WtumasSRiyuliNS8JCauJkMdmHVbTt_tUcNVa6Nu3ps-ISTYCw1xHgUkpwlnKkX2QTgRJWd1Rg"
	DoJwtTest(t, "/jwtrs256", 401, futureExpiry)
}

func TestJwtRS256FutureIatFails(t *testing.T) {
	futureIat := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJzdWIiOiJhZG1pbiIsImp0aSI6ImEwYjliOTIzLWI5YWYtNDE2Ny05NDU1LTk3NjdlM2RlZDA2MSIsImlhdCI6MTQ3NTg3ODM1NzAwMH0.OCeUvlEW54sCtdd0pXAIpaBgEMvpEKMDhdoCXrXWkEUxykJ7E4yfBB1Dixb-a0kRwbSez6X571xOlrMOH92bS1_nuY4qsCTqbknQjA8m6CNzOjZ9HS0xQaBRrO_Ku4Hn2jIV0eGTJPpt_XuK6dn35xh6u-6CJd_sx-D3TAPeGA05353pk5BJYYq84mn-aKXMQgFJtuOnfi2YhQGYXWb5OEXFnwaM_ZSaCsn3iWl2UQxRQHpHi8jOc7trrO-bw6FBdkrPiuUqyUvLoudr8tVG4HJYxdYtTwKgwDYZikKhL7eyA7m7OEUfP-AkGWyyyxDKBFofH6BPdPLPovCoKV71WIy07bVzjfaF1XGfCE-LbJ0V6jOWjSUY6CvV-672NqXYkAPfJn2gxQla_ZUJj1QIiQP67OvZP9BInzRcljuCkZ9rtj9p1NyPdx09Isvt89EqI_lxkToHf3n0Mj6GOTd0Jey2L8CeVinNAylbIY4P-72xnKCUTWOYIjfGHg9HpQqsJNtPGLwcobqVI2hSdFUDCv6idUljW3D5TO9znB5ko2vrcrwMIKOL82G-RqjKl9iv448llQzW_wsqy30SiCFr7pxDEcgkdq0pC0qKm778pFOyXjq0bq4rN5Mw0q7pNyehiia6X8nHQ6o88_r9F6nsZN33WpxAosSFgnwvflb1ODs"
	DoJwtTest(t, "/jwtrs256", 401, futureIat)
}

func TestJwtRS256PastIatPass(t *testing.T) {
	pastIat := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJzdWIiOiJhZG1pbiIsImp0aSI6IjVhNWVhMmJiLTk2M2UtNDRkOC1iNmNmLTQ5NjhiZDllOTIyMiIsImlhdCI6MTYwODkzNTMyM30.jdHwTshGmR-0TaF-4-NKKdreuyPYFO7UkUn9KubY3uXPEdC1ukhussRzgl9xsTHGttp5jExYbyBTcnjVgxe2hAwArvZyC6CZvrsGFskDpmrqfe1MG_L7_J4XsMo0Lr11P6dGcYj25b8mKWN7Ka6N0DCyfbK0XXNQsG0Z2gkNZ3_5i-XKFloi8Biy69Ggou4GaZTt_E5Umtn5AByJYjPIAe827B35aHHWWEH1mIOnVqOXxMRL5CtmoT2iqv2eRpc_VLzCO-2s6ecuGyEz97iMob9jgmY3OP5gKYobpt8WE9-34c8BoJ-qPg3a_zfX8m6BJNdMhaM_joxR6wIz5VFNfuEMAWV-kv_rkRPzbYkmI6Owj2phT0n72C5Sz_buR3yxtMdlLi8vJ_jWcnLF8ppDZ9y0CALYY9XUljLGA0Kw8jLpmnQSgOQMXUkReelOu4hRvbfz_TESPBAqYDTvLDxtRNN1eDcD-fX5oItEElp5IAzfu5nZ_FiPQ2-mPXeUdYuMUIHxRWffztjce-xOqLiZuPJAhXT4MolwXsiF8JI7jBxrdhqL5Xs4Ap0Ql_T7jaqVMAqx-AXS-KlHEiQEtHrZijHBIbE8lXVBrv4bAKRvyjGyL2t7-wZ9r7jV6_2LgCPULVSj-n02-0SRIJB0KlAtw8h0pA2p1oG2rFcG39RAH6g"
	DoJwtTest(t, "/jwtrs256", 200, pastIat)
}

func TestJwtRS256FutureNbfFails(t *testing.T) {
	futureNbf := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJzdWIiOiJhZG1pbiIsImp0aSI6IjJhZmY5NDg5LWNlYTctNGIzNC1iYTNkLTNhYTg3YjEwNzY4MyIsImlhdCI6MTYwODkzNTk3MCwibmJmIjo5OTc1ODc4MzU3fQ.p0L-0fs_SLp607POmXIRqirLRgs_OqdKvDCN89udFcJMCNwQ9forCqYQcwJfufhaYy6A_gT8Kb24q2f9kw0v94gYc4Y8aAbFJdNldEU-IiZ3peYs9YolZefCy5JFujPxt4Vv6YKDV_njWYCQIzOzxa2VQQ0kVQf6vO-ffjbPQo6VwpCkJHmJ1B6PFH-eRRY-_mAivgAL3KGCeAmhanrl1m-UH6xt-5xINw8Y0o_4fQki--J8YZCmeDNzKMukRIv3qwQqDfabukGR3m-vBo_8_9rXzDhE4TNk_msx8RYlTEpzIRFmBlG_vy1ZuLZxLWseD2d053U8477WtAjENFE0YJNe_ojNXpPq4w5o1nb5ibdnPG5vTsbsVg8UB6h4v8QdsRZyn_9-D4nSc-Q0_qOajBElvLDt4qM6hSo3RWZlBDcvZjDL2Wg9Yue--q1WPCBEmemXBy1CQzmdsIWbJENVIx1EtuWgXJEOvOdw0-oXoxkpHBoxbG45QjxaQvRTSO7ZR6wOAZrNlhR7IltOSRyxY8aspe0-aWSQMFJnNYGUg5bay5SVhh_bGTWhtDAJYfCDhtbC63qbZKhxrJtoIajlBJfBucqikJ_24e9llrrKFKHCjDTXM90Cii3c37t3VPf2yNdnhQFVw0w4S2RdcD7dvB7ciPLcSSQz7LWdNxmXw20"
	DoJwtTest(t, "/jwtrs256", 401, futureNbf)
}

func TestJwtRS256PastNbfPass(t *testing.T) {
	pastNbf := "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJpc3N1ZXIiLCJzdWIiOiJhZG1pbiIsImp0aSI6ImYzYjc2MThjLThjOWYtNGY0Yi05YTJjLTEwZThjNjY0NGFkMSIsImlhdCI6MTYwODkzNjEyMiwibmJmIjoxNDc1ODc4MzU3fQ.pXq8gd9qqQJrwGn5lYIohVXEomgBr1eDLq92OpJsIspCUlGLIh0k-Z0YjBspo488At6QPqyWIborLT5aOnbi5hZkitGeD1AV4q-u3tCSn-Kgll1Wfa4s3cnpeN9SzYCsnIaEGi4AQOaiEqURUCb8XSdymSHdOALaNA5r2D6dzMjfA68Apzed6PdeoTghfcq5EL7JoKI41Ha-tMOaJiJGRud1bA5YyThhXPLcq1ze6F8FYSZ_Wnmu4P2oUGimrKBr8ix9_yXA_Q49_KSfYwmdPH7BHQcKCMX9yxtZbsjD3geiPo4HMsu4l-AZW04ol9jwyvRYn7h8CflkuBcfsGjF_1uuhGjQbaMiyvWVuVVKiKn6JM7b0mptV0GT1vAZFKTxq7wcxKSd2V4W-T1fmrH_Z_YucA51p3e0SIqIVMbUmS-59cy22lXP_zsTdbl8a3hhO4rXjX6ZB4xUGvpEUxuh7aeeJ5snv64xhqboGNDC_VpBSlWy8rrpnPKzQVVESGtI6L6t8QcmfXwSZcwv5aEmGg1-W0omY0t0We7oqUguswasDTXNFXMdmlb2nw6bfzSKfSPh1xLZUfXhaOfkiczu79snwW-3r0w9aR8UmWuNZ9n5unJlrr9LcHgbugIG4xZFVMmOGF2CN0SAzSV6q7gRWgzwuUvq6v7nmR3oLvDMV9E"
	DoJwtTest(t, "/jwtrs256", 200, pastNbf)
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
