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




## Using Apache Kafka

### Run apache kafka using docker-compose.yml
Requirements:
- You need to install docker and docker-compose

You can run kafka docker containers by the following command:
```
docker-compose up -d
```

Now we need our topics to be created, so you can also manually create it by running below command:
```
docker-compose exec kafka kafka-topics --create --topic ratings --partitions 1 --replication-factor 1 --bootstrap-server kafka:9092
```

Publisher client -> Wrote a simple app at the top folder level -> cmd/ratingingester
- Reads events from ratingdata.json and publish them to locally running kafka topic(rating) thanks to docker.

It will use kafka-topics binary which installed inside kafka container and use it to create a topic named as 'ratings'
We use this topic name to publish messages from example app(cmd/ratingingester) which reads event data from ratingsdata.json file and publish it to ratings topic

Struct settings -> Created a Event struct for async functionallity of the Ratings API(Haven't done this using protoc gen go). -> ratings/pkg/model/ratings.go
Ingester(consumer) settings -> Then we implemented our rating service which can be found under the rating/ folder to consume messages from this 'ratings' topic -> rating/internal/ingester/kafka
Controller settings -> Then controller implementation can be found under -> rating/internal/controller/rating/controller.go. And for now since I didn't automate this process(I use in-memory storage) 
Program start up settings -> I don't prefer to initialize ingester for now, will initalize later.(rating/cmd/main.go -> new.ratinggateway(repo, nil(nil is for ingester)))




## Using MySQL as data storage
Requirements:
- You need to install docker and docker-compose

It is inside docker-compose.yaml file alongside kafka and zookeper
```
docker-compose up -d
```

Still you need to run the following command to create table

```
  CREATE DATABASE movie
```

And change the running container name inside the below command which uses schema.sql file under the schema folder to create tables inside movie database

```
  docker exec -i container_name mysql movie -h localhost -p 3306 --protocol=tcp -uroot -ppassword < schema/schema.sql
```

You can check if the tables were created successfully by running the following command:
# TODO: this doesn't work fix it
```
  docker exec -i container_name mysql movie -h localhost -p 3306 --protocol=tcp -uroot -ppassword -e "SHOW tables"
```





## Testing the API with grpcurl
After having every resources up and running

Run to see AggregatedRating Value 
```
  grpcurl -plaintext -v -d '{"record_id":"1","record_type":"movie"}' localhost:8082 RatingService/GetAggregatedRating
```
however you will end up having code: NotFound, since we haven't added any record let's add some.


```
grpcurl -plaintext -d '{"record_id":"1","record_type":"movie","user_id":"keke","rating_value":3}' localhost:8082 RatingService/PutRating
```

After that, we will have records at our MySQL table, you can run first grpcurl command to see aggregatedvalue response

