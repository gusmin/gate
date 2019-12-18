package shell

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestReadline(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name         string
		line         string
		expectedLine string
		err          string
	}{
		{
			name:         "valid line",
			line:         "foo",
			expectedLine: "foo",
			err:          "",
		},
		{
			name:         "EOF",
			line:         "",
			expectedLine: "",
			err:          "EOF",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := mockInput(assert, tc.line)
			defer os.Remove(f.Name())

			prompt, err := NewSecureGatePrompt(f, nil)
			assert.NoError(err)
			defer prompt.Close()

			line, err := prompt.Readline("test$")
			if err != nil {
				assert.Equalf(tc.err, err.Error(),
					"expected error was %v but got %v", tc.err, err)
			}

			assert.Equalf(tc.expectedLine, line,
				"expected line was %s but got %s", tc.expectedLine, line)
		})
	}
}

func TestReadPassword(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name         string
		line         string
		expectedLine string
		err          string
	}{
		{
			name:         "valid lines",
			line:         "foo",
			expectedLine: "foo",
			err:          "",
		},
		{
			name:         "EOF",
			line:         "",
			expectedLine: "",
			err:          "EOF",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			f := mockInput(assert, tc.line)
			defer os.Remove(f.Name())

			prompt, err := NewSecureGatePrompt(f, nil)
			assert.NoError(err)
			defer prompt.Close()

			line, err := prompt.ReadPassword("test$")
			if err != nil {
				assert.Equalf(tc.err, errors.Cause(err).Error(),
					"expected error was %v but got %v", tc.err, err)
			}

			assert.Equalf(tc.expectedLine, line,
				"expected line was %s but got %s", tc.expectedLine, line)
		})
	}
}
