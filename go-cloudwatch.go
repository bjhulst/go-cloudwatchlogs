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
	cliRegion      = kingpin.Flag("region", "Region which logs reside").Default("ap-southeast-2").String()
	cliGroup       = kingpin.Flag("group", "CloudWatch Logs group").Required().String()
	cliStream      = kingpin.Flag("stream", "CloudWatch Logs stream").String()
	cliStart       = kingpin.Flag("start", "Time ago to search from").Default("10m").String()
	clienv         = kingpin.Flag("env", "environment to filter by").Default(".").String()
	cliListStreams = kingpin.Flag("list-streams", "list streams for log group").Short('l').HintOptions(" ").Bool()
)

func getlogs(group *string, Stream *string) {
	dist, _ := time.ParseDuration(*cliStart)
	timefrom := aws.TimeUnixMilli(time.Now().Add(-dist).UTC())
	timeto := aws.TimeUnixMilli(time.Now().UTC())

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

	for _, e := range resp.Events {
		fmt.Println(*e.Message)
	}
}

func getstreams(groupname *string) {
	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(*groupname), // Required
		Descending:   aws.Bool(true),
		Limit:        aws.Int64(4),
		//    LogStreamNamePrefix: aws.String("LogStreamName"),
		//   NextToken:           aws.String("NextToken"),
		OrderBy: aws.String("LogStreamName"),
	}
	resp, err := svc.DescribeLogStreams(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var match bool = false
	numstreams := len(resp.LogStreams)
	r, _ := regexp.Compile("(?i)" + *clienv + ".*")
	pulledstream := ""
	for i := 0; i < numstreams; i++ {
		pulledstream = *resp.LogStreams[i].LogStreamName
		match = r.MatchString(pulledstream)
		if match {
			//			output := fmt.Sprintf("\nStream %s\n", pulledstream)
			//			fmt.Println(output)
			fmt.Println(pulledstream)
			getlogs(groupname, &pulledstream)
		}
	}
	//  fmt.Println(*resp)
}

func getloggroups() {
	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})
	params := &cloudwatchlogs.DescribeLogGroupsInput{
	//		Limit:              aws.Int64(1),
	//		LogGroupNamePrefix: aws.String("LogGroupName"),
	//		NextToken:          aws.String("NextToken"),
	}
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
			if len(*cliStream) > 0 {
				//				fmt.Println("\nStream:")
				output := fmt.Sprintf("\nStream %s\n", *cliStream)
				fmt.Println(output)
				//				getstreams(&pulledgroup)
				getlogs(&pulledgroup, cliStream)
			} else {
				fmt.Println("\nStream(s):")
				getstreams(&pulledgroup)
			}
		}
	}
}

func main() {
	kingpin.Parse()

	if *cliListStreams {
		fmt.Println("\nList of matched streams")
		getstreams(cliGroup)
		return
	} else {
		getloggroups()
	}
}
