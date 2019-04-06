# GoNTPd
[![GoDoc](https://godoc.org/github.com/mengzhuo/gontpd?status.svg)](https://godoc.org/github.com/mengzhuo/gontpd)
[![Build Status](https://travis-ci.org/mengzhuo/gontpd.svg?branch=master)](https://travis-ci.org/mengzhuo/gontpd)

gontpd is an experimental high-performance NTP server written in Go. 
It does not implement a full NTP client and relies on another NTP client and server to be running on the system instead. It periodically updates its state to mirror the real NTP client/server and uses multiple threads to serve the current system time.
Inspired by [rsntp](https://github.com/mlichvar/rsntp)

## Install or Build

```
# require go1.11
go get github.com/mengzhuo/gontpd/cmd/gontpd
```
or
```
# require fpm
./make.sh #  build debian package your
```

## Run
```
gontpd -c config.yml
```

## Config
```
# listen: gontpd service listen port (UDP)
listen: ':123'

# worker_num: goroutines per connection
worker_num: 1

# conn_num: numbers of connections
conn_num: 1

# rate: LRU size of rate limmiter
# if drop is true, limmiter will drop the client request instead of sending RATE KoD response to client.
rate_size: 8196
rate_drop: true

# metric: prometheus stat listen port
metric: ':7370'

# peer_list: upstream peer list that sync to
up_state: 127.0.0.1

# drop_cidr: remote address within this list will be drop
# suggest to drop private net request(mostly are spoof request)
drop_cidr:
    - "192.168.0.0/16"
    - "172.16.0.0/12"
    - "10.0.0.0/8"
    - "100.64.0.0/10"

```

## Operation

iptables
```
-t raw -A PREROUTING -p udp -m udp --dport 123 -j NOTRACK
-t raw -A OUTPUT -p udp -m udp --sport 123 -j NOTRACK
```
sysctl
```
net.core.rmem_default = 512992
net.core.rmem_max = 512992
net.core.wmem_default = 512992
net.core.wmem_max = 512992
```

## Performance
```
Intel(R) Core(TM) i7-4790 CPU @ 3.60GHz
~180kpps @ GOMAXPROCS=1
```
