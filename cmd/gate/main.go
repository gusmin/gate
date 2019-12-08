// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gusmin/gate/pkg/agent"
	"github.com/gusmin/gate/pkg/backend"
	"github.com/gusmin/gate/pkg/commands"
	"github.com/gusmin/gate/pkg/config"
	"github.com/gusmin/gate/pkg/i18n"
	"github.com/gusmin/gate/pkg/session"
	"github.com/gusmin/gate/pkg/shell"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/gofrs/flock"
	"github.com/sirupsen/logrus"
)

const (
	configFile      = "config"
	logFile         = "/var/log/securegate/gate/gate.log"
	translationsDir = "translations"
)

func main() {
	cfg, err := config.FromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	config.Debug(&cfg)

	backendClient := backend.NewClient(cfg.BackendURI)

	agentClient := agent.NewClient(cfg.AgentAuthToken, nil)

	// Log rotation	to not pollute disk space
	rotatingLogFile := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    1,
		MaxBackups: 2,
		MaxAge:     10,
		Compress:   true,
	}
	defer rotatingLogFile.Close()
	// Avoid several instances of the shell to write
	// at the same time to the same location
	flockWriter := &flockWriter{
		writer: rotatingLogFile,
		locker: flock.New(logFile),
	}
	logger := initializeLogger(flockWriter, backendClient)

	translator := i18n.NewTranslatorFromTOMLFile(cfg.Language, translationsDir)

	sess := session.New(
		cfg.SSHUser,
		backendClient,
		agentClient,
		logger,
		translator,
	)
	command := commands.NewSecureGateCommand(sess)
	prompt := shell.NewSecureGatePrompt(sess)
	defer prompt.Close()
	sh := shell.New(prompt, command, sess)

	log.Fatal(sh.Run())
}

// flockWriter blocks until it can get an exclusive file lock
// to write with the underlying writer to the file and then
// unlock it for other incoming writes.
type flockWriter struct {
	writer io.Writer    // underlying writer
	locker *flock.Flock // file locker
}

func (w *flockWriter) Write(p []byte) (n int, err error) {
	err = w.locker.Lock()
	if err != nil {
		return
	}
	defer w.locker.Unlock()

	n, err = w.writer.Write(p)
	return
}

// dummyFormatter formats a *logrus.Entry into a simple message
type dummyFormatter struct{}

func (f *dummyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message), nil
}

// writerHook is an hook that writes formatted logs of specified logLevels
// with specified writer and send them to the backend
type writerHook struct {
	writer    io.Writer
	logLevels []logrus.Level
	formatter logrus.Formatter

	backendClient *backend.Client
}

func (hook *writerHook) Levels() []logrus.Level {
	return hook.logLevels
}

func (hook *writerHook) Fire(entry *logrus.Entry) error {
	logEntry, err := hook.formatter.Format(entry)
	if err != nil {
		return err
	}

	logObj := struct {
		Level   string    `json:"level"`
		Machine string    `json:"machine"`
		Msg     string    `json:"msg"`
		Time    time.Time `json:"time"`
		User    string    `json:"user"`
	}{}
	_ = json.Unmarshal(logEntry, &logObj)

	if logObj.Machine == "" {
		_, err = hook.writer.Write(logEntry)
		return err
	}

	// Send logs to the backend
	input := backend.MachineLogInput{
		MachineID: logObj.Machine,
		UserID:    logObj.User,
		Log:       logObj.Msg,
		// Timestamp in millisecond
		Timestamp: float64(logObj.Time.Unix() * 1000),
	}
	res, err := hook.backendClient.AddMachineLog(
		context.Background(),
		[]backend.MachineLogInput{input},
	)
	if err != nil {
		return err
	}
	if !res.AddMachineLog.Success {
		return errors.New("could not add machine log to the backend")
	}

	return nil
}

// initializeLogger adds hooks to send logs to different destinations
// depending on level and send them to the backend.
func initializeLogger(w io.Writer, client *backend.Client) session.StructuredLogger {
	logger := logrus.New()

	// Send all logs to nowhere by default
	logger.SetOutput(ioutil.Discard)

	// Send all logs excepting debug level to log file
	logger.AddHook(&writerHook{
		writer:    w,
		logLevels: logrus.AllLevels,
		formatter: new(logrus.JSONFormatter),
		// TODO: replace by specific tool to centralize logs
		backendClient: client,
	})

	// Send logs with level higher or equal to warning to stderr
	logger.AddHook(&writerHook{
		writer: os.Stderr,
		logLevels: []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			// logrus.WarnLevel,
		},
		formatter: new(dummyFormatter),
	})

	// Send info and debug logs to stdout
	logger.AddHook(&writerHook{
		writer: os.Stdout,
		logLevels: []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		},
		formatter: new(dummyFormatter),
	})

	return session.NewLogrusLogger(logger)
}
