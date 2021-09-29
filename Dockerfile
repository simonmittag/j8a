FROM golang:1.17-alpine AS build

RUN apk update && apk upgrade && apk add --no-cache bash git

WORKDIR .
COPY . /proj/

RUN /bin/bash
RUN cd /proj && CGO_ENABLED=0 go build github.com/simonmittag/j8a/cmd/j8a

#multistage build uses output from previous image
FROM alpine
COPY --from=build /proj/j8a /j8a

ENV LOGLEVEL="DEBUG"

EXPOSE 80
EXPOSE 443
ENTRYPOINT "/j8a"