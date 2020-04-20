![](jabba.png)

Achuta! Jabba is a json friendly reverse proxy for micro-services, written in golang.

[![Circleci Builds](https://circleci.com/gh/simonmittag/jabba.svg?style=shield)](https://circleci.com/gh/simonmittag/jabba)
[![Github Workflows](https://github.com/simonmittag/jabba/workflows/Go/badge.svg)](https://github.com/simonmittag/jabba/actions)
[![Github Issues](https://img.shields.io/github/issues/simonmittag/jabba)](https://img.shields.io/github/issues/simonmittag/jabba)
[![Github Activity](https://img.shields.io/github/commit-activity/m/simonmittag/jabba)](https://img.shields.io/github/commit-activity/m/simonmittag/jabba)  
[![Codeclimate Maintainability](https://api.codeclimate.com/v1/badges/326d3c304e32645811c8/maintainability)](https://codeclimate.com/github/simonmittag/jabba/maintainability)
[![Codeclimate Test Coverage](https://api.codeclimate.com/v1/badges/326d3c304e32645811c8/test_coverage)](https://codeclimate.com/github/simonmittag/jabba/test_coverage)
[![Go Version](https://img.shields.io/github/go-mod/go-version/simonmittag/jabba)](https://img.shields.io/github/go-mod/go-version/simonmittag/jabba)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Features
* HTTP 1.1 compatible reverse proxy w/ TLS support.
* JSON API friendly, configurable recovery of idempotent requests.
* policy based routing to multiple upstream servers for a/b testing
* highly concurrent with multiprocessor support

# What's new
Jabba is heavily under development and currently alpha grade. 

## 19/04/2020
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
