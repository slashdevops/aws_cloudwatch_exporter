package config

type Metrics struct {
	AWS AWSMetrics
}

//AWSConfig reference https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
type AWSMetrics struct {
	MetricDataQueries []struct {
		ID         string `json:"Id"`
		MetricStat struct {
			Metric struct {
				Namespace  string `json:"Namespace"`
				MetricName string `json:"MetricName"`
				Dimensions []struct {
					Name  string `json:"Name"`
					Value string `json:"Value"`
				}
			}
			Period int32  `json:"Period"`
			Stat   string `json:"Stat"`
			Unit   string `json:"Unit"`
		}
	}
}
