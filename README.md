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






## Using Protocol Buffers

### Requirements
- install protoc-gen-go package -> Since I use mac, I used brew to install it 
```
brew install protoc-gen-go
```


You can directly install the binary
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

Follow the docs for package details -> https://pkg.go.dev/github.com/golang/protobuf/protoc-gen-go

### For simple benchmark results 
change directory to -> encodingbenchmark/sizeandcompare
You can check the benchmark results at encodingbenchmark/sizeandcompare/README.md and also see it yourself


Since Protocol Buffers offers faster encoding/decoding speed and smaller output size we are going to use it with gRPC.
Define your data model inside api/ folder at movie.proto 
After we defined our go struct data model inside proto file we can generate codes with below command

```
protoc -I=api --go_out=. movie.proto
```



To create service code in gRPC format (MetadataService(and it's rpc funcs and other services etc.))
Define service structures and their request, response methods' structures, run the below command
```
  protoc -I=api --go_out=. --go-grpc_out=. movie.proto
```

It's similar to the command that we used in the previous struct generation however, it also passes a --go-grpc_out 
flag to the compiler. This flag tells the Protocol buffers compiler to generate the service code in gRPC format.
