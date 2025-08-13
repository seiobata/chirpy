package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hashedPass, err := HashPassword(password)
	hashedPass2, _ := HashPassword(password)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if hashedPass == "" {
		t.Fatal("Expected non-empty hash")
	}

	if hashedPass == password {
		t.Fatal("Hash should not equal original password")
	}

	if hashedPass == hashedPass2 {
		t.Fatal("Expected different hashes for the same password")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "anotherpassword"
	hashedPass, err := HashPassword(password)

	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	err = CheckPasswordHash(password, hashedPass)
	if err != nil {
		t.Fatalf("Passwords are not equal: %v", err)
	}

	err = CheckPasswordHash("wrongpassword", hashedPass)
	if err == nil {
		t.Fatalf("Passwords should not be equal: %v", err)
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "secret"

	token, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	validatedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed: %v", err)
	}

	if validatedID != userID {
		t.Fatalf("Expected userID %v, got %v", userID, validatedID)
	}

	_, err = ValidateJWT(token, "not secret")
	if err == nil {
		t.Fatal("Expected error for wrong secret, got nil")
	}

	_, err = ValidateJWT("", secret)
	if err == nil {
		t.Fatal("Expected error for missing token, got nil")
	}

	// check for token expiration
	token, err = MakeJWT(userID, secret, time.Millisecond)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	time.Sleep(2 * time.Millisecond)
	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}
}
