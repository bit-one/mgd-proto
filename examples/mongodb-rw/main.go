package main

import (
	"bytes"
	"context"
	"log"
	"time"

	types "github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/jsonpb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	codecs "github.com/bit-one/mgd-proto"
)

func main() {
	log.Printf("connecting to MongoDB...")

	// Register custom codecs for protobuf Timestamp and wrapper types
	reg := codecs.Register(bson.NewRegistryBuilder()).Build()

	// Create MongoDB client with registered custom codecs for protobuf Timestamp and wrapper types
	// NOTE: "mongodb+srv" protocol means connect to Altas cloud MongoDB server
	//       use just "mongodb" if you connect to on-premise MongoDB server
	client, err := mongo.NewClient(options.Client().
		ApplyURI("mongodb://@127.0.0.1:27017/experiments").
		SetRegistry(reg),
	)

	if err != nil {
		log.Fatalf("failed to create new MongoDB client: %#v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect client
	if err = client.Connect(ctx); err != nil {
		log.Fatalf("failed to connect to MongoDB: %#v", err)
	}

	log.Printf("connected successfully")

	// Get collection from database
	coll := client.Database("experiments").Collection("proto")

	// Create protobuf Timestamp value from golang Time
	ts := types.TimestampNow()

	// Fill in data structure
	in := Data{
		BoolValue:      true,
		BoolProtoValue: &types.BoolValue{Value: true},

		BytesValue:      []byte{1, 2, 3, 4, 5},
		BytesProtoValue: &types.BytesValue{Value: []byte{1, 2, 3, 4, 5}},

		DoubleValue:      123.45678,
		DoubleProtoValue: &types.DoubleValue{Value: 123.45678},

		FloatValue:      123.45,
		FloatProtoValue: &types.FloatValue{Value: 123.45},

		Int32Value:      -12345,
		Int32ProtoValue: &types.Int32Value{Value: -12345},

		Int64Value:      -123456789000,
		Int64ProtoValue: &types.Int64Value{Value: -123456789000},

		StringValue:      "qwerty",
		StringProtoValue: &types.StringValue{Value: "qwerty"},

		Uint32Value:      12345,
		Uint32ProtoValue: &types.UInt32Value{Value: 12345},

		Uint64Value:      123456789000,
		Uint64ProtoValue: &types.UInt64Value{Value: 123456789000},

		Timestamp: ts,
		Listvalue: &types.ListValue{
			Values: []*types.Value{
				&types.Value{Kind: &types.Value_StringValue{StringValue: "a"}},
				&types.Value{Kind: &types.Value_NumberValue{NumberValue: 1.0}},
				&types.Value{Kind: &types.Value_BoolValue{BoolValue: false}},
				&types.Value{Kind: &types.Value_NullValue{NullValue: types.NullValue_NULL_VALUE}},
				&types.Value{Kind: &types.Value_StructValue{
					StructValue: &types.Struct{
						Fields: map[string]*types.Value{
							"helo": &types.Value{Kind: &types.Value_StringValue{StringValue: "world"}},
							"gee":  &types.Value{Kind: &types.Value_NumberValue{NumberValue: 3.0}},
						},
					},
				}},
				&types.Value{Kind: &types.Value_ListValue{
					ListValue: &types.ListValue{
						Values: []*types.Value{
							&types.Value{Kind: &types.Value_StringValue{StringValue: "b"}},
							&types.Value{Kind: &types.Value_NumberValue{NumberValue: 2.0}},
						},
					},
				}},
			},
		},
	}

	log.Printf("insert data into collection <experiments.proto>...")

	// Insert data into the collection
	res, err := coll.InsertOne(ctx, &in)
	if err != nil {
		log.Fatalf("insert data into collection <experiments.proto>: %#v", err)
	}
	id := res.InsertedID
	log.Printf("inserted new item with id=%v successfully", id)

	// Create filter and output structure to read data from collection
	var out Data
	filter := bson.D{{Key: "_id", Value: id}}

	// Read data from collection
	err = coll.FindOne(ctx, filter).Decode(&out)
	if err != nil {
		log.Fatalf("failed to read data (id=%v) from collection <experiments.proto>: %#v", id, err)
	}
	log.Printf("out: %#v", out)

	var b bytes.Buffer
	m := &jsonpb.Marshaler{Indent: "  "}
	if err := m.Marshal(&b, &out); err != nil {
		log.Fatalf("jsonpb.Marshaler.Marshal error = %v", err)
	}

	log.Printf("read successfully:\n%s", b.String())
}
