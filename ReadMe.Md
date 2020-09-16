![](images/j8a.png)

Achuta! j8a [ dʒʌbbʌ ] is a TLS reverse proxy server for JSON APIs written in golang.

[![Circleci Builds](https://circleci.com/gh/simonmittag/j8a.svg?style=shield)](https://circleci.com/gh/simonmittag/j8a)
[![Github Workflows](https://github.com/simonmittag/j8a/workflows/Go/badge.svg)](https://github.com/simonmittag/j8a/actions)
[![Github Issues](https://img.shields.io/github/issues/simonmittag/j8a)](https://github.com/simonmittag/j8a/issues)
[![Github Activity](https://img.shields.io/github/commit-activity/m/simonmittag/j8a)](https://img.shields.io/github/commit-activity/m/simonmittag/j8a)  
[![Codeclimate Maintainability](https://api.codeclimate.com/v1/badges/326d3c304e32645811c8/maintainability)](https://codeclimate.com/github/simonmittag/jabba/maintainability)
[![Codeclimate Test Coverage](https://api.codeclimate.com/v1/badges/326d3c304e32645811c8/test_coverage)](https://codeclimate.com/github/simonmittag/jabba/code)
[![Go Version](https://img.shields.io/github/go-mod/go-version/simonmittag/j8a)](https://img.shields.io/github/go-mod/go-version/simonmittag/j8a)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Features
* HTTP 1.1 compatible, buffering reverse proxy.
* Supports TLS1.2, 1.3 with A+ SSLLabs rating*.
* JSON API friendly, with auto-recovery of idempotent requests.
* policy based routing to multiple upstream servers for a/b testing
* highly concurrent with multiprocessor support, tested to 10k req/s

## What's new
j8a has released alpha version 0.6.2 in August 2020. See project milestones
 for details on general availability.

## Up and running

```bash
go install github.com/simonmittag/j8a/cmd/j8a && j8a
```

## Supported headers

Specify these headers to control server behaviour


### X-REQUEST-DEBUG
Set this to true to include a full trace of the request in the server log

```bash
X-REQUEST-DEBUG: true
```

## Supported environment variables

Specify these env vars to control server behaviour


### LOGLEVEL
Loglevel for stdout, one of TRACE, DEBUG, INFO or WARN

```bash
LOGLEVEL="DEBUG"
```

### TZ
TimeZone

```bash
TZ="Australia/Sydney"
```

## Contributions
The j8a team welcomes all [contributors](https://github.com/simonmittag/j8a/blob/master/CONTRIBUTING.md). Everyone interacting with the project's codebase, issue trackers, chat rooms and mailing lists
is expected to follow the [code of conduct](https://github.com/simonmittag/j8a/blob/master/CODE_OF_CONDUCT.md)
