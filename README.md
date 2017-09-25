# gontpd

A sntp daemon written in Go
Support Linux/Darwin(Mac os)

## install
```
go get github.com/mengzhuo/gontpd/cmd/gontpd
```

## config
```
# ntp listen address should be :123
listen: :5123 

# prometheus listen address
metric: :9090 

# maxmind geo db path
geodb: helloWorld.geo 

# how many worker to handle ntp request
worker: 7 

# server upper stratum
server:
    - time1.aliyun.com
    - time2.aliyun.com
    - time3.aliyun.com
```
