package commands

import (
	"testing"

	"github.com/gusmin/gate/pkg/backend"
	"github.com/gusmin/gate/pkg/core"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

type mockTranslator struct{}

func (t mockTranslator) Translate(msg string) string {
	return msg
}

func TestExecute(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name string
		cmd  string
		err  string
	}{
		{
			name: "empty command",
			cmd:  "",
			err:  "",
		},
		{
			name: "only whitespace",
			cmd:  "\t\n",
			err:  "",
		},
		{
			name: "execute a command successfully",
			cmd:  "me",
			err:  "",
		},
	}

	core := core.New(
		"randomuser",
		nil,
		nil,
		logrus.StandardLogger(),
		&mockTranslator{},
		nil,
	)
	cmd := NewSecureGateCommand(core)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := cmd.Execute(tc.cmd)
			if err != nil {
				assert.Equalf(tc.err, errors.Cause(err).Error(),
					"expected errror was %v, but got %v", tc.err, err)
			}
		})
	}
}

func TestMe(t *testing.T) {
	fs := afero.NewMemMapFs()
	assert := require.New(t)

	user := backend.User{
		ID:        "foobar42",
		Email:     "foo.bar@gmail.com",
		FirstName: "foo",
		LastName:  "bar",
		Job:       "gopher",
	}

	file, err := afero.TempFile(fs, "", "")
	assert.NoError(err)
	defer fs.Remove(file.Name())

	logrus.SetOutput(file)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	me(user, logrus.StandardLogger(), &mockTranslator{})

	const expected = "level=info msg=\"+-------------------+-----------+----------+--------+\\n|       EMAIL       | FIRSTNAME | LASTNAME |  JOB   |\\n+-------------------+-----------+----------+--------+\\n| foo.bar@gmail.com | foo       | bar      | gopher |\\n+-------------------+-----------+----------+--------+\\nMeCaption\\n\\n\" user=foobar42\n"

	b, err := afero.ReadFile(fs, file.Name())
	assert.NoError(err)
	actual := string(b)

	assert.Equalf(expected, actual, "expected output was %s but actual is %s", expected, actual)
}

func TestList(t *testing.T) {
	fs := afero.NewMemMapFs()
	assert := require.New(t)

	user := backend.User{
		ID:        "foobar42",
		Email:     "foo.bar@gmail.com",
		FirstName: "foo",
		LastName:  "bar",
		Job:       "gopher",
	}

	machines := []backend.Machine{
		{
			ID:        "nowhere42",
			Name:      "nowhere",
			IP:        "localhost",
			AgentPort: 3002,
		},
	}

	file, err := afero.TempFile(fs, "", "")
	assert.NoError(err)
	defer fs.Remove(file.Name())

	logrus.SetOutput(file)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})

	list(user, machines, logrus.StandardLogger(), &mockTranslator{})

	const expected = "level=info msg=\"+-----------+---------+-----------+-----------+\\n|    ID     |  NAME   |    IP     | AGENTPORT |\\n+-----------+---------+-----------+-----------+\\n| nowhere42 | nowhere | localhost |      3002 |\\n+-----------+---------+-----------+-----------+\\nListCaption\\n\" user=foobar42\n"
	b, err := afero.ReadFile(fs, file.Name())
	assert.NoError(err)
	actual := string(b)

	assert.Equalf(expected, actual, "expected output was %s but actual is %s", expected, actual)
}

func TestConnect(t *testing.T) {
	assert := require.New(t)

	tt := []struct {
		name      string
		connectTo string
		machines  []backend.Machine
		err       string
	}{
		{
			name:      "machine does not exist",
			connectTo: "NASA",
			machines:  []backend.Machine{},
			err:       "NASA is not part of accessible machines",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := connect(
				"foo",
				tc.connectTo,
				backend.User{ID: "foobar42"},
				tc.machines,
				logrus.StandardLogger(),
			)
			if err != nil {
				assert.Equalf(tc.err, err.Error(),
					"expected error was: %v, but it is: %v", tc.err, err)
			}
		})
	}
}
