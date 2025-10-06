package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "tested_password 666 1234 $%^&*"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("error hashing password '%s': %s", password, err)
	}
	if strings.Contains(hash, password) {
		t.Errorf("hash contains password")
	}
	if !strings.Contains(hash, "argon2id$v=") {
		t.Errorf("hash doesn't have the expected format: %s", hash)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "tested_password 666 1234 $%^&*"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("error hashing password '%s': %s", password, err)
	}
	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("error checking password against hash: %s", err)
	}
	if !match {
		t.Errorf("password %s doesn't match hash %s", password, hash)
	}
}

func TestCheckPasswordHash_WeirdCharacters(t *testing.T) {
	password := "śżćśł自分おパスワード👍🍑🍆"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("error hashing password '%s': %s", password, err)
	}
	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("error checking password against hash: %s", err)
	}
	if !match {
		t.Errorf("password %s doesn't match hash %s", password, hash)
	}
}

func TestMakeJWT(t *testing.T) {
	code := "SecretCode"
	exp, err := time.ParseDuration("5s")
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	testJWT, err := MakeJWT(id, code, exp)
	if err != nil {
		t.Errorf("Error generating the jwt: %s", err)
	}
	if testJWT == "" {
		t.Errorf("Error: jwt came out empty")
	}
}

func TestValidateJWT(t *testing.T) {
	code := "SecretCode"
	exp, err := time.ParseDuration("5s")
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	testJWT, err := MakeJWT(id, code, exp)
	if err != nil {
		t.Errorf("Error generating the jwt: %s", err)
	}
	if testJWT == "" {
		t.Errorf("Error: jwt came out empty")
	}

	newid, err := ValidateJWT(testJWT, code)
	if err != nil {
		t.Errorf("Error validating jwt: %s", err)
	}
	if id != newid {
		t.Errorf("The uuids before and after validation don't match\nOriginal uuid: %v\nValidated uuid: %v", id.String(), newid.String())
	}
}

func TestJWTExpiration(t *testing.T) {
	code := "SecretCode"
	exp, err := time.ParseDuration("5s")
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}

	testJWT, err := MakeJWT(id, code, exp)
	if err != nil {
		t.Errorf("Error generating the jwt: %s", err)
	}
	if testJWT == "" {
		t.Errorf("Error: jwt came out empty")
	}

	newid, err1 := ValidateJWT(testJWT, code)
	if err1 != nil {
		t.Errorf("Error validating jwt: %s", err)
	}
	if id != newid {
		t.Errorf("The uuids before and after validation don't match\nOriginal uuid: %v\nValidated uuid: %v", id.String(), newid.String())
	}
	sleeptime, err := time.ParseDuration("6s")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(sleeptime)

	newid, err2 := ValidateJWT(testJWT, code)
	if err2 != nil && err2.Error() != "token has invalid claims: token is expired" {
		t.Errorf("Error revalidating jwt: %s", err)
	}
	if id == newid {
		t.Errorf("The new id matches the previous one despite the token expiring.")
	}
}
