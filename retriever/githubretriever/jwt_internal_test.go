package githubretriever

import (
	"crypto/rsa"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignAppJWT_SmallKeyFails covers the rsa.SignPKCS1v15 error branch.
// A synthetic key with a 17-bit modulus is constructed directly (bypassing
// rsa.GenerateKey's minimum-size check) and is too small for SHA-256 PKCS#1 v1.5.
func TestSignAppJWT_SmallKeyFails(t *testing.T) {
	tinyKey := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{
			N: big.NewInt(65537 * 65537), // ~32 bits — far below the 1024-bit RSA minimum
			E: 65537,
		},
	}

	_, err := signAppJWT(12345, tinyKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to sign")
}
