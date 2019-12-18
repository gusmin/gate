package shell

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAskCredentials(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name                            string
		emailInput, expectedEmail       string
		passwordInput, expectedPassword string
		err                             string
	}{
		{
			name:             "valid inputs",
			emailInput:       "email",
			expectedEmail:    "email",
			passwordInput:    "pass",
			expectedPassword: "pass",
			err:              "",
		},
		{
			name:             "email EOF",
			emailInput:       "",
			expectedEmail:    "",
			passwordInput:    "",
			expectedPassword: "",
			err:              "could not read email: EOF",
		},
		{
			name:             "password EOF",
			emailInput:       "email",
			expectedEmail:    "email",
			passwordInput:    "",
			expectedPassword: "",
			err:              "could not read password: EOF",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var sb strings.Builder

			if tc.emailInput != "" {
				sb.WriteString(tc.emailInput)
				if tc.passwordInput != "" {
					sb.WriteRune('\n')
					sb.WriteString(tc.passwordInput)
				}
			}

			f := mockInput(assert, sb.String())
			defer os.Remove(f.Name())

			prompt, err := NewSecureGatePrompt(f, nil)
			assert.NoError(err)
			defer prompt.Close()

			sh := NewSecureGateShell(prompt, nil, nil)

			email, password, err := sh.askForCredentials()
			if err != nil {
				assert.Equalf(tc.err, err.Error(),
					"expected error was: %v, but got %v", tc.err, err)
			}

			assert.Equalf(tc.expectedEmail, email,
				"expected input was: %s, but got %s", tc.expectedEmail, email)

			assert.Equalf(tc.expectedPassword, password,
				"expected input was: %s, but got %s", tc.expectedPassword, password)
		})
	}
}
