#!/bin/bash
set -x

echo "prebuild setup"

eval sysctl -w net.ipv4.conf.all.forwarding=1
#eval sysctl -w net.ipv4.conf.eth0.route_localnet=1
eval sysctl -w net.ipv4.conf.all.accept_local=1
eval /etc/init.d/rsyslog restart

eval iptables -I INPUT -j ACCEPT
eval iptables -I OUTPUT -j ACCEPT #-p http,tcp,ftp,icmp
eval iptables -t mangle -I POSTROUTING -p tcp -m tcp -j TEE --gateway 172.17.0.2

export hostip=$(ip -4 addr show eth0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}')
echo "$hostip"


#eval iptables -A FORWARD ACCEPT
## Local to Local traffic
#eval iptables -t mangle -I INPUT 1 -p tcp -j TEE --gateway 192.168.0.8
#eval iptables -t mangle -I OUTPUT 1 -p tcp -j TEE --gateway 192.168.0.8
## flows from network interface to local process -- Inbound
#eval iptables -t mangle -A PREROUTING -s 172.17.0.2 -d 172.17.0.2 -j TEE --gateway 10.68.244.176
#eval iptables -t mangle -A POSTROUTING -p tcp -j TEE --gateway 10.68.232.43
#eval iptables -t mangle -I PREROUTING 1 -p tcp -j ACCEPT


## flows from local process to network interface -- Outbound
#eval iptables -t mangle -A PREROUTING -p tcp -j TEE --gateway 10.68.232.43
#eval iptables -t mangle -I PREROUTING 1 -p tcp -j ACCEPT


## Log 
#eval iptables -t mangle -I PREROUTING 1 -j LOG --log-level=info
#eval iptables -t mangle -I POSTROUTING 1 -p tcp -j LOG --log-level=info

#eval iptables -t mangle -I PREROUTING -p -tcp --tcp-flags PSH ACK -j TEE --gateway 192.168.0.8


#TEE --gateway 192.168.0.9

#eval iptables -t mangle -I INPUT -p tcp -j TEE --gateway 192.168.0.9
#eval iptables -t mangle -I OUTPUT -p tcp -j TEE --gateway 192.168.0.9

#eval iptables -t nat -A PREROUTING -j MASQUERADE
#eval iptables -t nat -A POSTROUTING -j MASQUERADE

#eval iptables -t mangle -A PREROUTING -d 0.0.0.0/0 -j TEE --gateway 172.17.0.0
#eval iptables -t mangle -A POSTROUTING -s 0.0.0.0/0 -j TEE --gateway 172.17.0.0

#eval iptables -t mangle -A PREROUTING -d 0.0.0.0/0 -p tcp -j TEE --gateway 192.168.0.9
#eval iptables -t mangle -A POSTROUTING -s 0.0.0.0/0 -p tcp -j TEE --gateway 192.168.0.9

#eval iptables -t mangle -A PREROUTING -i eth0 -p tcp -j TEE --gateway 192.168.0.9
# -i eth0 -p tcp 
#eval iptables -t mangle -A POSTROUTING -j TEE --gateway 192.168.0.9
# DNAT --to-destination 11.11.11.11 


/go/bin/accept


