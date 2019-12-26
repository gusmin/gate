package config

import (
	"log"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Configuration is the config used by the Gate.
type Configuration struct {
	SSHUser        string `mapstructure:"ssh_user"`
	BackendURI     string `mapstructure:"backend_uri"`
	AgentAuthToken string `mapstructure:"agent_authentication_token"`
	Language       string `mapstructure:"language"`
	DBPath         string `mapstructure:"db_path"`
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

// FromFile load the configuration from the given file path.
func FromFile(path string) (Configuration, error) {
	v := viper.New()
	setDefaults(v)

	// Get the filename without the extension
	base := filepath.Base(path)
	filename := strings.TrimSuffix(base, filepath.Ext(base))
	v.SetConfigName(filename)

	// Get the directory where the config file is located
	path = filepath.Dir(path)
	v.AddConfigPath(path)

	if err := v.ReadInConfig(); err != nil {
		return Configuration{}, errors.Wrapf(err, "%s could not be loaded", path)
	}

	var cfg Configuration
	if err := v.UnmarshalExact(&cfg); err != nil {
		return Configuration{}, errors.Wrapf(err, "%s could not be loaded", path)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("ssh_user", "secure")
	v.SetDefault("language", "en")
	v.SetDefault("db_path", "/var/lib/securegate/gate/securegate.db")
}
