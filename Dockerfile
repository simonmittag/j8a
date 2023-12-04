# This is multiarch including linux/amd64 and linux/arm64/v8, see: docker buildx imagetools inspect golang:1.21.4-alpine
FROM alpine
RUN apk update && apk upgrade && apk add --no-cache bash git

WORKDIR .
COPY j8a /j8a

ENV LOGLEVEL="DEBUG"

EXPOSE 80
EXPOSE 443
ENTRYPOINT "/j8a"