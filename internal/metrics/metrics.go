package metrics

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/camelcase"
)

type Metrics interface {
	// Used to assemble the AWS GetMetricDataInput data structure
	GetMetricDataInput(time.Time, time.Time, time.Duration, string) *cloudwatch.GetMetricDataInput

	//
	SetMetric(id string, metric prometheus.Metric)
	GetMetricDesc(id string) *prometheus.Desc
	GetMetricsDesc() map[string]*prometheus.Desc
	GetMetrics() map[string]prometheus.Metric
}

type metrics struct {
	// Metrics queries structure assemble from metrics queries yaml files
	MetricDataQueriesConf *config.MetricDataQueriesConf

	// The prometheus metrics created from MetricDataQueriesConf but without values
	PrometheusMetricsDesc map[string]*prometheus.Desc
	PrometheusMetrics     map[string]prometheus.Metric
}

func New(conf *config.All) Metrics {
	return &metrics{
		MetricDataQueriesConf: &conf.MetricDataQueriesConf,
		PrometheusMetricsDesc: createPrometheusMetricsDesc(conf),
		PrometheusMetrics:     make(map[string]prometheus.Metric),
	}
}

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_GetMetricData.html
func (m *metrics) GetMetricDataInput(st time.Time, et time.Time, p time.Duration, nt string) *cloudwatch.GetMetricDataInput {
	dataQry := m.getMetricDataQuery(p)

	mdi := &cloudwatch.GetMetricDataInput{
		StartTime:         aws.Time(st),
		EndTime:           aws.Time(et),
		MetricDataQueries: dataQry,
		ScanBy:            aws.String(cloudwatch.ScanByTimestampDescending), // Get the fresh data first
	}

	// conditional parameters are added after creation
	if len(nt) > 0 {
		mdi.NextToken = aws.String(nt)
	}

	return mdi
}

// This function is used to transform the structure config.MetricDataQueriesConf which contains
// the values read from config file metrics.yaml to a cloudwatch.MetricDataQuery structure which is
// the default structure used to get cloudwatch metrics data
func (m *metrics) getMetricDataQuery(p time.Duration) []*cloudwatch.MetricDataQuery {

	// time.Duration is in nanoseconds, and the CW API need it in seconds
	period := int64(p / time.Second)

	var dataQry []*cloudwatch.MetricDataQuery

	for _, m := range m.MetricDataQueriesConf.MetricDataQueries {

		// If the metric has set the Period, override global MetricStatPeriod
		if m.MetricStat.Period != 0 {
			period = m.MetricStat.Period
		}

		// Fill the internal struct with dimension
		var dimQry []*cloudwatch.Dimension
		for _, dim := range m.MetricStat.Metric.Dimensions {
			d := &cloudwatch.Dimension{
				Name:  aws.String(dim.Name),
				Value: aws.String(dim.Value),
			}
			dimQry = append(dimQry, d)
		}

		metricsQry := &cloudwatch.MetricDataQuery{
			Id:    aws.String(m.ID),
			Label: aws.String(camelcase.ToSnake(m.MetricStat.Metric.Namespace) + "_" + camelcase.ToSnake(m.MetricStat.Metric.MetricName) + "_" + camelcase.ToSnake(m.MetricStat.Stat)),
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Dimensions: dimQry,
					MetricName: aws.String(m.MetricStat.Metric.MetricName),
					Namespace:  aws.String(m.MetricStat.Metric.Namespace),
				},
				Period: aws.Int64(period),
				Stat:   aws.String(m.MetricStat.Stat),
			},
			ReturnData: aws.Bool(true), // Return the timestamps and raw data values of this metric.
		}

		// Conditional field will be filled after
		if len(m.MetricStat.Unit) > 0 {
			metricsQry.MetricStat.Unit = aws.String(m.MetricStat.Unit)
		}

		dataQry = append(dataQry, metricsQry)
	}
	return dataQry
}

func (m *metrics) SetMetric(id string, metric prometheus.Metric) {
	m.PrometheusMetrics[id] = metric
}

func (m *metrics) GetMetricDesc(id string) *prometheus.Desc {
	i, ok := m.PrometheusMetricsDesc[id]
	if !ok {
		log.Errorf("Metric Description id: %s does not exist", id)
	}
	return i
}

func (m *metrics) GetMetricsDesc() map[string]*prometheus.Desc {
	return m.PrometheusMetricsDesc
}

