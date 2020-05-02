package config

import log "github.com/sirupsen/logrus"

type All struct {
	ServerConf         `mapstructure:",squash"`
	ApplicationConf    `mapstructure:",squash"`
	CredentialsConf    `mapstructure:",squash"`
	MetricsQueriesConf `mapstructure:",squash"`
}
type ServerConf struct {
	Server struct {
		Address     string `mapstructure:"address" json:"Address"`
		Port        uint16 `mapstructure:"port" json:"Port"`
		MetricsPath string `mapstructure:"metricsPath" json:"MetricsPath"`
	}
}

// No File conf
type ApplicationConf struct {
	Application struct {
		Name        string `mapstructure:"name" json:"Name"`
		Description string `mapstructure:"description" json:"Description"`
		Version     string `mapstructure:"version" json:"Version"`
		Logger      *log.Logger
	}
}

type CredentialsConf struct {
	Credentials struct {
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
}

// File conf metrics.yaml
// Will be filled with que Metrics Queries
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
// https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
type MetricsQueriesConf struct {
	MetricDataQueries []struct {
		ID         string `mapstructure:"Id" json:"Id"`
		MetricStat struct {
			Metric struct {
				Namespace  string `mapstructure:"Namespace" json:"Namespace"`
				MetricName string `mapstructure:"MetricName" json:"MetricName"`
				Dimensions []struct {
					Name  string `mapstructure:"Name" json:"Name"`
					Value string `mapstructure:"Value" json:"Value"`
				}
			}
			Period int32  `mapstructure:"Period" json:"Period"`
			Stat   string `mapstructure:"Stat" json:"Stat"`
			Unit   string `mapstructure:"Unit" json:"Unit"`
		}
	}
}
