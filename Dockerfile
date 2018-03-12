# Sample build via docker cli:
# docker build --build-arg cores=8 -t blocknetdx/dxregress:1.0.0 .
# docker run --name dxregress -v /var/run/docker.sock:/var/run/docker.sock blocknetdx/dxregress:1.0.0
FROM ubuntu:trusty

ARG cores=32
ENV ecores=$cores

# Packages
RUN apt update \
  && apt install -y --no-install-recommends \
     software-properties-common \
     ca-certificates apt-transport-https \
     wget curl git python vim \
  && add-apt-repository ppa:gophers/archive && apt update \
  && apt install -y --no-install-recommends \
     golang-1.9-go \
  && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# Install go 1.9
RUN mkdir -p /opt/src && mkdir -p /opt/bin \
  && export PATH=$PATH:/usr/lib/go-1.9/bin:/opt/bin \
  && export GOPATH=/opt && export GOROOT=/usr/lib/go-1.9 \
  && go get -u github.com/BlocknetDX/dxregress \
  && cd /opt/src/github.com/BlocknetDX/dxregress \
  && go install

# Install docker 1.11.2
RUN apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D \
  && echo "deb https://apt.dockerproject.org/repo ubuntu-trusty main" > /etc/apt/sources.list.d/docker.list \
  && apt update && apt install -y docker-engine=1.11.2-0~trusty \
  && apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV PATH=${PATH}:/usr/lib/go-1.9/bin:/opt/bin

CMD ["dxregress"]