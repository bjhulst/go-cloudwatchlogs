package main

import (
	"fmt"
	//	"strconv"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliRegion = kingpin.Flag("region", "Region which logs reside").Default("ap-southeast-2").String()
	cliGroup  = kingpin.Flag("group", "CloudWatch Logs group").Required().String()
	cliStream = kingpin.Flag("stream", "CloudWatch Logs stream").Required().String()
	clistart  = kingpin.Flag("start", "CloudWatch Logs stream").Default("10m").String()
)

//func getlogs(timefrom, timeto int64) {
func getlogs(dist time.Duration) {

	timefrom := aws.TimeUnixMilli(time.Now().Add(-dist).UTC())
	timeto := aws.TimeUnixMilli(time.Now().UTC())

	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})
	resp, err := svc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  cliGroup,
		LogStreamName: cliStream,
		StartFromHead: aws.Bool(true),
		StartTime:     &timefrom,
		EndTime:       &timeto,
	})

	if err != nil {
		fmt.Println(err)
		return
	}

	// Pretty-print the response data.
	for _, e := range resp.Events {
		fmt.Println(*e.Message)
	}
}
func getstreams(groupname string) {
	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(groupname), // Required
		Descending:   aws.Bool(true),
		Limit:        aws.Int64(4),
		//    LogStreamNamePrefix: aws.String("LogStreamName"),
		//i   NextToken:           aws.String("NextToken"),
		OrderBy: aws.String("LogStreamName"),
	}
	resp, err := svc.DescribeLogStreams(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	numstreams := len(resp.LogStreams)
	for i := 0; i < numstreams; i++ {
		fmt.Println(*resp.LogStreams[i].LogStreamName)
	}
	//  fmt.Println(*resp)
}

func main() {
	kingpin.Parse()
	dist, _ := time.ParseDuration(*clistart)

	getlogs(dist)
	getstreams("previousnext-previousnext-d7")
}
