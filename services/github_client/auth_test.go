package github_client

import (
	"os"
	"testing"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/google/go-github/v70/github"
	"github.com/stretchr/testify/assert"
)

func TestMakeJWT(t *testing.T) {
	fs, err := filesystem.New()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Nonexistent file", func(t *testing.T) {
		_, err := makeJWT(configuration.GitConfig{
			PrivateKeyPath: "./nonexistent_file",
		}, fs)
		assert.ErrorIs(t, err, errors.ErrWithParent(errReadKeyfile, os.ErrNotExist))
	})

	t.Run("Non-key file", func(t *testing.T) {
		_, err := makeJWT(configuration.GitConfig{
			PrivateKeyPath: "./foo.txt",
		}, fs)
		assert.ErrorIs(t, err, errors.Err(errKeyfileNotPEM))
	})

	t.Run("PKCS8 PEM", func(t *testing.T) {
		_, err := makeJWT(configuration.GitConfig{
			PrivateKeyPath: "./fake_pkcs8_key.pem",
			AppID:          "1",
		}, fs)
		assert.ErrorIs(t, err, errors.Err(errFailParsePrivateKey))
	})

	t.Run("Valid file", func(t *testing.T) {
		jwt, err := makeJWT(configuration.GitConfig{
			PrivateKeyPath: "./fake_key.pem",
			AppID:          "1",
		}, fs)
		assert.NoError(t, err)
		assert.NotEmpty(t, jwt)
	})
}

func TestInstallMatchesAppID(t *testing.T) {
	appID := int64(1)
	assert.True(t, installMatchesAppID("1")(&github.Installation{
		AppID: &appID,
	}))
	assert.False(t, installMatchesAppID("2")(&github.Installation{
		AppID: &appID,
	}))
}
