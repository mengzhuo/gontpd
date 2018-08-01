# gontpd

A high performance sntp daemon focus on stratum > 2 serving

Only support Linux

## install
```
go get github.com/mengzhuo/gontpd/cmd/gontpd
```
or
```
./make.sh # to build debian package
```

## run
```
gontpd -c config.yml -f 0
```

## config
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

## Performance
```
KVM Intel(R) Xeon(R) CPU E5-2697 v4 @ 2.30GHz
~60kpps per thread
```
