#bash
 
protoc --proto_path=. --proto_path=../../proto --proto_path=../../proto/third_party --proto_path=../third_party --proto_path=$GOPATH/src/github.com/gogo/protobuf/protobuf --gofast_out=Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types:. data.proto

# protoc --proto_path=. --proto_path=../../proto --proto_path=../../proto/third_party --proto_path=../third_party --gotag_out=xxx="bson+\"-\"",output_path=.:. data.proto
