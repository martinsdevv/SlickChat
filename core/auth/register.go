package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/martinsdevv/slickchat/core/contracts"
	"github.com/martinsdevv/slickchat/core/domain"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost           = 12
	maxDiscriminatorRetries = 10
	discriminatorDigits  = 4
)

type RegisterUseCase struct {
	users contracts.UserRepository
}

func NewRegisterUseCase(users contracts.UserRepository) *RegisterUseCase {
	return &RegisterUseCase{users: users}
}

type RegisterInput struct {
	Username string
	Password string
}

type RegisterResult struct {
	User           *domain.User
	RecoveryKeyRaw string
}

func (uc *RegisterUseCase) Execute(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	if err := validateUsername(input.Username); err != nil {
		return nil, err
	}

	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	recoveryKeyRaw, recoveryKeyHash, err := generateRecoveryKey()
	if err != nil {
		return nil, fmt.Errorf("generate recovery key: %w", err)
	}

	user := &domain.User{
		ID:              uuid.New(),
		Username:        strings.ToLower(input.Username),
		PasswordHash:    string(passwordHash),
		RecoveryKeyHash: recoveryKeyHash,
		ParanoidMode:    false,
		CreatedAt:       time.Now().UTC(),
	}

	for range maxDiscriminatorRetries {
		discriminator, err := randomDiscriminator()
		if err != nil {
			return nil, fmt.Errorf("generate discriminator: %w", err)
		}

		user.Discriminator = discriminator

		err = uc.users.Save(ctx, user)
		if err == nil {
			return &RegisterResult{
				User:           user,
				RecoveryKeyRaw: recoveryKeyRaw,
			}, nil
		}

		if !errors.Is(err, domain.ErrHandleTaken) {
			return nil, fmt.Errorf("save user: %w", err)
		}
	}

	return nil, fmt.Errorf("could not find available discriminator for username %q after %d attempts", input.Username, maxDiscriminatorRetries)
}

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 32 {
		return fmt.Errorf("%w: username must be between 3 and 32 characters", domain.ErrInvalidHandle)
	}

	for _, r := range username {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return fmt.Errorf("%w: username contains invalid character %q", domain.ErrInvalidHandle, r)
		}
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", domain.ErrInvalidCredentials)
	}

	return nil
}

func randomDiscriminator() (string, error) {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	n := (int(b[0])<<8 | int(b[1])) % 10000
	return fmt.Sprintf("%04d", n), nil
}

func generateRecoveryKey() (raw, hashed string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}

	raw = hex.EncodeToString(b)
	sum := sha256.Sum256([]byte(raw))
	hashed = hex.EncodeToString(sum[:])
	return raw, hashed, nil
}
