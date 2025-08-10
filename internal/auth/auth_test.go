package auth

import "testing"

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
