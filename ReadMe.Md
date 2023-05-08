[![Circleci Builds](https://circleci.com/gh/simonmittag/j8a.svg?style=shield)](https://circleci.com/gh/simonmittag/j8a)
[![Github Workflows](https://github.com/simonmittag/j8a/workflows/Go/badge.svg)](https://github.com/simonmittag/j8a/actions)
[![Dependabot](https://badgen.net/badge/Dependabot/enabled/green?icon=dependabot)](https://github.com/simonmittag/j8a/pulls?q=is%3Aopen+is%3Apr)
[![Github Issues](https://img.shields.io/github/issues/simonmittag/j8a)](https://github.com/simonmittag/j8a/issues)
[![Github Activity](https://img.shields.io/github/commit-activity/m/simonmittag/j8a)](https://img.shields.io/github/commit-activity/m/simonmittag/j8a)  
[![Go Report](https://goreportcard.com/badge/github.com/simonmittag/j8a)](https://goreportcard.com/report/github.com/simonmittag/j8a)
[![CodeClimate Maintainability](https://api.codeclimate.com/v1/badges/326d3c304e32645811c8/maintainability)](https://codeclimate.com/github/simonmittag/jabba/maintainability)
[![CodeClimate Test Coverage](https://api.codeclimate.com/v1/badges/326d3c304e32645811c8/test_coverage)](https://codeclimate.com/github/simonmittag/jabba/code)
[![SSLLabs Rating](https://img.shields.io/badge/SSLLabs-A%2B-brightgreen)](https://www.ssllabs.com/ssltest/analyze.html?d=j8a.io)
[![Go Version](https://img.shields.io/github/go-mod/go-version/simonmittag/j8a)](https://img.shields.io/github/go-mod/go-version/simonmittag/j8a)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Docker Pulls](https://img.shields.io/docker/pulls/simonmittag/j8a)](https://hub.docker.com/repository/docker/simonmittag/j8a)
[![Version](https://img.shields.io/docker/v/simonmittag/j8a/1.0.0)](https://hub.docker.com/repository/docker/simonmittag/j8a)

## What's new?

### v1.0.0 (07/05/2023)
* support for virtual host based routing inc subdomains.
* timeZone and logLevel are now options in config.yml as opposed to env variables.
* re-introduced support for HTTP PATCH method.
* OPTIONS requests to '*' now list legal server HTTP methods
* update to go 1.20, and multiple dependency updates

### v0.9.7 (10/11/2022)
* update to go 1.19

### v0.9.6 (01/08/2022)
* allows go `{{template}}` variables in config file to substitute env vars.
* new params for acmeEmail, gracePeriodDays to configure letsencrypt provider
* new param for disableXRequestInfo to disable debug mode

## What is j8a?

[j8a [ dʒʌbbʌ ]](https://j8a.io) is a modern **TLS 1.3 reverse proxy server** designed as a lightweight API gateway for
REST APIs. 

![](images/route.gif)

## Features

* Fast! 5k POST req/s traffic.
* Secure. TLS1.2, TLS/1.3 termination w/ [A+ SSLLabs rating](https://www.ssllabs.com/ssltest/analyze.html?d=j8a.io).
* Zero downtime. Auto-renew certs using ACME [RFC 8555](https://tools.ietf.org/html/rfc8555).
* Observable. API request tracing w/ correlation ID for upstream microservices.
* APM. CPU, memory logging built-in. Daily TLS health and validity check for your full certificate chain.
* JWT token validation with full JWK(S) support for [RFC 7519](https://tools.ietf.org/html/rfc7519).
* HTTP/1.1 and HTTP/2 support w/ upstream <> downstream protocol translation.
* Websocket Support for [RFC 6455](https://datatracker.ietf.org/doc/html/rfc6455)
* [Docker](https://hub.docker.com/repository/docker/simonmittag/j8a) native

## Up and running

### Docker
```
docker pull simonmittag/j8a &&
  docker run -e J8ACFG_YML -p443:443 simonmittag/j8a
```

### Homebrew
```bash
brew tap simonmittag/cli && 
  brew install j8a && 
  j8a
```

### Golang
```bash
go install github.com/simonmittag/j8a/cmd/j8a &&
  j8a
```

## Configuration

### J8ACFG_YML

Supply the server configuration to the running process in .yml format

```bash
J8ACFG_YML="---
            timeZone: UTC
            logLevel: INFO
            connection:
              downstream:
                readTimeoutSeconds: 3
                roundTripTimeoutSeconds: 20
                idleTimeoutSeconds: 30
                http:
                  port: 80
                maxBodyBytes: 65535
              upstream:
                socketTimeoutSeconds: 3
                readTimeoutSeconds: 30
                idleTimeoutSeconds: 10
                maxAttempts: 4
                poolSize: 8
                tlsInsecureSkipVerify: true
            routes:
              - path: "/todos"
                resource: jsonplaceholder
            resources:
              jsonplaceholder:
                - url:
                    scheme: https
                    host: jsonplaceholder.typicode.com
                    port: 443"
```

## Contributions

The j8a team welcomes all [contributors](https://github.com/simonmittag/j8a/blob/master/CONTRIBUTING.md). Everyone
interacting with the project's codebase, issue trackers, chat rooms and mailing lists is expected to follow
the [code of conduct](https://github.com/simonmittag/j8a/blob/master/CODE_OF_CONDUCT.md)
