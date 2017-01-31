package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cliRegion = kingpin.Flag("region", "Region which logs reside").Default("ap-southeast-2").String()
	cliGroup  = kingpin.Flag("group", "CloudWatch Logs group").Required().String()
	cliStream = kingpin.Flag("stream", "CloudWatch Logs stream").String()
	cliStart  = kingpin.Flag("start", "Time ago to search from").Default("10m").String()
	cliEnd    = kingpin.Flag("end", "Time ago to end search").Default("0").String()
)

func getlogs(group *string, Stream *string) {
	diststart, _ := time.ParseDuration(*cliStart)
	distend, _ := time.ParseDuration(*cliEnd)
	timefrom := aws.TimeUnixMilli(time.Now().Add(-diststart).UTC())
	timeto := aws.TimeUnixMilli(time.Now().Add(-distend).UTC())

	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})
	resp, err := svc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  group,
		LogStreamName: Stream,
		StartFromHead: aws.Bool(true),
		StartTime:     &timefrom,
		EndTime:       &timeto,
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	if len(resp.Events) > 0 {
		fmt.Println(*Stream)
	}
	for _, e := range resp.Events {
		fmt.Println(*e.Message)
	}
}

func getstreams(groupname *string) {
	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(*groupname), // Required
		Descending:   aws.Bool(true),
		OrderBy:      aws.String("LogStreamName"),
	}
	resp, err := svc.DescribeLogStreams(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var match bool = false
	numstreams := len(resp.LogStreams)
	r, _ := regexp.Compile("(?i)" + *cliStream + ".*")
	pulledstream := ""
	for i := 0; i < numstreams; i++ {
		pulledstream = *resp.LogStreams[i].LogStreamName
		match = r.MatchString(pulledstream)
		if match {
			getlogs(groupname, &pulledstream)
		}
	}
}

/*
func getloggroups() {
	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})
	params := &cloudwatchlogs.DescribeLogGroupsInput{}
	resp, err := svc.DescribeLogGroups(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var match bool = false
	numgroups := len(resp.LogGroups)
	r, _ := regexp.Compile("(?i)" + *cliGroup + ".*")
	pulledgroup := ""
	for i := 0; i < numgroups; i++ {
		pulledgroup = *resp.LogGroups[i].LogGroupName
		match = r.MatchString(pulledgroup)
		if match {
			fmt.Println("\nLog group:")
			fmt.Println(pulledgroup)
			getstreams(&pulledgroup)
		}
	}
}
*/

func main() {
	kingpin.Parse()
	//	getloggroups()
	getstreams(cliGroup)
}
