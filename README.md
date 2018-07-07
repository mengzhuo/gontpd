# gontpd

A sntp daemon focus on high performance stratum >2 serving
Only support Linux

## install
```
go get github.com/mengzhuo/gontpd/cmd/gontpd
```

## run
```
gontpd -c config.yml -f 0
```

## config
```
# ntp listen address should be :123
listen: :5123

# prometheus listen address
metric: :7370

# maxmind geo db path
geodb: helloWorld.geo

# how many worker(CPU) to handle ntp request
worker: 7

# server upper stratum
server:
    - time1.aliyun.com
    - time2.aliyun.com
    - time3.aliyun.com
```
