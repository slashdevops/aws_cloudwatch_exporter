package metrics

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/prometheus/common/log"
	"github.com/slashdevops/aws_cloudwatch_exporter/config"
	"gopkg.in/yaml.v2"
)

func prepareMetrics() *config.MetricDataQueriesConf {
	MetricDataQueriesYaml := `
MetricDataQueries:
  - Id: m1
    MetricStat:
      Metric:
        Namespace: AWS/EC2
        MetricName: CPUUtilization
        Dimensions:
          - Name: AutoScalingGroupName
            Value: my-asg
      Stat: Average
`
	c := config.MetricDataQueriesConf{}
	err := yaml.Unmarshal([]byte(MetricDataQueriesYaml), &c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return &c
}

func prepareAWSMetrics() *cloudwatch.GetMetricDataInput {

	return &cloudwatch.GetMetricDataInput{
		EndTime: aws.Time(parseDate("2020-05-10T11:10:00Z", time.RFC3339)),
		MetricDataQueries: []*cloudwatch.MetricDataQuery{&cloudwatch.MetricDataQuery{
			Id: aws.String("m1"),
			MetricStat: &cloudwatch.MetricStat{
				Metric: &cloudwatch.Metric{
					Dimensions: []*cloudwatch.Dimension{
						{
							Name:  aws.String("AutoScalingGroupName"),
							Value: aws.String("my-asg"),
						},
					},
					MetricName: aws.String("CPUUtilization"),
					Namespace:  aws.String("AWS/EC2"),
				},
				Period: aws.Int64(300),
				Stat:   aws.String("Average"),
			},
			ReturnData: aws.Bool(true),
		}},
		ScanBy:    aws.String("TimestampDescending"),
		StartTime: aws.Time(parseDate("2020-05-10T11:00:00Z", time.RFC3339)),
	}
}

func Test_metrics_GetMetricDataInput(t *testing.T) {
	type fields struct {
		MetricsQueries *config.MetricDataQueriesConf
	}
	type args struct {
		st time.Time
		et time.Time
		p  time.Duration
		nt string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *cloudwatch.GetMetricDataInput
	}{
		{
			name: "FirstCase",
			fields: fields{
				MetricsQueries: prepareMetrics(),
			},
			args: args{
				st: parseDate("2020-05-10T11:00:00Z", time.RFC3339),
				et: parseDate("2020-05-10T11:10:00Z", time.RFC3339),
				p:  parseDuration("5m"),
				nt: "",
			},
			want: prepareAWSMetrics(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &metrics{
				MetricsQueries: tt.fields.MetricsQueries,
			}
			if got := m.GetMetricDataInput(tt.args.st, tt.args.et, tt.args.p, tt.args.nt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetricDataInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func parseDuration(d string) time.Duration {
	td, err := time.ParseDuration(d)
	if err != nil {
		log.Errorf("Error parsing period: %v, %v", d, err)
	}
	return td
}

func parseDate(d string, l string) time.Time {
	td, err := time.Parse(l, d)
	if err != nil {
		log.Errorf("Error parsing date: %v, %v", d, err)
	}
	return td
}

func Test_GetTimeStamps(t *testing.T) {
	type args struct {
		now time.Time
		p   string
	}
	tests := []struct {
		name          string
		args          args
		wantStartTime time.Time
		wantEndTime   time.Time
		wantPeriod    time.Duration
	}{
		{
			name: "Test5mPeriodExact",
			args: args{
				now: parseDate("2020-05-10T11:05:00Z", time.RFC3339),
				p:   "5m",
			},
			wantStartTime: parseDate("2020-05-10T11:00:00Z", time.RFC3339),
			wantEndTime:   parseDate("2020-05-10T11:10:00Z", time.RFC3339),
			wantPeriod:    parseDuration("5m"),
		},
		{
			name: "Test5mPeriodBefore",
			args: args{
				now: parseDate("2020-05-10T11:04:59Z", time.RFC3339),
				p:   "5m",
			},
			wantStartTime: parseDate("2020-05-10T10:55:00Z", time.RFC3339),
			wantEndTime:   parseDate("2020-05-10T11:05:00Z", time.RFC3339),
			wantPeriod:    parseDuration("5m"),
		},
		{
			name: "Test5mPeriodAfter",
			args: args{
				now: parseDate("2020-05-10T11:05:59Z", time.RFC3339),
				p:   "5m",
			},
			wantStartTime: parseDate("2020-05-10T11:00:00Z", time.RFC3339),
			wantEndTime:   parseDate("2020-05-10T11:10:00Z", time.RFC3339),
			wantPeriod:    parseDuration("5m"),
		},
		{
			name: "Test5mPeriodEndDay",
			args: args{
				now: parseDate("2020-05-11T00:04:59Z", time.RFC3339),
				p:   "5m",
			},
			wantStartTime: parseDate("2020-05-10T23:55:00Z", time.RFC3339),
			wantEndTime:   parseDate("2020-05-11T00:05:00Z", time.RFC3339),
			wantPeriod:    parseDuration("5m"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStartTime, gotEndTime, gotPeriod := GetTimeStamps(tt.args.now, tt.args.p)
			if gotStartTime != tt.wantStartTime {
				t.Errorf("GetTimeStamps() gotStartTime = %v, want %v", gotStartTime, tt.wantStartTime)
			}
			if gotEndTime != tt.wantEndTime {
				t.Errorf("GetTimeStamps() gotEndTime = %v, want %v", gotEndTime, tt.wantEndTime)
			}
			if gotPeriod != tt.wantPeriod {
				t.Errorf("GetTimeStamps() gotPeriod = %v, want %v", gotPeriod, tt.wantPeriod)
			}
		})
	}
}
