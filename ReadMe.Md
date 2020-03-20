![](jabba.png)

Achuta! Jabba is a json friendly reverse proxy for micro-services, written in golang.

## Features
* HTTP 1.1, native TLS 1.2 support with A+ rating
* map URLs to upstream HTTP and TLS based micro-services, inc. policy based routing for a/b testing
* highly concurrent with multiprocessor support



# What's new
## 20/03/2020
Jabba is partially complete, under development, unreleased, alpha grade software.

# Up and running

```bash
./jabba
```


# Supported environment variables

Specify these env vars to control server behaviour

## VERSION
Server version for binary distribution

```bash
VERSION="alpha"
```

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

