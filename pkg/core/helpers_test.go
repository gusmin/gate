package core

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestExist(t *testing.T) {
	assert := require.New(t)

	wd, err := os.Getwd()
	assert.NoError(err)

	tt := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "existing path",
			path:     wd,
			expected: true,
		},
		{
			name:     "not existing path",
			path:     "pathwhichshouldnotexistactually",
			expected: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := exist(tc.path)
			assert.Equalf(tc.expected, actual,
				"expected: %v, actual: %v\n", tc.expected, actual)
		})
	}
}

func TestGenerateSSHKeyPair(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name        string
		pubKeyPath  string
		privKeyPath string
		err         string
	}{
		{
			name:        "valid paths",
			pubKeyPath:  path.Join(os.TempDir(), "id_rsa.pub"),
			privKeyPath: path.Join(os.TempDir(), "id_rsa"),
			err:         "",
		},
		{
			name:        "invalid private key path",
			pubKeyPath:  path.Join(os.TempDir(), "id_rsa.pub"),
			privKeyPath: "",
			err:         "open : no such file or directory",
		},
		{
			name:        "invalid public key path",
			pubKeyPath:  "",
			privKeyPath: path.Join(os.TempDir(), "id_rsa"),
			err:         "open : no such file or directory",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := generateSSHKeyPair(tc.pubKeyPath, tc.privKeyPath)
			if err != nil {
				assert.Equalf(tc.err, errors.Cause(err).Error(),
					"expected error was %v, but got %v", tc.err, err)
				return
			}
			defer func() {
				os.Remove(tc.pubKeyPath)
				os.Remove(tc.privKeyPath)
			}()

			// check wether generated authorized key exists and is valid
			b, err := ioutil.ReadFile(tc.pubKeyPath)
			assert.NoError(err)
			_, _, _, _, err = ssh.ParseAuthorizedKey(b)
			assert.NoError(err)

			// check wether generated private key exists and is valid
			privateKey, err := ioutil.ReadFile(tc.privKeyPath)
			assert.NoError(err)
			_, err = ssh.ParsePrivateKey(privateKey)
			assert.NoError(err)
		})
	}
}
