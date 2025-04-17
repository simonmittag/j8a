# Minimal runtime image
FROM alpine:latest

WORKDIR /

COPY j8a /j8a

ENV LOGLEVEL="DEBUG"

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/j8a"]
