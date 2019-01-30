#! /bin/bash
set -x

for i in {1..10}; do
	echo "Creating Container send_$i"
	docker run -d --privileged -h "send" --name "send_$i" "send" &
done

# docker kill $(docker ps | grep send | awk '{ print $1 }')
# docker rm $(docker ps -a | grep send | awk '{ print $1 }')
