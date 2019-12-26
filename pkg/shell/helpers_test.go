package shell

import (
	"os"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// mockInput creates a new open file descriptor which contains
// user inputs and makes the caller test fail if any error occurs.
func mockInput(fs afero.Fs, assert *require.Assertions, input string) afero.File {
	tmp, err := afero.TempFile(fs, "", "")
	assert.NoError(err)

	_, err = tmp.WriteString(input)
	assert.NoError(err)

	_, err = tmp.Seek(0, os.SEEK_SET)
	assert.NoError(err)

	return tmp
}
