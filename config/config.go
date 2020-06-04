package config

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/yaml.v2"
)

type All struct {
	ServerConf            `mapstructure:",squash"`
	ApplicationConf       `mapstructure:",squash"`
	AWSConf               `mapstructure:",squash"`
	MetricDataQueriesConf `mapstructure:",squash"`
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

// This is a convenient structure to allow config files nested (server.[keys])
// server.conf file
// server:
//   address:
//   port:
type ServerConf struct {
	Server `mapstructure:"server" json:"Server" yaml:"Server"`
}

type Server struct {
	Address string `mapstructure:"address" json:"Address"`
	Port    uint16 `mapstructure:"port" json:"Port"`
}

// This is a convenient structure to allow config files nested (application.[keys])
// No File conf
// application:
//   name:
//   description:
type ApplicationConf struct {
	Application `mapstructure:"application" json:"Application" yaml:"Application"`
}

type Application struct {
	Name             string   `mapstructure:"name" json:"Name" yaml:"Name"`
	Description      string   `mapstructure:"description" json:"Description" yaml:"Description"`
	Version          string   `mapstructure:"version" json:"Version" yaml:"Version"`
	Revision         string   `mapstructure:"revision" json:"Revision" yaml:"Revision"`
	Branch           string   `mapstructure:"branch" json:"Branch" yaml:"Branch"`
	BuildUser        string   `mapstructure:"buildUser" json:"BuildUser" yaml:"BuildUser"`
	BuildDate        string   `mapstructure:"buildDate" json:"BuildDate" yaml:"BuildDate"`
	GoVersion        string   `mapstructure:"goVersion" json:"GoVersion" yaml:"GoVersion"`
	Info             string   `mapstructure:"info" json:"Info" yaml:"Info"`
	BuildInfo        string   `mapstructure:"buildInfo" json:"BuildInfo" yaml:"BuildInfo"`
	Namespace        string   `mapstructure:"namespace" json:"Namespace" yaml:"Namespace"`
	ServerFile       string   `mapstructure:"serverFile" json:"ServerFile" yaml:"ServerFile"`
	CredentialsFile  string   `mapstructure:"credentialsFile" json:"SharedCredentialsFile" yaml:"SharedCredentialsFile"`
	MetricsPath      string   `mapstructure:"metricsPath" json:"MetricsPath" yaml:"MetricsPath"`
	MetricsFiles     []string `mapstructure:"metricsFiles" json:"MetricsFiles" yaml:"MetricsFiles"`
	MetricStatPeriod string   `mapstructure:"metricStatPeriod" json:"MetricStatPeriod" yaml:"MetricStatPeriod"`
	MetricTimeWindow string   `mapstructure:"metricTimeWindow" json:"MetricTimeWindow" yaml:"MetricTimeWindow"`
	Gatherer         prometheus.Gatherer
	AWSSession       *session.Session
}

// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
// This is a convenient structure to allow config files nested (aws.[keys])
// credentials.yaml file
// Nested:
// aws:
//   access_key_id: ""
//   secret_access_key: ""
//   profile: ""
type AWSConf struct {
	AWS `mapstructure:"AWS" json:"AWS" yaml:"AWS"`
}

type AWS struct {
	AccessKeyID           string   `mapstructure:"access_key_id" json:"AccessKeyID" yaml:"AccessKeyID"`
	SecretAccessKey       string   `mapstructure:"secret_access_key" json:"SecretAccessKey" yaml:"SecretAccessKey"`
	SessionToken          string   `mapstructure:"session_token" json:"SessionToken" yaml:"SessionToken"`
	Region                string   `mapstructure:"region" json:"Region" yaml:"Region"`
	Profile               string   `mapstructure:"profile" json:"Profile" yaml:"Profile"`
	RoleArn               string   `mapstructure:"role_arn" json:"RoleArn" yaml:"RoleArn"`
	RoleSessionName       string   `mapstructure:"role_session_name" json:"RoleSessionName" yaml:"RoleSessionName"`
	WebIdentityTokenFile  string   `mapstructure:"web_identity_token_file" json:"WebIdentityTokenFile" yaml:"WebIdentityTokenFile"`
	ExternalID            string   `mapstructure:"external_id" json:"ExternalID" yaml:"ExternalID"`
	MFASerial             string   `mapstructure:"mfa_serial" json:"MFASerial" yaml:"MFASerial"`
	SharedConfigState     bool     `mapstructure:"shared_config_state" json:"SharedConfigState" yaml:"SharedConfigState"`
	SharedCredentialsFile []string `mapstructure:"shared_credential_file" json:"SharedCredentialsFile" yaml:"SharedCredentialsFile"`
	ConfigFile            []string `mapstructure:"config_file" json:"ConfigFile" yaml:"ConfigFile"`
}

// This is a convenient structure to allow config files nested (MetricDataQueries.[keys])
// File conf metrics.yaml
// Will be filled with que Metrics Queries
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricStat.html
// NOTE: This structure is nested because I don't use its internal structure in anywhere
type MetricDataQueriesConf struct {
	MetricDataQueries []MetricDataQuery `mapstructure:"MetricDataQueries" json:"MetricDataQueries" yaml:"MetricDataQueries"`
}

//
type MetricDataQuery struct {
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
		Unit   string `mapstructure:"Unit" json:"Unit" yaml:"Unit"`
	} `mapstructure:"MetricStat" json:"MetricStat" yaml:"MetricStat"`
}
