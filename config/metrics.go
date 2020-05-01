package config

type Metrics struct {
	AWS AWSMetrics
}

//AWSConfig reference https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
type AWSMetrics struct {
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
