FROM ubuntu:22.04

RUN apt update && apt install -y ruby-full rubygems

COPY Gemfile /test/Gemfile

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]

WORKDIR /test
