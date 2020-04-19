![](jabba.png)

Achuta! Jabba is a json friendly reverse proxy for micro-services, written in golang.

[![circleci](https://circleci.com/gh/simonmittag/jabba.svg?style=shield)](https://circleci.com/gh/simonmittag/jabba)

## Features
* HTTP 1.1 compatible reverse proxy w/ TLS support.
* JSON API friendly, configurable recovery of idempotent requests.
* policy based routing to multiple upstream servers for a/b testing
* highly concurrent with multiprocessor support

# What's new
Jabba is heavily under development and currently alpha grade. 

## 19/04/20
* converted to go mod
* added circleci support

## 01/04/2020
* HTTP proxying is mostly complete and in testing.
* TLS is still missing

# Up and running

```bash
go install github.com/simonmittag/jabba/cmd/jabba && jabba
```

# Supported environment variables

Specify these env vars to control server behaviour


## LOGLEVEL
Loglevel for stdout, one of TRACE, DEBUG, INFO or WARN

```bash
LOGLEVEL="DEBUG"
```

## LOGCOLOR
Enable color logging on stdout for development or workstation environment (not recommended in production)

```bash
LOGCOLOR="true"
```

## TZ
TimeZone

```bash
TZ="Australia/Sydney"
```
