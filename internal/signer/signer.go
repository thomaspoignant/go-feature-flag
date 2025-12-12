package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Sign is using the data and the secret to compute a HMAC(SHA256) to sign the body of the request.
// so the webhook can use this signature to verify that no data have been compromised.
func Sign(payloadBody, secretToken []byte) string {
	mac := hmac.New(sha256.New, secretToken)
	_, _ = mac.Write(payloadBody)
	expectedMAC := mac.Sum(nil)
	return "sha256=" + hex.EncodeToString(expectedMAC)
}
