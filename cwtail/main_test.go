package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArg(t *testing.T) {
	assert := assert.New(t)

	for _, tc := range []struct {
		Arg       string
		Expect    *LogStreamLocation
		ExpectErr error
	}{
		{
			Arg: "log-group@log-stream1",
			Expect: &LogStreamLocation{
				GroupName:  "log-group",
				StreamName: "log-stream1",
			},
		},
		{
			Arg: "log-group@stream@log-stream1",
			Expect: &LogStreamLocation{
				GroupName:  "log-group",
				StreamName: "stream@log-stream1",
			},
		},
		{
			Arg: "log-group@@",
			Expect: &LogStreamLocation{
				GroupName:  "log-group",
				StreamName: "@",
			},
		},
		{
			Arg:       "log-group",
			ExpectErr: errInvalidLogStreamLocation,
		},
		{
			Arg:       "log-group@",
			ExpectErr: errInvalidLogStreamLocation,
		},
	} {
		loc, err := ParseArg(tc.Arg)
		if tc.ExpectErr != nil {
			assert.Equal(tc.ExpectErr, err)
		} else {
			assert.Equal(tc.Expect.GroupName, loc.GroupName)
			assert.Equal(tc.Expect.StreamName, loc.StreamName)
		}
	}
}
