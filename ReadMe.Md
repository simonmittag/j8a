![](jabba.png)

Achuta! Jabba is a json friendly reverse proxy for micro-services.

## Features
* map URLs to upstream HTTP and TLS based micro-services 
* policy based routing for a/b testing
* HTTP 1.1, native TLS 1.2 support with A+ rating
* GZIP auto-encoding in both directions


# What's new
## 20/03/2019
* partially complete and under development

# Up and running

```bash
./jabba
```


# Supported environment variables

Specify these env vars to control server behaviour

## VERSION
Server version for binary distribution

```bash
VERSION="v.0.1.1"
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

