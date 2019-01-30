FROM golang:1.10

RUN echo 'nameserver 8.8.8.8' >> /etc/resolv.conf
#RUN wget http://apt.ntop.org/stretch/all/apt-ntop-stable.deb
#RUN dpkg -i apt-ntop-stable.deb
#RUN apt-get update && apt-get update -y && apt-get install pfring
RUN apt-get update && apt-cache search libpcap && apt-get install libpcap-dev libpcap0.8 libpcap0.8-dev
#nprobe ntopng ntopng-data n2disk cento nbox

WORKDIR /go/src/relay
COPY . .

#RUN /go/src/gotcpDump/installpfring.sh
#RUN /go/src/gotcpDump/pcapinstall.sh

RUN go install

ENV hubip=172.17.0.3

ENTRYPOINT ["/go/src/relay/start.sh"]

