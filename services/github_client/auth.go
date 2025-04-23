package github_client

import (
	"crypto/x509"
	"encoding/pem"
	"strconv"
	"time"

	"github.com/ammesonb/ubiquiti-config-generator/internal/errors"
	"github.com/ammesonb/ubiquiti-config-generator/services/configuration"
	"github.com/ammesonb/ubiquiti-config-generator/services/filesystem"
	"github.com/golang-jwt/jwt"
	"github.com/google/go-github/v70/github"
)

var (
	errReadKeyfile          = "failed to read keyfile"
	errListingInstallations = "failed to list installations"
	errNoInstallations      = "no installations found"
	errNoAppIDMatch         = "no installation with matching application ID found"
)

func makeJWT(gitConfig configuration.GitConfig, fsService filesystem.Service) (string, error) {
	keyfile, err := fsService.ReadFile(gitConfig.PrivateKeyPath)
	if err != nil {
		return "", errors.ErrWithParent(errReadKeyfile, err)
	}

	block, _ := pem.Decode(keyfile)
	x509Encoded := block.Bytes
	privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

	t := jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"iat": time.Now().Unix(),
			"exp": time.Now().Unix() + 600,
			"iss": gitConfig.AppID,
		})
	return t.SignedString(privateKey)
}

func installMatchesAppID(appID string) func(installation *github.Installation) bool {
	return func(installation *github.Installation) bool {
		return strconv.Itoa(int(*installation.AppID)) == appID
	}
}
