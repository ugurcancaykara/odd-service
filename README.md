# Odd-service

### A simple movie api 

- Consists of 3 microservice
  - Movie metadata service
  - Rating service
  - Movie(Client-Facing API) service  



## Staring microservices

Run consul locally
```
  docker run -d -p 8500:8500 -p 8600:8600/udp --name=dev-consul consul:1.13 agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0
```

Run each microservice by executing
```
  go run *.go
```


Optional: You can optionally add some additional instances of each service by running:
```
go run *.go --port <PORT>
```
* If you run the preceding command, replace `<PORT>` placeholder with unique port numbers that are not in use yet 
* We used 8081,8082 and 8083, so you can run with port numbers starting with 8084

## Open Consul UI if you use consul client-side service discovery implementation

Open browser and enter:
```
  http://localhost:8500/
```

## Testing API

To test API requests, ensure you have at least one healthy instance of each service and make the following request to a movie service
```
  curl -v localhost:8083/movie?id=1
```
