package cloudwatchlogs

import (
	"math"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// QueryParams which get passed to the Query function.
type QueryParams struct {
	Group  string
	Prefix string
	Start  int64
	End    int64
}

// QueryOutput which gets returned fromm Query function.
type QueryOutput struct {
	Logs []Log
}

// Log which contains messages from streams.
type Log struct {
	Stream    string
	Timestamp int64
	Message   string
}

// Query for logs.
func Query(client *cloudwatchlogs.CloudWatchLogs, params QueryParams) (QueryOutput, error) {
	var output QueryOutput

	streams, err := getStreams(client, params.Group, params.Prefix, params.Start)
	if err != nil {
		return output, err
	}

	if len(streams) == 0 {
		return output, nil
	}

	events, err := getLogs(client, params.Group, streams, params.Start, params.End)
	if err != nil {
		return output, err
	}

	for _, event := range events {
		output.Logs = append(output.Logs, Log{
			Stream:    *event.LogStreamName,
			Timestamp: *event.Timestamp,
			Message:   *event.Message,
		})
	}

	return output, nil
}

// Helper function to get streams.
func getStreams(client *cloudwatchlogs.CloudWatchLogs, group, prefix string, start int64) ([]*string, error) {
	var streams []*string

	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(group),
		Descending:   aws.Bool(true),
		OrderBy:      aws.String(cloudwatchlogs.OrderByLastEventTime),
	}

	for {
		resp, err := client.DescribeLogStreams(params)
		if err != nil {
			return streams, err
		}

		for _, stream := range resp.LogStreams {
			// Ensure only logstreams with specified prefix are included.
			if prefix != "" && strings.Index(*stream.LogStreamName, prefix) != 0 {
				continue
			}

			if *stream.LastEventTimestamp < start {
				return streams, nil
			}

			streams = append(streams, stream.LogStreamName)
		}

		if resp.NextToken == nil {
			return streams, nil
		}

		params.NextToken = resp.NextToken
	}
}

type safeEvents struct {
	events []*cloudwatchlogs.FilteredLogEvent
	mux    sync.Mutex
}

// Append adds events to the store.
func(s *safeEvents) Append(events []*cloudwatchlogs.FilteredLogEvent) {
	s.mux.Lock()
	defer s.mux.Unlock()

	for _, v := range events {
		s.events = append(s.events, v)
	}
}

// Get returns the events.
func(s *safeEvents) Get() []*cloudwatchlogs.FilteredLogEvent {
	return s.events
}

// Helper function to get logs.
func getLogs(client *cloudwatchlogs.CloudWatchLogs, group string, streams []*string, start, end int64) ([]*cloudwatchlogs.FilteredLogEvent, error) {
	// Cloudwatch has a maximum of 100 logstreams per api request.
	var wg sync.WaitGroup
	requestPools := int(math.Ceil(float64(len(streams)) / 100))
	wg.Add(requestPools)

	events := safeEvents{}
	for i := 0; i < requestPools; i++ {
		min := i
		max := i + 100
		var poolStreams = []*string{}
		for ii := min; ii < max; ii++ {
			if len(streams) <= ii { break }
			poolStreams = append(poolStreams, streams[ii])
		}
		go func(streams []*string) {
			defer wg.Done()

			params := &cloudwatchlogs.FilterLogEventsInput{
				LogGroupName:   aws.String(group),
				LogStreamNames: streams,
				StartTime:      &start,
				EndTime:        &end,
				Interleaved:    aws.Bool(true),
			}

			for {
				resp, err := client.FilterLogEvents(params)
				if err != nil {
					return
				}

				events.Append(resp.Events)

				if resp.NextToken == nil {
					return
				}

				params.NextToken = resp.NextToken
			}
		}(poolStreams)
	}

	wg.Wait()
	return events.Get(), nil
}
