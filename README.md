
# Relay - Container Network Capture and Graph


Current POC View based on Network traffic capture of packet data and displayed as particles in D3.js through websocket.


![intention](https://stash.aexp.com/stash/users/cholzing/repos/relay_poc/raw/slides/images/relay_hub_many1.gif)




## Build

docker build -t relay ./relay

docker build -t hub ./hub

docker build -t accept ./accept

docker build -t send ./send



## Start up

docker run -d --privileged --name relay relay

docker run -d -p 9999:9999 --name hub hub

docker run -d --privileged -h accept --name accept accept 

./send/sender.sh
