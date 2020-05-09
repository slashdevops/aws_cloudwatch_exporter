package config

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/prometheus/client_golang/prometheus"
)

type All struct {
	ServerConf         `mapstructure:",squash"`
	ApplicationConf    `mapstructure:",squash"`
	CredentialsConf    `mapstructure:",squash"`
	MetricsQueriesConf `mapstructure:",squash"`
}

// server.conf file
// Nested:
// server:
//   address:
//   port:
type ServerConf struct {
	Server `mapstructure:",squash"`
}

type Server struct {
	Address     string `mapstructure:"address" json:"Address"`
	Port        uint16 `mapstructure:"port" json:"Port"`
	MetricsPath string `mapstructure:"metricsPath" json:"MetricsPath"`
}

// No File conf
// Nested:
// application:
//   name:
//   description:
type ApplicationConf struct {
	Application `mapstructure:",squash"`
}

type Application struct {
	Name        string `mapstructure:"name" json:"Name"`
	Description string `mapstructure:"description" json:"Description"`
	Version     string `mapstructure:"version" json:"Version"`
	Namespace   string `mapstructure:"namespace" json:"Namespace"`
	Gatherer    prometheus.Gatherer
	AWSSession  *session.Session
}

// credentials.yaml file
// Nested:
// credentials:
//   aws_access_key_id:
//   aws_secret_access_key:
type CredentialsConf struct {
	Credentials `mapstructure:",squash"`
}

type Credentials struct {
	AccessKeyID          string   `mapstructure:"aws_access_key_id" json:"AccessKeyID" yaml:"AccessKeyID"`
	SecretAccessKey      string   `mapstructure:"aws_secret_access_key"`
	SessionToken         string   `mapstructure:"aws_session_token"`
	Region               string   `mapstructure:"region"`
	Profile              string   `mapstructure:"profile" json:"Profile" yaml:"Profile"`
	RoleArn              string   `mapstructure:"role_arn"`
	RoleSessionName      string   `mapstructure:"role_session_name"`
	WebIdentityTokenFile string   `mapstructure:"web_identity_token_file"`
	ExternalID           string   `mapstructure:"external_id"`
	MFASerial            string   `mapstructure:"mfa_serial"`
	SharedConfigState    bool     `mapstructure:"shared_config_state"`
	CredentialsFile      []string `mapstructure:"aws_shared_credential_file"`
	ConfigFile           []string `mapstructure:"aws_config_file"`
}

// File conf metrics.yaml
// Will be filled with que Metrics Queries
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
// NOTE: This structure is nested because I don't use its internal structure in anywhere
type MetricsQueriesConf struct {
	MetricDataQueries []struct {
		ID         string `mapstructure:"Id" json:"Id" yaml:"Id"`
		MetricStat struct {
			Metric struct {
				Namespace  string `mapstructure:"Namespace" json:"Namespace" yaml:"Namespace"`
				MetricName string `mapstructure:"MetricName" json:"MetricName" yaml:"MetricName"`
				Dimensions []struct {
					Name  string `mapstructure:"Name" json:"Name" yaml:"Name"`
					Value string `mapstructure:"Value" json:"Value" yaml:"Value"`
				} `mapstructure:"Dimensions" json:"Dimensions" yaml:"Dimensions"`
			} `mapstructure:"Metric" json:"Metric" yaml:"Metric"`
			Period int64  `mapstructure:"Period" json:"Period" yaml:"Period"`
			Stat   string `mapstructure:"Stat" json:"Stat" yaml:"Stat"`
		} `mapstructure:"MetricStat" json:"MetricStat" yaml:"MetricStat"`
	} `mapstructure:"MetricDataQueries" json:"MetricDataQueries" yaml:"MetricDataQueries"`
}
