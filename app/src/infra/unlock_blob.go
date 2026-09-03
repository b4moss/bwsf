package infra

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bnema/bitwarden-go-sdk/bitwarden"
)

const unlockBlobVersion = 1

// unlockBlobV1 is the versioned opaque encoding of SessionMaterial (no master password).
type unlockBlobV1 struct {
	Version      int    `json:"v"`
	AccountID    string `json:"account_id"`
	UserKeyB64   string `json:"user_key"`
	AccessB64    string `json:"access_token,omitempty"`
	RefreshB64   string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	RememberB64  string `json:"remembered_2fa,omitempty"`
}

func encodeUnlockBlob(material bitwarden.SessionMaterial) (string, error) {
	if material.AccountID == "" || len(material.UserKey) == 0 {
		return "", fmt.Errorf("incomplete session material")
	}
	blob := unlockBlobV1{
		Version:    unlockBlobVersion,
		AccountID:  material.AccountID,
		UserKeyB64: base64.StdEncoding.EncodeToString(material.UserKey),
		TokenType:  material.Tokens.TokenType,
	}
	if len(material.Tokens.AccessToken) > 0 {
		blob.AccessB64 = base64.StdEncoding.EncodeToString(material.Tokens.AccessToken)
	}
	if len(material.Tokens.RefreshToken) > 0 {
		blob.RefreshB64 = base64.StdEncoding.EncodeToString(material.Tokens.RefreshToken)
	}
	if len(material.Tokens.RememberedTwoFactorToken) > 0 {
		blob.RememberB64 = base64.StdEncoding.EncodeToString(material.Tokens.RememberedTwoFactorToken)
	}
	if !material.Tokens.ExpiresAt.IsZero() {
		blob.ExpiresAt = material.Tokens.ExpiresAt.UTC().Format(time.RFC3339Nano)
	}
	raw, err := json.Marshal(blob)
	if err != nil {
		return "", fmt.Errorf("encode unlock blob: %w", err)
	}
	return "v1:" + base64.StdEncoding.EncodeToString(raw), nil
}

func decodeUnlockBlob(blob string) (bitwarden.SessionMaterial, error) {
	blob = strings.TrimSpace(blob)
	if blob == "" {
		return bitwarden.SessionMaterial{}, fmt.Errorf("empty unlock blob")
	}
	payload := blob
	if strings.HasPrefix(blob, "v1:") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(blob, "v1:"))
		if err != nil {
			return bitwarden.SessionMaterial{}, fmt.Errorf("invalid unlock blob encoding: %w", err)
		}
		payload = string(decoded)
	} else if strings.HasPrefix(blob, "{") {
		// raw JSON without prefix (still must declare v)
	} else {
		return bitwarden.SessionMaterial{}, fmt.Errorf("unknown unlock blob version")
	}

	var parsed unlockBlobV1
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return bitwarden.SessionMaterial{}, fmt.Errorf("invalid unlock blob: %w", err)
	}
	if parsed.Version != unlockBlobVersion {
		return bitwarden.SessionMaterial{}, fmt.Errorf("unsupported unlock blob version %d", parsed.Version)
	}
	if parsed.AccountID == "" || parsed.UserKeyB64 == "" {
		return bitwarden.SessionMaterial{}, fmt.Errorf("incomplete unlock blob")
	}
	userKey, err := base64.StdEncoding.DecodeString(parsed.UserKeyB64)
	if err != nil || len(userKey) == 0 {
		return bitwarden.SessionMaterial{}, fmt.Errorf("invalid unlock blob user key")
	}
	material := bitwarden.SessionMaterial{
		AccountID: parsed.AccountID,
		UserKey:   userKey,
		Tokens: bitwarden.TokenSet{
			AccountID: parsed.AccountID,
			TokenType: parsed.TokenType,
		},
	}
	if parsed.AccessB64 != "" {
		b, err := base64.StdEncoding.DecodeString(parsed.AccessB64)
		if err != nil {
			return bitwarden.SessionMaterial{}, fmt.Errorf("invalid unlock blob access token")
		}
		material.Tokens.AccessToken = b
	}
	if parsed.RefreshB64 != "" {
		b, err := base64.StdEncoding.DecodeString(parsed.RefreshB64)
		if err != nil {
			return bitwarden.SessionMaterial{}, fmt.Errorf("invalid unlock blob refresh token")
		}
		material.Tokens.RefreshToken = b
	}
	if parsed.RememberB64 != "" {
		b, err := base64.StdEncoding.DecodeString(parsed.RememberB64)
		if err != nil {
			return bitwarden.SessionMaterial{}, fmt.Errorf("invalid unlock blob 2fa token")
		}
		material.Tokens.RememberedTwoFactorToken = b
	}
	if parsed.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339Nano, parsed.ExpiresAt)
		if err != nil {
			t, err = time.Parse(time.RFC3339, parsed.ExpiresAt)
			if err != nil {
				return bitwarden.SessionMaterial{}, fmt.Errorf("invalid unlock blob expiry")
			}
		}
		material.Tokens.ExpiresAt = t
	}
	return material, nil
}
