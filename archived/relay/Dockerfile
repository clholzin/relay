FROM golang:1.10
#docker run --name relay relay &

#RUN echo 'Acquire::http::Proxy "http://localhost";' >> /etc/apt/apt.conf
RUN echo 'nameserver 8.8.8.8' >> /etc/resolv.conf
RUN apt-get update && apt-get update -y && apt-get install -y tcpdump

WORKDIR /go/src/relay
COPY . .

RUN go install
#EXPOSE 8222
#EXPOSE 8221

ENV hubip=172.17.0.3

ENTRYPOINT ["/go/src/relay/start.sh"]

