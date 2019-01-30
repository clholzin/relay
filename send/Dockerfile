FROM golang:1.10
#docker run --privileged send &

RUN apt-get update && apt-get update -y && apt-get install -y iptables && iptables -h

WORKDIR /go/src/send
COPY . .

#RUN go get -d -v ./...
RUN go install

EXPOSE 8222
EXPOSE 8221

ENV conntype=send \
hostip=172.17.0.4

ENTRYPOINT ["/go/src/send/start.sh"]
