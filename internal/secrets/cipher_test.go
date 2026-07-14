package secrets

import (
	"bytes"
	"testing"
)

func TestCipherRoundTrip(t *testing.T) {
	cipher, err := NewCipher(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		t.Fatalf("NewCipher returned error: %v", err)
	}

	encrypted, err := cipher.Encrypt("gantry-control-api-key", []byte("secret-token"))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if bytes.Contains(encrypted, []byte("secret-token")) {
		t.Fatal("ciphertext contains plaintext")
	}

	decrypted, err := cipher.Decrypt("gantry-control-api-key", encrypted)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}
	if string(decrypted) != "secret-token" {
		t.Fatalf("plaintext = %q", decrypted)
	}
}

func TestCipherBindsCiphertextToSecretName(t *testing.T) {
	cipher, err := NewCipher(bytes.Repeat([]byte{2}, 32))
	if err != nil {
		t.Fatalf("NewCipher returned error: %v", err)
	}
	encrypted, err := cipher.Encrypt("first", []byte("secret-token"))
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if _, err := cipher.Decrypt("second", encrypted); err == nil {
		t.Fatal("Decrypt returned nil error for a different secret name")
	}
}
