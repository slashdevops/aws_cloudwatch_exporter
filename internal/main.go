package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	conf := &aws.Config{
		Region: aws.String("eu-west-1"),
	}

	// sess, err := session.NewSession(conf)
	// if err != nil {
	// 	log.Println(err.Error())
	// }

	// svc := s3.New(sess)
	// input := &s3.ListBucketsInput{}

	// result, err := svc.ListBuckets(input)
	// if err != nil {
	// 	if aerr, ok := err.(awserr.Error); ok {
	// 		switch aerr.Code() {
	// 		default:
	// 			fmt.Println(aerr.Error())
	// 		}
	// 	} else {
	// 		// Print the error, cast err to awserr.Error to get the Code and
	// 		// Message from an error.
	// 		fmt.Println(err.Error())
	// 	}
	// 	return
	// }

	// fmt.Println(result)
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:                  *conf,
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		Profile:                 "aleaplay-prod",
	}))

	svc := cloudwatch.New(sess)

	// lmi := &cloudwatch.ListMetricsInput{
	// 	MetricName: aws.String("NetworkOut"),
	// 	Namespace:  aws.String("AWS/EC2"),
	// 	Dimensions: []*cloudwatch.DimensionFilter{
	// 		&cloudwatch.DimensionFilter{
	// 			Name: aws.String("AutoScalingGroupName"),
	// 		},
	// 	},
	// }
	// result, err := svc.ListMetrics(lmi)
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	// fmt.Println("Metrics", result.Metrics)

	msi := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/EC2"),
		MetricName: aws.String("CPUUtilization"),
		Dimensions: []*cloudwatch.Dimension{
			&cloudwatch.Dimension{
				Name:  aws.String("AutoScalingGroupName"),
				Value: aws.String("eks-prod-01-apps-01-aeg"),
			},
		},
		Period:    aws.Int64(300),
		StartTime: aws.Time(time.Now().Add(time.Duration(-2) * time.Hour)),
		EndTime:   aws.Time(time.Now()),
		//ExtendedStatistics: []*string{aws.String("p25"), aws.String("p50"), aws.String("p75")},
		Statistics: []*string{aws.String("Average")},
	}
	resp, err := svc.GetMetricStatistics(msi)
	if err != nil { // resp is now filled
		fmt.Println(err)
	}
	fmt.Println(resp)
}
