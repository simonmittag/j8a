[![Circleci Builds](https://circleci.com/gh/simonmittag/j8a.svg?style=shield)](https://circleci.com/gh/simonmittag/j8a)
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
[![Version](https://img.shields.io/docker/v/simonmittag/j8a/v1.1.2)](https://hub.docker.com/repository/docker/simonmittag/j8a)

# What is j8a?
[j8a [ dʒʌbbʌ ]](https://j8a.io) is a modern **TLS 1.3 reverse proxy server** designed as a lightweight API gateway for
REST APIs.

![](images/route.gif)

## What's new?

### v1.1.3 (23/04/2025)
* upgraded to `github.com/lestrrat-go/jwx/v3`, retired older dependencies.

### v1.1.2 (17/04/2025)
* Update to Go 1.24
* Update to goreleaser 2.8 and multiplatform docker build
* Bumped several dependencies

## CVEs
🛡️ J8a v1.1.3 has [zero](https://www.cvedetails.com/google-search-results.php?q=j8a&sa=Search) reported CVEs as of 23/04/2025. 
If you are a security researcher, the project team would like to [hear from you](https://github.com/simonmittag/j8a/issues).

### Trivy
```aiignore
λ trivy image simonmittag/j8a:v1.1.3
2025-04-23T01:09:26+10:00       INFO    [gobinary] Detecting vulnerabilities...

Report Summary
┌────────────────────────────────────────┬──────────┬─────────────────┬─────────┐
│                 Target                 │   Type   │ Vulnerabilities │ Secrets │
├────────────────────────────────────────┼──────────┼─────────────────┼─────────┤
│ simonmittag/j8a:v1.1.3 (alpine 3.21.3) │  alpine  │        0        │    -    │
├────────────────────────────────────────┼──────────┼─────────────────┼─────────┤
│ j8a                                    │ gobinary │        0        │    -    │
└────────────────────────────────────────┴──────────┴─────────────────┴─────────┘
```

### Docker Scout
```aiignore
λ docker scout cves simonmittag/j8a:v1.1.3
    ✓ Indexed 56 packages
    ✓ No vulnerable package detected

                    │       Analyzed Image         
────────────────────┼──────────────────────────────
  Target            │  simonmittag/j8a:v1.1.3      
    digest          │  54aae504c670                
    platform        │ linux/arm64                  
    vulnerabilities │    0C     0H     0M     0L   
    size            │ 9.2 MB                       
    packages        │ 56                         
```

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
go install github.com/simonmittag/j8a/cmd/j8a@latest &&
  j8a
```

## Usage
J8a runs from the cli. The process will attempt to load its configuration from the environment 
variable `J8ACFG_YML`, a specified config file, or the default config file `j8acfg.yml` (in that order). 
You can validate the config using the `-o` flag without actually starting the server, which 
is useful for ci/cd pipelines. In validate mode, the process successfully exits with `0`, or `-1` in case of failure.

### Commandline Options

```bash
λ j8a -h
j8a[v1.0.1] Goodde da lodia!
Usage: j8a [-c] [-o] | [-v] | [-h]
  -c string
        config file location (default "j8acfg.yml").
  -h    print usage instructions.
  -o    validate config file, then exit.
  -v    print version.
```

### Examples

Validate Server Configuration supplied as `myconfig.yml`
```bash
λ j8a -c myconfig.yml -o
6:25PM INF hostName determined hostName=MacBook-Pro-16.local
6:25PM INF version determined version=v1.0.1
6:25PM INF srvID determined srvID=c91cda8c
...
```

Start Server With Configuration supplied as `myconfig.yml`
```bash
λ j8a -c myconfig.yml 
6:25PM INF hostName determined hostName=MacBook-Pro-16.local
6:25PM INF version determined version=v1.0.1
6:25PM INF srvID determined srvID=c91cda8c
...
```

Start Server With Configuration supplied via `J8ACFG_YML`
```bash
J8ACFG_YML="---
            connection:
              downstream:
                http:
                  port: 80
            routes:
              - path: "/todos"
                resource: jsonplaceholder
            resources:
              jsonplaceholder:
                - url:
                    scheme: https
                    host: jsonplaceholder.typicode.com
                    port: 443" j8a
```

## Contributions

The j8a team is looking for and welcomes all [contributors](https://github.com/simonmittag/j8a/blob/master/CONTRIBUTING.md). Everyone
interacting with the project's codebase, issue trackers, chat rooms and mailing lists is expected to follow
the [code of conduct](https://github.com/simonmittag/j8a/blob/master/CODE_OF_CONDUCT.md)

## Release Cycle
* Now that 1.0 is reached, we usually release 3/4 minor versions per year.
* Bug-fixes (e.g. 1.0.1, 1.0.2) are released as needed (no additional features are delivered in those versions, bug-fixes only).
* Each version is supported until the next one is released (e.g. 1.1.x will be supported until 1.2.0 is out).

We use [Semantic Versioning](https://semver.org/).
