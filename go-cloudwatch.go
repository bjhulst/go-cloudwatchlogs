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

func getlogs(group *string, Stream *string) []int64 {
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
		return nil
	}
	/*	if len(resp.Events) > 0 {
			fmt.Println(fmt.Sprintf("\nStream %s:\n", *Stream))
		}
	*/
	var stamp []int64
	for _, e := range resp.Events {
		// human readable time
		//		timeoutput := time.Unix(*e.Timestamp/1000, 0).Format("15:04:05 02/01/06")
		//		stamp = append(stamp, timeoutput)

		//		msg := fmt.Sprintf("%s %s: %s", timeoutput, *Stream, *e.Message)
		//		var stamp []string
		//timeoutput := time.Unix(*e.Timestamp/1000, 0)
		//fmt.Println(timeoutput)
		stamp = append(stamp, *e.Timestamp/1000)
		//fmt.Println(msg)
	}
	for i := 0; i < len(stamp); i++ {
		//fmt.Println(stamp[i])
	}
	if len(stamp) > 0 {
		return stamp
	}
	return nil
}

func getstreams(groupname *string) {
	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: cliRegion})
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(*groupname),
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
	var slice []int64
	var slice2 []int64
	for i := 0; i < numstreams; i++ {
		pulledstream = *resp.LogStreams[i].LogStreamName
		match = r.MatchString(pulledstream)
		if match {
			slice = getlogs(groupname, &pulledstream)
			if len(slice) > 0 {
				//slice = append(slice, slice2)
				//fmt.Println("current slice")
				//fmt.Println(slice)
				for i := 0; i < len(slice); i++ {
					slice2 = append(slice2, slice[i])
					///				slice2	fmt.Println(slice[i])
				}
			}
		}
	}
	fmt.Println(slice2)
}

func main() {
	kingpin.Parse()
	getstreams(cliGroup)
}
