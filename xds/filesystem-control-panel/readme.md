### xds

xds implement demo how to use xds server.

### to run sample:

1. start xds server
```shell
./server/app> go run .
```

2. run pixiu 
```shell
dubbo-go-pixiu > pixiu gateway start -c ./samples/xds/local-control-panel/pixiu/conf.yaml -g test/configs/log.yml
```

3. check result
```shell
curl -v  'localhost:8888/get'

## will get result below 
{
  "args": {},
  "headers": {
    "Accept": "*/*",
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "curl/7.64.1",
    "X-Amzn-Trace-Id": "Root=1-61ba16a5-3ea1961217b2ffa7124ea2c2"
  },
  "origin": "223.104.41.209",
  "url": "http://httpbin.org/get"
}
```
4. change pixiu config file & check result