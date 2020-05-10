package config

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
)

type All struct {
	ServerConf         `mapstructure:",squash"`
	ApplicationConf    `mapstructure:",squash"`
	CredentialsConf    `mapstructure:",squash"`
	MetricsQueriesConf `mapstructure:",squash"`
}

func (c *All) ToJson() string {
	out, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func (c *All) ToYaml() string {
	out, err := yaml.Marshal(c)
	if err != nil {
		panic(err)
	}
	return string(out)
}

// server.conf file
// Nested:
// server:
//   address:
//   port:
type ServerConf struct {
	Server `mapstructure:"server" json:"Server" yaml:"Server"`
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
	Application `mapstructure:"application" json:"Application" yaml:"Application"`
}

type Application struct {
	Name            string   `mapstructure:"name" json:"Name" yaml:"Name"`
	Description     string   `mapstructure:"description" json:"Description" yaml:"Description"`
	Version         string   `mapstructure:"version" json:"Version" yaml:"Version"`
	Namespace       string   `mapstructure:"namespace" json:"Namespace" yaml:"Namespace"`
	ServerFile      string   `mapstructure:"serverFile" json:"ServerFile" yaml:"ServerFile"`
	CredentialsFile string   `mapstructure:"credentialsFile" json:"CredentialsFile" yaml:"CredentialsFile"`
	MetricsFiles    []string `mapstructure:"metricsFiles" json:"MetricsFiles" yaml:"MetricsFiles"`
	StatsPeriod     string   `mapstructure:"statsPeriod" json:"StatsPeriod" yaml:"StatsPeriod"`
	Gatherer        prometheus.Gatherer
	AWSSession      *session.Session
}

// credentials.yaml file
// Nested:
// credentials:
//   aws_access_key_id:
//   aws_secret_access_key:
type CredentialsConf struct {
	Credentials `mapstructure:"Credentials" json:"Credentials" yaml:"Credentials"`
}

type Credentials struct {
	AccessKeyID          string   `mapstructure:"access_key_id" json:"AccessKeyID" yaml:"AccessKeyID"`
	SecretAccessKey      string   `mapstructure:"secret_access_key" json:"SecretAccessKey" yaml:"SecretAccessKey"`
	SessionToken         string   `mapstructure:"session_token" json:"SessionToken" yaml:"SessionToken"`
	Region               string   `mapstructure:"region" json:"Region" yaml:"Region"`
	Profile              string   `mapstructure:"profile" json:"Profile" yaml:"Profile"`
	RoleArn              string   `mapstructure:"role_arn" json:"RoleArn" yaml:"RoleArn"`
	RoleSessionName      string   `mapstructure:"role_session_name" json:"RoleSessionName" yaml:"RoleSessionName"`
	WebIdentityTokenFile string   `mapstructure:"web_identity_token_file" json:"WebIdentityTokenFile" yaml:"WebIdentityTokenFile"`
	ExternalID           string   `mapstructure:"external_id" json:"ExternalID" yaml:"ExternalID"`
	MFASerial            string   `mapstructure:"mfa_serial" json:"MFASerial" yaml:"MFASerial"`
	SharedConfigState    bool     `mapstructure:"shared_config_state" json:"SharedConfigState" yaml:"SharedConfigState"`
	CredentialsFile      []string `mapstructure:"aws_shared_credential_file" json:"CredentialsFile" yaml:"CredentialsFile"`
	ConfigFile           []string `mapstructure:"aws_config_file" json:"ConfigFile" yaml:"ConfigFile"`
}

// File conf metrics.yaml
// Will be filled with que Metrics Queries
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
// NOTE: This structure is nested because I don't use its internal structure in anywhere
type MetricsQueriesConf struct {
	MetricDataQueries `mapstructure:"MetricDataQueries" json:"MetricDataQueries" yaml:"MetricDataQueries"`
}

type MetricDataQueries []struct {
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
}
