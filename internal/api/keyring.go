package api

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const keyringService = "salja"

// KeyringTokenStore encrypts tokens at rest using the OS keyring.
type KeyringTokenStore struct {
	// Fallback stores tokens on disk when the OS keyring is unavailable.
	Fallback *TokenStore
}

// NewKeyringTokenStore creates a token store backed by the OS keyring.
func NewKeyringTokenStore(fallback *TokenStore) *KeyringTokenStore {
	return &KeyringTokenStore{Fallback: fallback}
}

func (k *KeyringTokenStore) Get(service string) (*Token, error) {
	data, err := keyring.Get(keyringService, service)
	if err != nil {
		// Fall back to file-based store and migrate if found
		if k.Fallback != nil {
			token, fbErr := k.Fallback.Get(service)
			if fbErr != nil {
				return nil, fbErr
			}
			// Migrate to keyring for future use
			_ = k.Set(service, token)
			return token, nil
		}
		return nil, fmt.Errorf("no token for service %q — run: salja auth login %s", service, service)
	}

	var token Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("failed to decode token from keyring for %q: %w", service, err)
	}
	return &token, nil
}

func (k *KeyringTokenStore) Set(service string, token *Token) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	err = keyring.Set(keyringService, service, string(data))
	if err != nil {
		// Fall back to file-based store
		if k.Fallback != nil {
			return k.Fallback.Set(service, token)
		}
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}
	return nil
}

func (k *KeyringTokenStore) Delete(service string) error {
	err := keyring.Delete(keyringService, service)
	if err != nil && k.Fallback != nil {
		return k.Fallback.Delete(service)
	}
	return err
}

// IsAvailable returns true if the OS keyring is accessible.
func KeyringAvailable() bool {
	err := keyring.Set(keyringService, "__probe__", "test")
	if err != nil {
		return false
	}
	_ = keyring.Delete(keyringService, "__probe__")
	return true
}
