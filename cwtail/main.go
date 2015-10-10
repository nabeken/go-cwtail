package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/jessevdk/go-flags"
)

const schema = "logs"

var (
	errInvalidSchema = errors.New("cwtail: schema must be `logs`")
	errInvalidURL    = errors.New("cwtail: invalid URL specified")
)

var opts struct {
	Follow   bool          `short:"f" description:"wait for additional data to be appended to the log stream"`
	Number   int64         `short:"n" default:"20" description:"the number of logs to fetch"`
	Interval time.Duration `short:"d" default:"1s" description:"interval for polling the log streams"`
}

func AWSConfig() *aws.Config {
	return defaults.DefaultConfig
}

func parseArgs(args []string) ([]*url.URL, error) {
	urls := make([]*url.URL, 0, len(args))
	for _, arg := range args {
		u, err := url.Parse(arg)
		if err != nil {
			return nil, err
		}
		if u.Scheme != schema {
			return nil, errInvalidSchema
		}
		if u.Host == "" || u.Path == "" || u.Path == "/" {
			return nil, errInvalidURL
		}
		u.Path = strings.TrimPrefix(u.Path, "/")
		urls = append(urls, u)
	}
	return urls, nil
}

func main() {
	os.Exit(realMain())
}

type Poller struct {
	logs     cloudwatchlogsiface.CloudWatchLogsAPI
	interval time.Duration
	dest     io.Writer
	limit    int64
}

func (p *Poller) Fetch(groupName, streamName string) (*cloudwatchlogs.GetLogEventsOutput, error) {
	req := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(groupName),
		LogStreamName: aws.String(streamName),
		Limit:         aws.Int64(p.limit),
	}
	return p.logs.GetLogEvents(req)
}

func (p *Poller) FetchNext(groupName, streamName, nextToken string) (*cloudwatchlogs.GetLogEventsOutput, error) {
	req := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(groupName),
		LogStreamName: aws.String(streamName),
	}

	if nextToken != "" {
		req.NextToken = aws.String(nextToken)
	}

	return p.logs.GetLogEvents(req)
}

func (p *Poller) PrintEvents(events []*cloudwatchlogs.OutputLogEvent) {
	for _, e := range events {
		fmt.Fprintln(p.dest, *e.Message)
	}
}

func (p *Poller) Poll(groupName, streamName string) {
	var nextToken string
	for range time.Tick(p.interval) {
		var resp *cloudwatchlogs.GetLogEventsOutput
		var err error

		// For the first time, we should limit a number of logs to print
		if nextToken == "" {
			resp, err = p.Fetch(groupName, streamName)
		} else {
			resp, err = p.FetchNext(groupName, streamName, nextToken)
		}

		if err != nil {
			// Ignore ResourceNotFoundException
			if logserr, ok := err.(awserr.Error); ok && logserr.Code() == "ResourceNotFoundException" {
				continue
			}
			log.Println(err)
			continue
		}
		p.PrintEvents(resp.Events)
		nextToken = *resp.NextForwardToken
	}
}

func realMain() int {
	args, err := flags.Parse(&opts)
	if err != nil {
		return 1
	}

	if len(args) < 1 {
		log.Println("Please specify at-least one logs")
		return 1
	}

	urls, err := parseArgs(args)
	if err != nil {
		log.Println(err)
		return 1
	}

	logs := cloudwatchlogs.New(AWSConfig())

	poller := &Poller{
		logs:     logs,
		interval: opts.Interval,
		dest:     os.Stdout,
		limit:    opts.Number,
	}

	for _, u := range urls {
		groupName := u.Host
		streamName := u.Path

		if opts.Follow {
			go poller.Poll(groupName, streamName)
		} else {
			resp, err := poller.Fetch(groupName, streamName)
			if err != nil {
				log.Println(err)
				return 1
			}
			poller.PrintEvents(resp.Events)
		}
	}

	// Wait forever if we follow streams
	if opts.Follow {
		select {}
	}

	return 0
}
