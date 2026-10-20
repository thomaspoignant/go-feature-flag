package githubretriever

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// signAppJWT creates a minimal RS256 JWT for GitHub App authentication using
// stdlib only (no external JWT library). The token is valid for 9 minutes,
// which is within GitHub's 10-minute maximum.
//
//   - iss: appID
//   - iat: now − 60 s  (clock-skew tolerance)
//   - exp: now + 9 min
func signAppJWT(appID int64, key *rsa.PrivateKey) (string, error) {
	now := time.Now()

	headerJSON, err := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	if err != nil {
		return "", fmt.Errorf("github app jwt: failed to marshal header: %w", err)
	}
	claimsJSON, err := json.Marshal(map[string]int64{
		"iss": appID,
		"iat": now.Add(-60 * time.Second).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
	})
	if err != nil {
		return "", fmt.Errorf("github app jwt: failed to marshal claims: %w", err)
	}

	signingInput := jwtBase64(headerJSON) + "." + jwtBase64(claimsJSON)

	digest := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		return "", fmt.Errorf("github app jwt: failed to sign: %w", err)
	}

	return signingInput + "." + jwtBase64(sig), nil
}

// jwtBase64 is base64url encoding without padding, as required by RFC 7515.
func jwtBase64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
