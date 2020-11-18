package integration

import (
	"fmt"
	"net/http"
	"testing"
)

func TestJwtNoneNotExpired(t *testing.T) {
	DoJwtTest(t, "/jwtnone", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiNDNkODEzOTYtOWZlMi00ZTRkLTgxMDYtOWQyOTFlY2FmMjVkIiwiaWF0IjoxNjA1NTY2MDU0LCJleHAiOjk0NzU4NzgzNTd9.")
}

func TestJwtNoneNoExpiry(t *testing.T) {
	DoJwtTest(t, "/jwtnone", 200, "eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJpc3MiOiJqb2UiLCJodHRwOi8vZXhhbXBsZS5jb20vaXNfcm9vdCI6dHJ1ZSwianRpIjoiZTI4ZGY3MTItMGNmNy00ZTVlLWE3NjUtYjE0MGQwM2IwNDRjIiwiaWF0IjoxNjA1NjY3NDEzLCJleHAiOjE2MDU2NzEwMTN9.")
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
