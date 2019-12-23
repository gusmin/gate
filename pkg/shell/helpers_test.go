package shell

import (
	"io/ioutil"
	"os"

	"github.com/stretchr/testify/require"
)

// mockInput creates a new open file descriptor which contains
// user inputs and makes the caller test fail if any error occurs.
func mockInput(assert *require.Assertions, input string) *os.File {
	tmp, err := ioutil.TempFile("", "")
	assert.NoError(err)

	_, err = tmp.WriteString(input)
	assert.NoError(err)

	_, err = tmp.Seek(0, os.SEEK_SET)
	assert.NoError(err)

	return tmp
}
