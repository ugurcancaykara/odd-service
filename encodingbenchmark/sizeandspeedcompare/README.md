# Just a simple benchmark for different serialization protocols

### XML and JSON uses Metadata data model from github.com/ugurcancaykara/odd-service/metadata/pkg/model/metadata.go 

```
  metadata := &model.Metadata{
		ID:          "12345",
		Title:       "The Movie 2",
		Description: "Sequel of the legendary The Movie",
		Director:    "Foo Bars",
	}
```



### Protobuf uses Metadata data model from github.com/ugurcancaykara/odd-service/gen/movie.pb.go 
```
  genMetadata := &gen.Metadata{
		Id:          "12345",
		Title:       "The Movie 2",
		Description: "Sequel of the legendary The Movie",
		Director:    "Foo Bars",
	}
```


### Size result:

```
JSON size: 	108B
XML size: 	150B
Proto size: 	65B
```

You can test it by running the following command inside encodingbenchmark/sizeandspeedcompare 
```
  go run main.go 
```


The result is quite interesting. The XML output almost 40% bigger than the JSON one. 
Ptorocol Buffers's output is more than 40% smaller than the JSON data and more than twice as
small as the XML result. This illustrates quite well how efficient the Protocol Buffers format 
is compared to the other two in terms of output size. By switching from JSON to Protocol Buffers,
we reduce the amount of data that we need to send over the network and make our communication faster.




### Benchmark result:

```
goos: darwin
goarch: arm64

BenchmarkSerializeToJSON-8    	 6534847	       159.3 ns/op
BenchmarkSerializeToXML-8     	 1000000	      1173 ns/op
BenchmarkSerializeToProto-8   	13086316	        91.49 ns/op
PASS
ok sizecompare	3.925s
```

You can test it by running the following command inside encodingbenchmark/sizeandspeedcompare
```
go test -bench=.
```


You can see the names of three functions that we just implemented and two numbers next to them:
- The first one is the number of times the function got executed
- The second is the average processing speed, measured in nanoseconds per operation

From the output, we can see that Protocol Buffers serialization on average took 91.49 nanoseconds, 
while JSON serialization was almost two times slower at 159.3 nanoseconds. XML serialization on average
took 1173 nanoseconds, being more that 11 times slower than Protocol Buffers, and more than 7 times slower 
than JSON serialization.
