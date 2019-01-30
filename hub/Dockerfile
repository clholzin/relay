FROM golang:1.10
# docker run --privileged -p 80:80 -p 9999:9999 -p 443:443 --name hub hub
#RUN echo 'Acquire::http::Proxy "http://proxy.aexp.com:8080";' >> /etc/apt/apt.conf
#RUN echo 'nameserver 8.8.8.8' >> /etc/resolv.conf

WORKDIR /go/src/hub
COPY . .

RUN go install
EXPOSE 8440
EXPOSE 9999
#EXPOSE 8221

ENV hostip=172.17.0.3 \
dir=/go/src/hub/app

ENTRYPOINT ["/go/src/hub/start.sh"]
