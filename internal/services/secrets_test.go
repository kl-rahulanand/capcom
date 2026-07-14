package services

import (
	"context"
	"database/sql"
	"testing"

	"capcom/internal/domain"
)

func TestSecretServiceCreateResolveAndRotate(t *testing.T) {
	repository := &fakeSecretRepository{}
	service := NewSecretService(repository, fakeAuditRepo{}, fakeSecretCipher{})

	created, err := service.Create(context.Background(), StoreSecretInput{
		Name: "gantry-control-api-key", Value: "first", Actor: "tester", Reason: "setup",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.Name != "gantry-control-api-key" {
		t.Fatalf("name = %q", created.Name)
	}
	resolved, err := service.Resolve(context.Background(), created.Name)
	if err != nil || resolved != "first" {
		t.Fatalf("Resolve = %q, %v", resolved, err)
	}

	if _, err := service.Rotate(context.Background(), StoreSecretInput{
		Name: created.Name, Value: "second", Actor: "tester", Reason: "rotation",
	}); err != nil {
		t.Fatalf("Rotate returned error: %v", err)
	}
	resolved, err = service.Resolve(context.Background(), created.Name)
	if err != nil || resolved != "second" {
		t.Fatalf("Resolve after rotate = %q, %v", resolved, err)
	}
}

func TestSecretServiceRejectsMissingAuditContext(t *testing.T) {
	service := NewSecretService(&fakeSecretRepository{}, fakeAuditRepo{}, fakeSecretCipher{})
	_, err := service.Create(context.Background(), StoreSecretInput{Name: "valid-name", Value: "value"})
	if err == nil {
		t.Fatal("Create returned nil error")
	}
}

type fakeSecretRepository struct {
	secret     domain.Secret
	ciphertext []byte
}

func (r *fakeSecretRepository) Create(_ context.Context, secret domain.Secret, ciphertext []byte) (domain.Secret, error) {
	secret.ID = "secret-1"
	r.secret = secret
	r.ciphertext = append([]byte(nil), ciphertext...)
	return secret, nil
}

func (r *fakeSecretRepository) Rotate(_ context.Context, name string, ciphertext []byte) (domain.Secret, error) {
	if r.secret.Name != name {
		return domain.Secret{}, sql.ErrNoRows
	}
	r.ciphertext = append([]byte(nil), ciphertext...)
	return r.secret, nil
}

func (r *fakeSecretRepository) GetCiphertext(_ context.Context, name string) ([]byte, error) {
	if r.secret.Name != name {
		return nil, sql.ErrNoRows
	}
	return append([]byte(nil), r.ciphertext...), nil
}

type fakeSecretCipher struct{}

func (fakeSecretCipher) Encrypt(_ string, plaintext []byte) ([]byte, error) {
	return append([]byte("encrypted:"), plaintext...), nil
}

func (fakeSecretCipher) Decrypt(_ string, payload []byte) ([]byte, error) {
	return append([]byte(nil), payload[len("encrypted:"):]...), nil
}
