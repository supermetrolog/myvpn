FROM golang:1.21 as base

FROM base as dev

RUN apt update
RUN apt install traceroute
RUN apt install net-tools
RUN apt install iproute2 -y
RUN apt install iptables -y
RUN apt install tcpdump -y
RUN apt install iputils-ping -y

RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

WORKDIR /opt/app
CMD ["air"]