package token

import (
	"crypto/rand"
	"testing"

	"github.com/jcroyoaun/totalcompmx/internal/assert"
)

func TestNew(t *testing.T) {
	t.Run("Generate token with lowercase base32 encoding and 26 characters", func(t *testing.T) {
		for range 10000 {
			token := New()
			assert.True(t, token != "")
			assert.MatchesRegexp(t, token, `^[a-z2-7]{26}$`)
		}
	})

	t.Run("Generate unique tokens on multiple calls", func(t *testing.T) {
		token1 := New()
		token2 := New()
		token3 := New()

		assert.NotEqual(t, token1, token2)
		assert.NotEqual(t, token1, token3)
		assert.NotEqual(t, token2, token3)
	})
}

func TestHash(t *testing.T) {
	t.Run("Create a lowercase hex-encoded sha-256 hash of a string", func(t *testing.T) {
		for range 10000 {
			hash := Hash(rand.Text())
			assert.True(t, hash != "")
			assert.MatchesRegexp(t, hash, `^[0-9a-f]{64}$`)
		}
	})

	t.Run("Generate consistent hash for same input", func(t *testing.T) {
		plaintext := "consistent-input"

		hash1 := Hash(plaintext)
		hash2 := Hash(plaintext)
		hash3 := Hash(plaintext)

		assert.Equal(t, hash1, hash2)
		assert.Equal(t, hash1, hash3)
		assert.Equal(t, hash2, hash3)
	})

	t.Run("Generate different hashes for different inputs", func(t *testing.T) {
		hash1 := Hash("input1")
		hash2 := Hash("input2")
		hash3 := Hash("input3")

		assert.NotEqual(t, hash1, hash2)
		assert.NotEqual(t, hash1, hash3)
		assert.NotEqual(t, hash2, hash3)
	})
}
