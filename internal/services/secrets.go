package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"capcom/internal/domain"
)

var secretNamePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._-]{0,127}$`)

var ErrSecretNotFound = errors.New("secret not found")

type SecretRepository interface {
	Create(ctx context.Context, secret domain.Secret, ciphertext []byte) (domain.Secret, error)
	Rotate(ctx context.Context, name string, ciphertext []byte) (domain.Secret, error)
	GetCiphertext(ctx context.Context, name string) ([]byte, error)
}

type SecretCipher interface {
	Encrypt(name string, plaintext []byte) ([]byte, error)
	Decrypt(name string, payload []byte) ([]byte, error)
}

type SecretService struct {
	repository SecretRepository
	audit      AuditRepository
	cipher     SecretCipher
}

type StoreSecretInput struct {
	Name   string
	Value  string
	Actor  string
	Reason string
}

func NewSecretService(repository SecretRepository, audit AuditRepository, cipher SecretCipher) SecretService {
	return SecretService{repository: repository, audit: audit, cipher: cipher}
}

func (s SecretService) Create(ctx context.Context, input StoreSecretInput) (domain.Secret, error) {
	input = normalizeStoreSecretInput(input)
	if err := validateStoreSecretInput(input); err != nil {
		return domain.Secret{}, err
	}
	ciphertext, err := s.cipher.Encrypt(input.Name, []byte(input.Value))
	if err != nil {
		return domain.Secret{}, fmt.Errorf("encrypt secret: %w", err)
	}
	secret, err := s.repository.Create(ctx, domain.Secret{Name: input.Name}, ciphertext)
	if err != nil {
		return domain.Secret{}, err
	}
	if err := s.auditMutation(ctx, secret, input.Actor, input.Reason, "secret.created"); err != nil {
		return domain.Secret{}, err
	}
	return secret, nil
}

func (s SecretService) Rotate(ctx context.Context, input StoreSecretInput) (domain.Secret, error) {
	input = normalizeStoreSecretInput(input)
	if err := validateStoreSecretInput(input); err != nil {
		return domain.Secret{}, err
	}
	ciphertext, err := s.cipher.Encrypt(input.Name, []byte(input.Value))
	if err != nil {
		return domain.Secret{}, fmt.Errorf("encrypt secret: %w", err)
	}
	secret, err := s.repository.Rotate(ctx, input.Name, ciphertext)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Secret{}, ErrSecretNotFound
		}
		return domain.Secret{}, err
	}
	if err := s.auditMutation(ctx, secret, input.Actor, input.Reason, "secret.rotated"); err != nil {
		return domain.Secret{}, err
	}
	return secret, nil
}

func (s SecretService) Resolve(ctx context.Context, ref string) (string, error) {
	name := strings.TrimSpace(ref)
	if !secretNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid secret reference")
	}
	ciphertext, err := s.repository.GetCiphertext(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrSecretNotFound
		}
		return "", err
	}
	plaintext, err := s.cipher.Decrypt(name, ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (s SecretService) auditMutation(ctx context.Context, secret domain.Secret, actor, reason, eventType string) error {
	if s.audit == nil {
		return fmt.Errorf("audit repository is required")
	}
	_, err := s.audit.Create(ctx, domain.AuditEvent{
		Actor:      actor,
		EventType:  eventType,
		TargetType: "secret",
		TargetID:   secret.Name,
		Reason:     reason,
		Result:     "succeeded",
		After: map[string]any{
			"name":       secret.Name,
			"updated_at": secret.UpdatedAt,
		},
	})
	if err != nil {
		return fmt.Errorf("audit %s: %w", eventType, err)
	}
	return nil
}

func normalizeStoreSecretInput(input StoreSecretInput) StoreSecretInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Actor = strings.TrimSpace(input.Actor)
	input.Reason = strings.TrimSpace(input.Reason)
	return input
}

func validateStoreSecretInput(input StoreSecretInput) error {
	if !secretNamePattern.MatchString(input.Name) {
		return fmt.Errorf("name must start with a letter and contain only letters, numbers, dot, underscore, or hyphen")
	}
	if input.Value == "" {
		return fmt.Errorf("value is required")
	}
	if input.Actor == "" {
		return fmt.Errorf("actor is required")
	}
	if input.Reason == "" {
		return fmt.Errorf("reason is required")
	}
	return nil
}
