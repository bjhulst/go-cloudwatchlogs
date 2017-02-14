package cloudwatchlogs

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func Events(region, group, stream, start, end string) (Logs, error) {
	var logs Logs

	svc := cloudwatchlogs.New(session.New(), &aws.Config{Region: aws.String(region)})

	startDuration, err := time.ParseDuration(start)
	if err != nil {
		return logs, err
	}

	endDuration, err := time.ParseDuration(end)
	if err != nil {
		return logs, err
	}

	var (
		from = aws.TimeUnixMilli(time.Now().Add(-startDuration).UTC())
		to   = aws.TimeUnixMilli(time.Now().Add(-endDuration).UTC())
	)

	done := false

	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(group),
		LogStreamNamePrefix: aws.String(stream),
		Descending:          aws.Bool(true),
	}

	for !done {
		resp, err := svc.DescribeLogStreams(params)
		if err != nil {
			return logs, err
		}

		for _, s := range resp.LogStreams {
			// Ensure that we are not querying for streams which have finished prior to
			if *s.LastEventTimestamp < from {
				continue
			}

			var newLogs Logs

			resp, err := svc.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
				LogGroupName:  aws.String(group),
				LogStreamName: s.LogStreamName,
				StartFromHead: aws.Bool(true),
				StartTime:     &from,
				EndTime:       &to,
			})
			if err != nil {
				return logs, err
			}

			for _, e := range resp.Events {
				newLogs = append(newLogs, &Log{
					Stream:    *s.LogStreamName,
					Timestamp: time.Unix(*e.Timestamp/1000, 0),
					Message:   *e.Message,
				})
			}

			if len(newLogs) > 0 {
				logs = MergeLogs(logs, newLogs)
			}
		}

		if resp.NextToken == nil {
			done = true
		} else {
			params.NextToken = resp.NextToken
		}
	}

	return logs, nil
}
