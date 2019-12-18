package backend

import (
	"io"
	"io/ioutil"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// assertGQLVarsEq asserts if expected GraphQL variables match the one
// in the request body. This function makes the caller test fail if any error
// occurs or if expected variables are different from the ones in the body.
func assertGQLVarsEq(assert *require.Assertions, expected string, body io.Reader) {
	b, err := ioutil.ReadAll(body)
	assert.NoError(err)

	// Get GraphQL variables in the body.
	vars := gjson.GetBytes(b, "variables")
	assert.JSONEq(expected, vars.Raw)
}
