package database

import (
	"testing"
	"time"

	"github.com/jcroyoaun/totalcompmx/internal/assert"
)

func TestInsertPasswordReset(t *testing.T) {
	t.Run("Successfully inserts password reset", func(t *testing.T) {
		db := newTestDB(t)

		hashedToken := "hashed_token_123"
		userID := testUsers["alice"].id
		ttl := 24 * time.Hour

		err := db.InsertPasswordReset(hashedToken, userID, ttl)
		assert.Nil(t, err)

		var passwordResets []PasswordReset
		err = db.Select(&passwordResets, "SELECT * FROM password_resets WHERE user_id = $1", testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, len(passwordResets), 1)
		assert.Equal(t, passwordResets[0].HashedToken, hashedToken)
		assert.Equal(t, passwordResets[0].UserID, userID)
		assert.True(t, passwordResets[0].Expiry.After(time.Now()))
	})

	t.Run("Fails with invalid user ID", func(t *testing.T) {
		db := newTestDB(t)

		hashedToken := "hashed_token_123"
		userID := 99999
		ttl := 24 * time.Hour

		err := db.InsertPasswordReset(hashedToken, userID, ttl)
		assert.NotNil(t, err)
	})
}

func TestGetPasswordReset(t *testing.T) {
	t.Run("Returns password reset when token exists and not expired", func(t *testing.T) {
		db := newTestDB(t)

		hashedToken := "hashed_token_123"
		userID := testUsers["alice"].id
		ttl := 24 * time.Hour

		err := db.InsertPasswordReset(hashedToken, userID, ttl)
		if err != nil {
			t.Fatal(err)
		}

		passwordReset, found, err := db.GetPasswordReset(hashedToken)
		assert.Nil(t, err)
		assert.True(t, found)
		assert.Equal(t, passwordReset.HashedToken, hashedToken)
		assert.Equal(t, passwordReset.UserID, userID)
		assert.True(t, passwordReset.Expiry.After(time.Now()))
	})

	t.Run("Returns not found when token does not exist", func(t *testing.T) {
		db := newTestDB(t)

		passwordReset, found, err := db.GetPasswordReset("nonexistent_token")
		assert.Nil(t, err)
		assert.False(t, found)
		assert.Equal(t, passwordReset, PasswordReset{})
	})

	t.Run("Returns not found when token is expired", func(t *testing.T) {
		db := newTestDB(t)

		hashedToken := "hashed_token_123"
		userID := testUsers["alice"].id
		ttl := -1 * time.Hour

		err := db.InsertPasswordReset(hashedToken, userID, ttl)
		if err != nil {
			t.Fatal(err)
		}

		passwordReset, found, err := db.GetPasswordReset(hashedToken)
		assert.Nil(t, err)
		assert.False(t, found)
		assert.Equal(t, passwordReset, PasswordReset{})
	})
}

func TestDeletePasswordResets(t *testing.T) {
	t.Run("Successfully deletes all password resets for user", func(t *testing.T) {
		db := newTestDB(t)

		ttl := 24 * time.Hour

		err := db.InsertPasswordReset("token1", testUsers["alice"].id, ttl)
		if err != nil {
			t.Fatal(err)
		}
		err = db.InsertPasswordReset("token2", testUsers["alice"].id, ttl)
		if err != nil {
			t.Fatal(err)
		}

		err = db.InsertPasswordReset("token3", testUsers["bob"].id, ttl)
		if err != nil {
			t.Fatal(err)
		}

		err = db.DeletePasswordResets(testUsers["alice"].id)
		assert.Nil(t, err)

		var count int
		err = db.Get(&count, "SELECT COUNT(*) FROM password_resets WHERE user_id = $1", testUsers["alice"].id)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, count, 0)

		err = db.Get(&count, "SELECT COUNT(*) FROM password_resets WHERE user_id = $1", testUsers["bob"].id)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 1, count)
	})

	t.Run("Does not error when no password resets exist for user", func(t *testing.T) {
		db := newTestDB(t)

		userID := 9999

		err := db.DeletePasswordResets(userID)
		assert.Nil(t, err)
	})
}
