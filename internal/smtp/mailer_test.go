package smtp

import (
	"strings"
	"testing"

	"github.com/jcroyoaun/totalcompmx/internal/assert"
)

func TestNewMailer(t *testing.T) {
	t.Run("Create mailer with valid configuration successfully", func(t *testing.T) {
		mailer, err := NewMailer("smtp.github.com/jcroyoaun/totalcompmx", 587, "user@github.com/jcroyoaun/totalcompmx", "password", "from@github.com/jcroyoaun/totalcompmx")

		assert.Nil(t, err)
		assert.NotNil(t, mailer)
		assert.Equal(t, mailer.from, "from@github.com/jcroyoaun/totalcompmx")
		assert.NotNil(t, mailer.client)
		assert.Equal(t, mailer.mockSend, false)
	})
}

func TestNewMockMailer(t *testing.T) {
	t.Run("Create mock mailer successfully", func(t *testing.T) {
		mailer := NewMockMailer("mock@github.com/jcroyoaun/totalcompmx")

		assert.NotNil(t, mailer)
		assert.Equal(t, mailer.from, "mock@github.com/jcroyoaun/totalcompmx")
		assert.Equal(t, mailer.mockSend, true)
		assert.Equal(t, len(mailer.SentMessages), 0)
	})
}

func TestSend(t *testing.T) {
	t.Run("Send email successfully with mock mailer", func(t *testing.T) {
		mailer := NewMockMailer("sender@github.com/jcroyoaun/totalcompmx")

		err := mailer.Send("recipient@github.com/jcroyoaun/totalcompmx", "test data", "testdata/test.tmpl")
		assert.Nil(t, err)
		assert.Equal(t, len(mailer.SentMessages), 1)
		assert.True(t, strings.Contains(mailer.SentMessages[0], "From: <sender@github.com/jcroyoaun/totalcompmx>"))
		assert.True(t, strings.Contains(mailer.SentMessages[0], "To: <recipient@github.com/jcroyoaun/totalcompmx>"))
		assert.True(t, strings.Contains(mailer.SentMessages[0], "Subject: Test subject"))
		assert.True(t, strings.Contains(mailer.SentMessages[0], "This is a test plaintext email with TEST DATA"))
		assert.True(t, strings.Contains(mailer.SentMessages[0], "<p>This is a test HTML email with TEST DATA</p>"))
	})

	t.Run("Send multiple emails and track all messages", func(t *testing.T) {
		mailer := NewMockMailer("sender@github.com/jcroyoaun/totalcompmx")

		err := mailer.Send("recipient1@github.com/jcroyoaun/totalcompmx", "test data", "testdata/test.tmpl")
		assert.Nil(t, err)

		err = mailer.Send("recipient2@github.com/jcroyoaun/totalcompmx", "test data", "testdata/test.tmpl")
		assert.Nil(t, err)
		assert.Equal(t, len(mailer.SentMessages), 2)
		assert.True(t, strings.Contains(mailer.SentMessages[0], "To: <recipient1@github.com/jcroyoaun/totalcompmx>"))
		assert.True(t, strings.Contains(mailer.SentMessages[1], "To: <recipient2@github.com/jcroyoaun/totalcompmx>"))
	})

	t.Run("Returns error for non-existent email template file", func(t *testing.T) {
		mailer := NewMockMailer("sender@github.com/jcroyoaun/totalcompmx")

		err := mailer.Send("recipient@github.com/jcroyoaun/totalcompmx", nil, "testdata/non-existent-file.tmpl")
		assert.NotNil(t, err)
	})
}
