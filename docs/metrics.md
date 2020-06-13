# Metrics file queries

This is a metrics queries configuration file(s) 

```yaml
MetricDataQueries:                                   # Type: array, This array contain all the metrics query
  - Id: m1                                          
    MetricStat:
      Metric:
        Namespace: AWS/EC2
        MetricName: CPUUtilization
        Dimensions:
          - Name: AutoScalingGroupName
            Value: eks-prod-01-apps-01-aeg
      Stat: Average
  - Id: m2
    MetricStat:
      Metric:
        Namespace: AWS/DynamoDB
        MetricName: ProvisionedWriteCapacityUnits
        Dimensions:
          - Name: TableName
            Value: Player
      Stat: Maximum
```

Field: Stat

* SampleCount
* Average
* Sum
* Minimum
* Maximum

## Help links

* https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_limits.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricDataQuery.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricStat.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_Metric.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/aws-services-cloudwatch-metrics.html
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/using-metric-math.html#metric-math-syntax
* https://docs.aws.amazon.com/sdk-for-go/api/service/cloudwatch/#CloudWatch.GetMetricData
* https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
* https://aws.amazon.com/premiumsupport/knowledge-center/cloudwatch-getmetricdata-api/
* https://aws.amazon.com/cloudwatch/pricing/