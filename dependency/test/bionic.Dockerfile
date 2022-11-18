FROM ubuntu:18.04

RUN apt update && apt install -y ruby-full rubygems

# upgrade to a version of rubygems that is compatible with bundler 2
RUN gem update --system --verbose

COPY Gemfile /test/Gemfile

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]

WORKDIR /test
