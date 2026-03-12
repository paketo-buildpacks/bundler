FROM ubuntu:22.04
ARG RUBY_VERSION=3.2.10

RUN apt-get -y update \
 && apt-get -y install --no-install-recommends \
	 autoconf \
	 bison \
	 build-essential \
	 ca-certificates \
	 curl \
	 libdb-dev \
	 libffi-dev \
	 libgdbm-compat-dev \
	 libgdbm-dev \
	 libncurses5-dev \
	 libreadline-dev \
	 libssl-dev \
	 libtool \
	 libyaml-dev \
	 pkg-config \
	 xz-utils \
	 zlib1g-dev \
 && curl -fsSL "https://cache.ruby-lang.org/pub/ruby/3.2/ruby-${RUBY_VERSION}.tar.xz" -o /tmp/ruby.tar.xz \
 && mkdir -p /tmp/ruby-src \
 && tar -xf /tmp/ruby.tar.xz -C /tmp/ruby-src --strip-components=1 \
 && cd /tmp/ruby-src \
 && ./configure --disable-install-doc \
 && make -j"$(nproc)" \
 && make install \
 && ruby --version \
 && gem --version \
 && rm -rf /tmp/ruby.tar.xz /tmp/ruby-src \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/*

COPY Gemfile /test/Gemfile

COPY entrypoint /entrypoint

ENTRYPOINT ["/entrypoint"]

WORKDIR /test
