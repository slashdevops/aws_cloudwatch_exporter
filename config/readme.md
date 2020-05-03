# 

## AWS DataQuery Values:



## MetricDataQueries

* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricStat.html
* https://docs.aws.amazon.com/cli/latest/reference/cloudwatch/get-metric-statistics.html

```go
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
			Period int64  `mapstructure:"Period" json:"Period"`
			Stat   string `mapstructure:"Stat" json:"Stat"`
		}
	}
}
```

### Stats

Where valid values are:

* SampleCount
* Average
* Sum
* Minimum
* Maximum

### Conf

```yaml
MetricDataQueries:
  - Id: m1
    MetricStat:
      Metric:
        Namespace: AWS/EC2
        MetricName: CPUUtilization
        Dimensions:
          - Name: AutoScalingGroupName
            Value: my-asg
      Period: 300
      Stat: Average
  - Id: m2
    MetricStat:
      Metric:
        Namespace: AWS/EC2
        MetricName: NetworkOut
        Dimensions:
          - Name: AutoScalingGroupName
            Value: my-asg
      Period: 300
      Stat: SampleCount
  - Id: m3
    MetricStat:
      Metric:
        Namespace: AWS/ELB
        MetricName: HealthyHostCount
        Dimensions:
          - Name: LoadBalancerName
            Value: my-lb-B
      Period: 300
      Stat: Sum
```
### Prometheus metrics name

```
appNamespace + MetricStat[Metric[MetricName]] + MetricStat[Metric[Stat]] (job="MetricStat(Metric(Namespace))",   )
aws_cloudwatch_cpu_utilization_average_sum(job="aws_ec2", instance="", auto_scaling_group_name="my-asg" ) 10
aws_cloudwatch_network_out_sample_count_sum(job="aws_ec2", instance="", auto_scaling_group_name="my-asg" ) 10
aws_cloudwatch_healthy_host_count_sum_sum(job="aws_ec2", instance="", load_balancer_name="my-lb-B" ) 10

```
