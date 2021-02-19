## Usage

Start a application running in a docker container with the name `www.bookinfo.com` on the `default` docker network
This name and container port determines the service URL / `serve-addr` for the application.
```
docker run -d --name bookinfo.backend.com docker.io/istio/examples-bookinfo-productpage-v1:1.16.2
```

Setup API Server 

```
getenvoy api-server local --listen-addr http://0.0.0.0:5050 --serve-addr http://bookinfo.backend.com:9080 --swagger-file bookinfo.yaml
```

On a different terminal, access the application on the `listen-addr` using `curl` 

```
 curl -ivv -H "Host:www.bookinfo.com" http://localhost:5050/api/v1/products
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 5050 (#0)
> GET /api/v1/products HTTP/1.1
> Host:www.bookinfo.com
> User-Agent: curl/7.64.1
> Accept: */*
> 
< HTTP/1.1 200 OK
HTTP/1.1 200 OK
< content-type: application/json
content-type: application/json
< content-length: 395
content-length: 395
< server: istio-envoy
server: istio-envoy
< date: Thu, 24 Jun 2021 00:34:56 GMT
date: Thu, 24 Jun 2021 00:34:56 GMT
< x-envoy-upstream-service-time: 1
x-envoy-upstream-service-time: 1

< 
* Connection #0 to host localhost left intact
[{"id": 0, "title": "The Comedy of Errors", "descriptionHtml": "<a href=\"https://en.wikipedia.org/wiki/The_Comedy_of_Errors\">Wikipedia Summary</a>: The Comedy of Errors is one of <b>William Shakespeare's</b> early plays. It is his shortest and one of his most farcical comedies, with a major part of the humour coming from slapstick and mistaken identity, in addition to puns and word play."}]* Closing connection 0
```
