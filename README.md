# go-cwtail

[![Build Status](https://travis-ci.org/nabeken/go-cwtail.svg)](https://travis-ci.org/nabeken/go-cwtail)

`tail` command for [CloudWatch Logs](http://aws.amazon.com/blogs/aws/cloudwatch-log-service/).

# Installation

Download from [releases](https://github.com/nabeken/go-cwtail/releases).

Or

```sh
go get -u github.com/nabeken/go-cwtail/cwtail
```

# Usage

```sh
Usage:
  cwtail [OPTIONS]

Application Options:
  -f          wait for additional data to be appended to the log stream
  -n=         the number of logs to fetch (20)
  -d=         interval for polling the log streams (1s)

Help Options:
  -h, --help  Show this help message
```

You should setup a credential for AWS SDK.
Even if you use the instance-profile, you must at-least set `AWS_REGION` variable.

```sh
export AWS_REGION=ap-northeast-1
```

See [SDK documentation](http://docs.aws.amazon.com/sdk-for-go/api/) for more details.

*cwtail* can tail multiple streams like this:

```sh
cwtail -f \
  logs://log-group/log-stream-1
  logs://log-group/log-stream-2
```
