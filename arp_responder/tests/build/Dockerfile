FROM ubuntu:latest

MAINTAINER Pavel Shirshov

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get upgrade -y  \
    && apt-get install -y --no-install-recommends \
        openssh-server \
        vim \
        python \
        python-pip \
        python-setuptools \
        supervisor \
        traceroute \
        lsof \
        tcpdump \
        libssl-dev \
        less \
        libboost-all-dev \
        g++ \
        wget \
        make \
    && apt-get -y install -f \
    && rm -rf /root/deps \
    && apt-get -y autoclean \
    && apt-get -y autoremove

RUN mkdir /var/run/sshd

RUN echo 'root:123456!' | chpasswd

ENV LANG C.UTF-8
ENV LANGUAGE C.UTF-8
ENV LC_ALL C.UTF-8

RUN sed -ri 's/^PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config
RUN sed -i '$aUseDNS no' /etc/ssh/sshd_config

EXPOSE 22

ADD supervisord.conf /etc/supervisor/conf.d/supervisord.conf

RUN wget http://archive.apache.org/dist/thrift/0.9.3/thrift-0.9.3.tar.gz \
 && tar xvfz thrift-0.9.3.tar.gz \
 && cd thrift-0.9.3 \
 && ./configure --prefix=/usr \
 && make \
 && make install \
 && cd .. \
 && rm -fr thrift-0.9.3 \
 && rm -f thrift-0.9.3.tar.gz

ENTRYPOINT ["/usr/bin/supervisord"]
