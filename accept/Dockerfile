FROM golang:1.10
#docker run --privileged -p 8222:8222 accept &
#docker run --dns-option 8.8.8.8 --privileged accept

#RUN echo 'Acquire::http::Proxy "http://proxy.aexp.com:8080";' >> /etc/apt/apt.conf
RUN echo 'nameserver 8.8.8.8' >> /etc/resolv.conf
RUN apt-get update && apt-get update -y && apt-get install -y iptables rsyslog tcpdump && iptables -h

WORKDIR /go/src/accept
COPY . .

#RUN go get -d -v ./...
RUN go install

EXPOSE 8222
EXPOSE 8221

ENV conntype=listen

ENTRYPOINT ["/go/src/accept/start.sh"]

