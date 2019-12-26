package shell

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestReadline(t *testing.T) {
	fs := afero.NewMemMapFs()
	assert := require.New(t)

	tt := []struct {
		name               string
		line, expectedLine string
		expectedErr        string
	}{
		{
			name:         "valid line",
			line:         "foo",
			expectedLine: "foo",
			expectedErr:  "",
		},
		{
			name:         "EOF",
			line:         "",
			expectedLine: "",
			expectedErr:  "EOF",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := mockInput(fs, assert, tc.line)
			defer fs.Remove(f.Name())

			prompt, err := NewSecureGatePrompt(f, nil)
			assert.NoError(err)
			defer prompt.Close()

			line, err := prompt.Readline("test$")
			if err != nil {
				assert.Equalf(tc.expectedErr, err.Error(),
					"expected error was %v but got %v", tc.expectedErr, err)
			}

			assert.Equalf(tc.expectedLine, line,
				"expected line was %s but got %s", tc.expectedLine, line)
		})
	}
}

func TestReadPassword(t *testing.T) {
	fs := afero.NewMemMapFs()
	assert := require.New(t)

	tt := []struct {
		name         string
		line         string
		expectedLine string
		expectedErr  string
	}{
		{
			name:         "valid lines",
			line:         "foo",
			expectedLine: "foo",
			expectedErr:  "",
		},
		{
			name:         "EOF",
			line:         "",
			expectedLine: "",
			expectedErr:  "EOF",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := mockInput(fs, assert, tc.line)
			defer fs.Remove(f.Name())

			prompt, err := NewSecureGatePrompt(f, nil)
			assert.NoError(err)
			defer prompt.Close()

			line, err := prompt.ReadPassword("test$")
			if err != nil {
				assert.Equalf(tc.expectedErr, err.Error(),
					"expected error was %v but got %v", tc.expectedErr, err)
			}

			assert.Equalf(tc.expectedLine, line,
				"expected line was %s but got %s", tc.expectedLine, line)
		})
	}
}
