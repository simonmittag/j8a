# This is multiarch including linux/amd64 and linux/arm64/v8, see: docker buildx imagetools inspect golang:1.21.4-alpine
FROM golang:1.21.4-alpine AS build

RUN apk update && apk upgrade && apk add --no-cache bash git

WORKDIR .
COPY . /proj/

# now build from source on each platform this Dockerfile gets executed
WORKDIR /proj
RUN CGO_ENABLED=0 go build github.com/simonmittag/j8a/cmd/j8a

# multistage build uses output from previous image
# base image for distro is also multiarch, see: docker buildx imagetools inspect alpine
FROM alpine
COPY --from=build /proj/j8a /j8a

ENV LOGLEVEL="DEBUG"

EXPOSE 80
EXPOSE 443
ENTRYPOINT "/j8a"