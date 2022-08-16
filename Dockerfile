FROM debian:buster

## Make apt-get non-interactive
ENV DEBIAN_FRONTEND=noninteractive

RUN echo "deb http://deb.debian.org/debian buster-backports main" >> /etc/apt/sources.list

RUN apt-get update \
 && apt-get install -y \
      vim \
      redis-server \
      supervisor \
      curl \
      bridge-utils \
      net-tools \
      libboost-serialization1.71-dev \
      libzmq5-dev

COPY debs /debs
RUN dpkg -i /debs/*.deb
RUN rm -fr /debs

COPY supervisor/supervisor.conf /etc/supervisor/conf.d/
COPY supervisor/rest_api.conf /etc/supervisor/conf.d/

RUN mkdir /usr/sbin/cert
RUN mkdir /usr/sbin/cert/client
RUN mkdir /usr/sbin/cert/server
COPY cert/client/* /usr/sbin/cert/client/
COPY cert/server/* /usr/sbin/cert/server/

RUN apt-get autoremove -y \
 && apt-get clean \
 && rm -fr /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENTRYPOINT ["/usr/bin/supervisord"]
