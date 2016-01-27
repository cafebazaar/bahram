FROM busybox:ubuntu-14.04

MAINTAINER Reza Mohammadi "<reza@cafebazaar.ir>"

ENTRYPOINT ["/app/bahram"]
WORKDIR /app

COPY bahram /app/
