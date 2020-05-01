package config

import log "github.com/sirupsen/logrus"

type Server struct {
	App    Application
	Server ServerConfig
	AWS    AWS
	Logger *log.Logger
}

type Application struct {
	Name        string
	Description string
	Version     string
}

type ServerConfig struct {
	Address     string `mapstructure:"Address" json:"Address"`
	Port        uint16 `mapstructure:"Port" json:"Port"`
	MetricsPath string `mapstructure:"MetricsPath" json:"MetricsPath"`
}

type AWS struct {
	AccessKeyID          string   `mapstructure:"aws_access_key_id"`
	SecretAccessKey      string   `mapstructure:"aws_secret_access_key"`
	SessionToken         string   `mapstructure:"aws_session_token"`
	Region               string   `mapstructure:"region"`
	Profile              string   `mapstructure:"profile"`
	RoleArn              string   `mapstructure:"role_arn"`
	RoleSessionName      string   `mapstructure:"role_session_name"`
	WebIdentityTokenFile string   `mapstructure:"web_identity_token_file"`
	ExternalID           string   `mapstructure:"external_id"`
	MFASerial            string   `mapstructure:"mfa_serial"`
	SharedConfigState    bool     `mapstructure:"shared_config_state"`
	CredentialsFile      []string `mapstructure:"aws_shared_credential_file"`
	ConfigFile           []string `mapstructure:"aws_config_file"`
}
