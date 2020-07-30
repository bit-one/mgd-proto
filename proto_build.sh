#bash
 
protoc --proto_path=proto/pmongo --gofast_out=../../../ objectid.proto

protoc -I=. -I=proto -I=proto/third_party -I=$GOPATH/src/github.com/gogo/protobuf/protobuf --gofast_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:. test/codecs_test.proto
