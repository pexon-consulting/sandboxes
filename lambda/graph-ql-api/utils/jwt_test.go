package utils_test

import (
	"context"
	"lambda/aws-sandbox/graph-ql-api/utils"
	"testing"
)

const tokenString string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJlbWFpbCI6Im1heEBtYXguZGUifQ.G9RcncmreaHa7XPoJZexiKHUBoBeVnjw3x3v9z4LPIU"

//const tokenString string = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImQwMWMxYWJlMjQ5MjY5ZjcyZWY3Y2EyNjEzYTg2YzlmMDVlNTk1NjciLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJuYmYiOjE2NDEzMDA4MDMsImF1ZCI6Ijk5ODQ4OTg5MzM1OC1zZ2Y3YzZxNGI3NHRoZ3RwNzRwaG44a3Y1ZGtyYWFnMy5hcHBzLmdvb2dsZXVzZXJjb250ZW50LmNvbSIsInN1YiI6IjExNzY1ODcyMDk1NDgwMTcyMzEyOCIsImhkIjoicGV4b24tY29uc3VsdGluZy5kZSIsImVtYWlsIjoibWFyaXVzLmhhYmVyc3RvY2tAcGV4b24tY29uc3VsdGluZy5kZSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJhenAiOiI5OTg0ODk4OTMzNTgtc2dmN2M2cTRiNzR0aGd0cDc0cGhuOGt2NWRrcmFhZzMuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJuYW1lIjoiTWFyaXVzIEhhYmVyc3RvY2siLCJnaXZlbl9uYW1lIjoiTWFyaXVzIiwiZmFtaWx5X25hbWUiOiJIYWJlcnN0b2NrIiwiaWF0IjoxNjQxMzAxMTAzLCJleHAiOjE2NDEzMDQ3MDMsImp0aSI6IjY3OThmNjZiYWE5OWUzNzAwMGJjMTgwMjY2M2M5MTZhOWExMTRlZWQifQ.X1CSTVvoQcmkjxSYokVAkb8IgKVlhywFw5BIAV1xySdeCyWGRIWpj9HEZm3ChckMbH2e9twMX4t0RMJcuyaNtvX7deJ6m2aq4ggUO4vcfZvZxDFsG2jERwOzy1Mx4tasUsAjJtLfs_V5rrCDqHl7Rqdx3HkEDG_abvrkymk-G8ToCSsQNev_S4i_FoWvJTY_weHSn4NVd7LfSJfxnQiTvp3GJgS830dOjrXZ-aNI4a8XO9FcnQyEm3dQvcXL8KIdtYptoRy3QaJaC4-oeOM1gYw3xUm506nXXXUEigrHeI6s0yLTVpfbs64hPgzgNaARm0-fyP_9tjpqFJaPvA-Rfg"

func TestParseJWT(t *testing.T) {
	_, err := utils.ParseJWT(context.TODO(), tokenString)
	if err != nil {
		t.Fatal("ParseJWT should not return an error")
	}

}

func TestParseJWTCorruptJwt(t *testing.T) {
	tokenString := "12331531431241341"
	_, err := utils.ParseJWT(context.TODO(), tokenString)

	if err == nil {
		t.Fatal("test should fails")
	}

	if err.Error() != "token contains an invalid number of segments" {
		t.Fatal("token should contains an invalid number of segments")
	}

}
