FROM ubuntu:22.04

RUN apt-get -y update
RUN apt-get -y install build-essential curl rubygems

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]
