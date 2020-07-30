package codecs

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/jsonpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bit-one/mgd-proto/pmongo"
	"github.com/bit-one/mgd-proto/test"
)

func TestCodecs(t *testing.T) {
	rb := bson.NewRegistryBuilder()
	r := Register(rb).Build()

	tm := time.Now()
	// BSON accuracy is in milliseconds
	tm = time.Date(tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(),
		(tm.Nanosecond()/1000000)*1000000, tm.Location())

	ts, err := types.TimestampProto(tm)
	if err != nil {
		t.Errorf("ptypes.TimestampProto error = %v", err)
		return
	}

	objectID := primitive.NewObjectID()
	id := pmongo.NewObjectId(objectID)

	t.Run("primitive object id", func(t *testing.T) {
		resultID, err := id.GetObjectID()
		if err != nil {
			t.Errorf("mongodb.ObjectId.GetPrimitiveObjectID() error = %v", err)
			return
		}

		if !reflect.DeepEqual(objectID, resultID) {
			t.Errorf("failed: primitive object ID=%#v, ID=%#v", objectID, id)
			return
		}
	})

	in := test.Data{
		BoolValue:   &types.BoolValue{Value: true},
		BytesValue:  &types.BytesValue{Value: make([]byte, 5)},
		DoubleValue: &types.DoubleValue{Value: 1.2},
		FloatValue:  &types.FloatValue{Value: 1.3},
		Int32Value:  &types.Int32Value{Value: -12345},
		Int64Value:  &types.Int64Value{Value: -123456789},
		StringValue: &types.StringValue{Value: "qwerty"},
		Uint32Value: &types.UInt32Value{Value: 12345},
		Uint64Value: &types.UInt64Value{Value: 123456789},
		Timestamp:   ts,
		Id:          id,
		Listvalue: &types.ListValue{
			Values: []*types.Value{
				&types.Value{Kind: &types.Value_StringValue{StringValue: "a"}},
				&types.Value{Kind: &types.Value_NumberValue{NumberValue: 1.0}},
				&types.Value{Kind: &types.Value_BoolValue{BoolValue: false}},
				&types.Value{Kind: &types.Value_NullValue{NullValue: types.NullValue_NULL_VALUE}},
				// &types.Value{Kind: &types.Value_StructValue{
				// 	StructValue: &types.Struct{
				// 		Fields: map[string]*types.Value{
				// 			"helo": &types.Value{Kind: &types.Value_StringValue{StringValue: "world"}},
				// 			"gee":  &types.Value{Kind: &types.Value_NumberValue{NumberValue: 3.0}},
				// 		},
				// 	},
				// }},
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
	fmt.Println(in)
	t.Run("marshal/unmarshal", func(t *testing.T) {
		b, err := bson.MarshalWithRegistry(r, &in)
		if err != nil {
			t.Errorf("bson.MarshalWithRegistry error = %v", err)
			return
		}

		var out test.Data
		if err = bson.UnmarshalWithRegistry(r, b, &out); err != nil {
			t.Errorf("bson.UnmarshalWithRegistry error = %v", err)
			return
		}

		fmt.Println(out)
		if !reflect.DeepEqual(in, out) {
			t.Errorf("failed: in=%#v, out=%#v", in, out)
			return
		}
	})

	t.Run("marshal-jsonpb/unmarshal-jsonpb", func(t *testing.T) {
		var b bytes.Buffer

		m := &jsonpb.Marshaler{}

		if err := m.Marshal(&b, &in); err != nil {
			t.Errorf("jsonpb.Marshaler.Marshal error = %v", err)
			return
		}

		var out test.Data
		if err = jsonpb.Unmarshal(&b, &out); err != nil {
			t.Errorf("jsonpb.Unmarshal error = %v", err)
			return
		}

		if !reflect.DeepEqual(in, out) {
			t.Errorf("failed: in=%#v, out=%#v", in, out)
			return
		}
	})
}
