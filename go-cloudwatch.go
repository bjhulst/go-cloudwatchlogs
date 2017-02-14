package cloudwatchlogs

import (
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func GetLogs(region, group, stream, start, end string) (Logs, error) {
	var logs Logs

	diststart, err := time.ParseDuration(start)
	if err != nil {
		return logs, err
	}

	distend, err := time.ParseDuration(end)
	if err != nil {
		return logs, err
	}

	var (
		timefrom = aws.TimeUnixMilli(time.Now().Add(-diststart).UTC())
		timeto   = aws.TimeUnixMilli(time.Now().Add(-distend).UTC())
	)

	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: aws.String(region)})
	resp, err := svc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
		StartFromHead: aws.Bool(true),
		StartTime:     &timefrom,
		EndTime:       &timeto,
	})
	if err != nil {
		return logs, err
	}

	for _, e := range resp.Events {
		logs = append(logs, &Log{
			Stream:    stream,
			Timestamp: time.Unix(*e.Timestamp/1000, 0),
			Message:   *e.Message,
		})
	}

	return logs, nil
}

func GetStreams(region, group, stream, start, end string) (Logs, error) {
	var logs Logs

	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: aws.String(region)})
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group),
		Descending:   aws.Bool(true),
		OrderBy:      aws.String("LogStreamName"),
	}
	resp, err := svc.DescribeLogStreams(params)
	if err != nil {
		return logs, err
	}

	r, err := regexp.Compile("(?i)" + stream + ".*")
	if err != nil {
		return logs, err
	}

	for _, s := range resp.LogStreams {
		if r.MatchString(*s.LogStreamName) {
			newLogs, err := GetLogs(region, group, *s.LogStreamName, start, end)
			if err != nil {
				return logs, err
			}

			if len(newLogs) > 0 {
				logs = MergeLogs(logs, newLogs)
			}
		}
	}

	return logs, nil
}