func (m *metrics) GetMetrics() map[string]prometheus.Metric {
	return m.PrometheusMetrics
}

// https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/cloudwatch_concepts.html
// Create the prometheus metrics scaffolding without any value
//   - Id: m1
//    MetricStat:
//      Metric:
//        Namespace: AWS/EC2
//        MetricName: CPUUtilization            --> AWS/EC2_CPUUtilization_Average{job="aws_cloudwatch_exporter", instance="", AutoScalingGroupName="eks-prod-01-apps-01-asg"} value_from_scrap
//        Dimensions:                               aws_ec2_cpu_utilization_average{job="aws_cloudwatch_exporter", instance="", auto_scaling_group_name="eks-prod-01-apps-01-asg"} value_from_scrap
//          - Name: AutoScalingGroupName            aws_cloudwatch_exporter_aws_ec2_cpu_utilization_average{job="aws_cloudwatch_exporter", instance="", auto_scaling_group_name="eks-prod-01-apps-01-asg"} value_from_scrap
//            Value: eks-prod-01-apps-01-asg
//      Stat: Average
func createPrometheusMetricsDesc(conf *config.All) map[string]*prometheus.Desc {
	mdqc := conf.MetricDataQueriesConf
	promMetricsDesc := make(map[string]*prometheus.Desc)

	var helpTmpl = "%s represent the AWS CloudWatch Metric: %s --> %s, Dimensions: [%s], Statistic: %s%s%s"

	// for every metric query defined into the yaml files
	for _, mdq := range mdqc.MetricDataQueries {

		// Add dimensions as prometheus metric labels
		mcl := make(prometheus.Labels)
		for _, v := range mdq.MetricStat.Metric.Dimensions {
			mcl[v.Name] = v.Value
		}

		// necessary to put dimensions keys in the help query string
		var dimKeys []string
		for k := range mcl {
			dimKeys = append(dimKeys, k)
		}
		dimArray := strings.Join(dimKeys, ",")

		var mu, mp string
		// Unit and Period are conditional and we want to added it to the help query string
		if len(mdq.MetricStat.Unit) > 0 {
			mu = ", Unit: " + mdq.MetricStat.Unit
		}
		if mdq.MetricStat.Period > 0 {
			mp = ", Period: " + strconv.FormatInt(mdq.MetricStat.Period, 10) + "s"
		}

		mn := camelcase.ToSnake(mdq.MetricStat.Metric.Namespace) + "_" + camelcase.ToSnake(mdq.MetricStat.Metric.MetricName) + "_" + camelcase.ToSnake(mdq.MetricStat.Stat)
		hs := fmt.Sprintf(
			helpTmpl,
			mn,
			mdq.MetricStat.Metric.Namespace,
			mdq.MetricStat.Metric.MetricName,
			dimArray,
			mdq.MetricStat.Stat,
			mu,
			mp)

		promMetricsDesc[mdq.ID] = prometheus.NewDesc(mn, hs, nil, mcl)
	}

	return promMetricsDesc
}

// Return the necessary inputs for function NewGetMetricDataInput
//              points     period        now()
//                ↓        ↓→  ←↓         ↓
// [(startTime)............................(endTime)] → time
//
// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_MetricStat.html
// this function calculate the right startTime, endTime and period from a time.Time and string period as a parameter
// using the p (period as string) and the t (time.Time) this function calculate the startTime and endTime
// as a multiple of the period.
// The startTime is the oldest time and multiple of the period
// The endTime is the newest time (future) and multiple of the period
// The period is a time.Duration representation of the p string passed as function arg
func GetTimeStamps(t time.Time, p string, tg string) (startTime time.Time, endTime time.Time, period time.Duration) {

	period, err := time.ParseDuration(p)
	if err != nil {
		log.Errorf("Error converting period: %v, %v", p, err)
	}
	timeGap, err := time.ParseDuration(tg)
	if err != nil {
		log.Errorf("Error converting time gap: %v, %v", tg, err)
	}

	//        now()                      truncate
	// 2020-05-10T11:06:25Z    ->   2020-05-10T11:05:00Z    -> +period (endTime)    2020-05-10T11:10:00Z
	//                                                      -> -timeGap (startTime) 2020-05-10T10:55:00Z
	endTime = t.Truncate(period).Add(period * 2)
	startTime = t.Truncate(period).Add(-timeGap)
	return
}
