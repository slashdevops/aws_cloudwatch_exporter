package config

type Config struct {
	App    AppConfig
	Server ServerConfig
	AWS    AWSConfig
}

type AppConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ServerConfig struct {
	Address     string `json:"Address"`
	Port        uint16 `json:"Port"`
	MetricsPath string `json:"MetricsPath"`
}

//AWSConfig reference https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
type AWSConfig struct {
	Credentials []struct {
		AWSAccessKeyID          string `json:"aws_access_key_id"`
		AWSSecretAccessKey      string `json:"aws_secret_access_key"`
		AWSRegion               string `json:"aws_region"`
		AWSProfile              string `json:"aws_profile"`
		AWSRoleArn              string `json:"aws_role_arn"`
		AWSWebIdentityTokenFile string `json:"aws_web_identity_token_file"`
	}
	MetricDataQueries []struct {
		ID         string `json:"Id"`
		MetricStat struct {
			Period int32  `json:"Period"`
			Stat   string `json:"Stat"`
			Unit   string `json:"Unit"`
			Metric struct {
				Namespace  string `json:"Namespace"`
				MetricName string `json:"MetricName"`
				Dimensions []struct {
					Name  string `json:"Name"`
					Value string `json:"Value"`
				}
			}
		}
	}
}
