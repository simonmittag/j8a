[![Circleci Builds](https://circleci.com/gh/simonmittag/j8a.svg?style=shield)](https://circleci.com/gh/simonmittag/j8a)
[![Github Workflows](https://github.com/simonmittag/j8a/workflows/Go/badge.svg)](https://github.com/simonmittag/j8a/actions)
[![Dependabot](https://badgen.net/badge/Dependabot/enabled/green?icon=dependabot)](https://github.com/simonmittag/j8a/pulls?q=is%3Aopen+is%3Apr)
[![Github Issues](https://img.shields.io/github/issues/simonmittag/j8a)](https://github.com/simonmittag/j8a/issues)
[![Github Activity](https://img.shields.io/github/commit-activity/m/simonmittag/j8a)](https://img.shields.io/github/commit-activity/m/simonmittag/j8a)  
[![Codeclimate Maintainability](https://api.codeclimate.com/v1/badges/326d3c304e32645811c8/maintainability)](https://codeclimate.com/github/simonmittag/jabba/maintainability)
[![Codeclimate Test Coverage](https://api.codeclimate.com/v1/badges/326d3c304e32645811c8/test_coverage)](https://codeclimate.com/github/simonmittag/jabba/code)
[![SSLLabs Rating](https://img.shields.io/badge/SSLLabs-A%2B-brightgreen)](https://www.ssllabs.com/ssltest/analyze.html?d=j8a.io)
[![Go Version](https://img.shields.io/github/go-mod/go-version/simonmittag/j8a)](https://img.shields.io/github/go-mod/go-version/simonmittag/j8a)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Docker Pulls](https://img.shields.io/docker/pulls/simonmittag/j8a)](https://hub.docker.com/repository/docker/simonmittag/j8a)
[![Version](https://img.shields.io/docker/v/simonmittag/j8a/0.8.1)](https://hub.docker.com/repository/docker/simonmittag/j8a)

## What's new?
### v0.8.1
* upgraded to lestrrat-go/jwx@1.2.1
* upgraded to gojq@0.12.4
* upgraded to Alpine/Go@v1.16.4
* minor refactoring, new log vars.

## What is j8a?
[j8a [ dʒʌbbʌ ]](https://j8a.io) is a modern **TLS 1.3 reverse proxy server** designed as a lightweight API gateway for 
micro-service style JSON APIs. It is built for DevOps engineers and allows fine-grained traceability of HTTP and 
Websocket traffic without sacrificing performance. 

![](images/route.gif)

## Features
* Fast! [20k req/s](https://app.circleci.com/pipelines/github/simonmittag/j8a/606/workflows/b8016dcf-0a2f-447f-9297-50b7737990a1/jobs/1992) HTTPS traffic on single CPU core.
* Secure. TLS1.2, TLS/1.3 termination with zero config [A+ SSLLabs rating](https://www.ssllabs.com/ssltest/analyze.html?d=j8a.io). No weak ciphers ever again.
* Observable. API request tracing w/ correlation ID for upstream micro-services.
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

### Linux / Darwin
```bash
go install github.com/simonmittag/j8a/cmd/j8a && j8a
```

## Configuration
### Environment Variables

#### J8ACFG_YML
Supply the server configuration to the running process in .yml format

```bash
J8ACFG_YML="---
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

#### LOGLEVEL
Loglevel for stdout, one of TRACE, DEBUG, INFO or WARN

```bash
LOGLEVEL="DEBUG"
```

#### TZ
TimeZone

```bash
TZ="Australia/Sydney"
```

## Contributions
The j8a team welcomes all [contributors](https://github.com/simonmittag/j8a/blob/master/CONTRIBUTING.md). Everyone interacting with the project's codebase, issue trackers, chat rooms and mailing lists
is expected to follow the [code of conduct](https://github.com/simonmittag/j8a/blob/master/CODE_OF_CONDUCT.md)
