package config

type Config struct {
	Server ServerConfig
	AWS    AWSConfig
}

type ServerConfig struct {
	Address     string `json:"Address"`
	Port        uint16 `json:"Port"`
	MetricsPath string `json:"MetricsPath"`
}

//AWSConfig reference https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
type AWSConfig struct {
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
