# GoNTPd
[![GoDoc](https://godoc.org/github.com/mengzhuo/gontpd?status.svg)](https://godoc.org/github.com/mengzhuo/gontpd)
[![Build Status](https://travis-ci.org/mengzhuo/gontpd.svg?branch=master)](https://travis-ci.org/mengzhuo/gontpd)

A high performance NTP daemon written in Go.

Only support Linux

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

# force_update: force update time if offset is over 128ms or terminal processs
force_update: true

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

# geo_db: MaxMind GeoLite2 DB path, it won't stat CountryCode is empty
# NOTE: This will cause high CPU usage, use with caution
geo_db: 

# max/min poll interval seconds (log2) to upstream peer
# i.e. 10 = 1024 seconds
max_poll: 9
min_poll: 4

# peer_list: upstream peer list that sync to
peer_list:
    - time1.apple.com
    - time2.apple.com
    - time3.apple.com
    - time4.apple.com

# max_std: maximum standard deviation of peer that we consider as a good peer.
max_std: 50ms

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
