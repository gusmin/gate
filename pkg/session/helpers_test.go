package session

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExist(t *testing.T) {
	assert := require.New(t)

	wd, err := os.Getwd()
	assert.NoError(err)

	testCases := []struct {
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

	for _, t := range testCases {
		actual := exist(t.path)
		assert.Equalf(t.expected, actual, "%s: expected: %v, actual: %v\n", t.expected, actual)
	}
}
