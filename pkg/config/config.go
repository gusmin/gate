package config

import (
	"log"
	"reflect"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const configPath = "/etc/securegate/gate/"

// Configuration is the config used by the Secure Gate shell.
type Configuration struct {
	BackendURI     string `mapstructure:"backend_uri"`
	AgentAuthToken string `mapstructure:"agent_authentication_token"`
	SSHUser        string `mapstructure:"ssh_user"`
	Language       string `mapstructure:"language"`
}

// FromFile load the configuration from the given file.
func FromFile(filename string) (Configuration, error) {
	v := viper.New()

	v.SetConfigName(filename)
	v.AddConfigPath(configPath)
	if err := v.ReadInConfig(); err != nil {
		return Configuration{}, errors.Wrapf(err, "%s could not be loaded", filename)
	}
	var cfg Configuration
	if err := v.Unmarshal(&cfg); err != nil {
		return Configuration{}, errors.Wrapf(err, "%s could not be loaded", filename)
	}
	return cfg, nil
}

// Debug prints the given configuration struct.
func Debug(cfg interface{}) {
	log.Println("--------------------------------------------------------------")
	v := reflect.ValueOf(cfg).Elem()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		log.Printf("%s %s = %v\n", f.Type(), v.Type().Field(i).Name, f.Interface())
	}
	log.Println("--------------------------------------------------------------")
}
