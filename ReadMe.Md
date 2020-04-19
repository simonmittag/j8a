[![circleci](https://circleci.com/gh/simonmittag/jabba.svg?style=svg)](https://circleci.com/gh/simonmittag/jabba)

![](jabba.png)

Achuta! Jabba is a json friendly reverse proxy for micro-services, written in golang.

## Features
* HTTP 1.1 server
* map URLs to upstream HTTP and TLS based micro-services, inc. policy based routing for a/b testing
* highly concurrent with multiprocessor support



# What's new
Jabba is heavily under development and currently alpha grade. 

## 01/04/2020
* HTTP proxying is mostly complete and in testing.
* TLS is still missing

# Up and running

```bash
./jabba
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
