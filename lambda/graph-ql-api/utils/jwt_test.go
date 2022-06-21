package utils_test

import (
	"lambda/aws-sandbox/graph-ql-api/utils"
	"testing"
)

const tokenString string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJlbWFpbCI6Im1heEBtYXguZGUifQ.G9RcncmreaHa7XPoJZexiKHUBoBeVnjw3x3v9z4LPIU"

func TestParseJWT(t *testing.T) {
	_, err := utils.ParseJWT(tokenString)
	if err != nil {
		t.Fatal("ParseJWT should not return an error")
	}

}

func TestParseJWTCorruptJwt(t *testing.T) {
	tokenString := "12331531431241341"
	_, err := utils.ParseJWT(tokenString)

	if err == nil {
		t.Fatal("test should fails")
	}

	if err.Error() != "token contains an invalid number of segments" {
		t.Fatal("token should contains an invalid number of segments")
	}

}
