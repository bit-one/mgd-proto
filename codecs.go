package codecs

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/bit-one/mgd-proto/pmongo"
	"github.com/gogo/protobuf/types"
)

var (
	// Protobuf types types
	boolValueType   = reflect.TypeOf(types.BoolValue{})
	bytesValueType  = reflect.TypeOf(types.BytesValue{})
	doubleValueType = reflect.TypeOf(types.DoubleValue{})
	floatValueType  = reflect.TypeOf(types.FloatValue{})
	int32ValueType  = reflect.TypeOf(types.Int32Value{})
	int64ValueType  = reflect.TypeOf(types.Int64Value{})
	stringValueType = reflect.TypeOf(types.StringValue{})
	uint32ValueType = reflect.TypeOf(types.UInt32Value{})
	uint64ValueType = reflect.TypeOf(types.UInt64Value{})

	//ListValue
	tListValueType     = reflect.TypeOf(types.ListValue{})
	tValue             = reflect.TypeOf(types.Value{})
	tNullValue         = reflect.TypeOf(types.NullValue(0))
	tValue_StringValue = reflect.TypeOf(types.Value_StringValue{})
	tValue_NullValue   = reflect.TypeOf(types.Value_NullValue{})
	tValue_NumberValue = reflect.TypeOf(types.Value_NumberValue{})
	tValue_BoolValue   = reflect.TypeOf(types.Value_BoolValue{})
	tValue_StructValue = reflect.TypeOf(types.Value_StructValue{})
	tValue_ListValue   = reflect.TypeOf(types.Value_ListValue{})
	tStruct            = reflect.TypeOf(types.Struct{})

	// Protobuf Timestamp type
	timestampType = reflect.TypeOf(types.Timestamp{})

	// Time type
	timeType = reflect.TypeOf(time.Time{})

	// ObjectId type
	objectIDType          = reflect.TypeOf(pmongo.ObjectId{})
	objectIDPrimitiveType = reflect.TypeOf(primitive.ObjectID{})

	// Codecs
	wrapperValueCodecRef = &wrapperValueCodec{}
	listValueCodecRef    = &listValueCodec{}
	valueCodecRef        = &valueCodec{}
	timestampCodecRef    = &timestampCodec{}
	objectIDCodecRef     = &objectIDCodec{}
	nullValueCodecRef    = &nullValueCodec{}
)

var (
	convFromType = map[bsontype.Type]reflect.Type{
		bsontype.String:           tValue_StringValue,
		bsontype.Boolean:          tValue_BoolValue,
		bsontype.Double:           tValue_NumberValue,
		bsontype.Array:            tValue_ListValue,
		bsontype.EmbeddedDocument: tValue_StructValue,
		bsontype.Null:             tValue_NullValue,
	}
)

var _ bsoncodec.ValueCodec
var _ bsoncodec.CodecZeroer

type nullValueCodec struct{}

func (e *nullValueCodec) EncodeValue(ectx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	return vw.WriteNull()
}

func (d *nullValueCodec) DecodeValue(ectx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val.Set(reflect.ValueOf(types.NullValue_NULL_VALUE))
	return vr.ReadNull()
}

type valueCodec struct{}

func (e *valueCodec) EncodeValue(ectx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val = reflect.Indirect(val)
	val = val.Field(0)
	enc, err := ectx.LookupEncoder(val.Type())
	if err != nil {
		return err
	}
	return enc.EncodeValue(ectx, vw, val)
}

// DecodeValue decodes BSON value to Protobuf type wrapper value
func (d *valueCodec) DecodeValue(ectx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val = reflect.Indirect(val)
	x := val.Field(0)
	enc, err := ectx.LookupDecoder(x.Type())
	if err != nil {
		return err
	}
	if err := enc.DecodeValue(ectx, vr, x); err != nil {
		return err
	}
	return nil
}

type listValueCodec struct{}

func (e *listValueCodec) EncodeValue(ectx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val = val.Field(0)
	aw, err := vw.WriteArray()
	if err != nil {
		return err
	}
	for idx := 0; idx < val.Len(); idx++ {
		v := val.Index(idx)
		vw, err := aw.WriteArrayElement()
		if err != nil {
			return err
		}
		for v.Kind() == reflect.Ptr {
			if v.IsNil() {
				vw.WriteNull()
				continue
			}
			v = v.Elem()
		}
		if v.Type() == tValue {
			v = v.Field(0)
		}
		x := v.Interface()
		t := reflect.TypeOf(x).Elem()
		enc, err := ectx.LookupEncoder(t)
		if err != nil {
			return err
		}
		err = enc.EncodeValue(ectx, vw, reflect.ValueOf(x))
		if err != nil {
			return err
		}
	}
	return aw.WriteArrayEnd()
}

func (d *listValueCodec) DecodeValue(ectx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	elems := make([]reflect.Value, 0)
	vals := val.Field(0)

	ar, err := vr.ReadArray()
	if err != nil {
		return err
	}

	for {
		vr, err := ar.ReadValue()
		if err == bsonrw.ErrEOA {
			break
		}
		if err != nil {
			return err
		}
		cvt, ok := convFromType[vr.Type()]
		if !ok {
			return errors.New(fmt.Sprintf("mgd-proto: Unexpected Type: %+v\n", vr.Type()))
		}
		decoder, err := ectx.LookupDecoder(cvt)
		if err != nil {
			return err
		}
		elem := reflect.New(cvt)
		err = decoder.DecodeValue(ectx, vr, elem)
		if err != nil {
			return err
		}
		target := reflect.New(tValue).Elem()
		target.Field(0).Set(elem)
		elems = append(elems, target)
	}

	vals.Set(reflect.ValueOf(make([]*types.Value, len(elems))))
	for idx, elem := range elems {
		vals.Index(idx).Set(elem.Addr())
	}
	return nil
}

// wrapperValueCodec is codec for Protobuf type types
type wrapperValueCodec struct {
}

// EncodeValue encodes Protobuf type wrapper value to BSON value
func (e *wrapperValueCodec) EncodeValue(ectx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	val = reflect.Indirect(val)
	val = val.Field(0)
	enc, err := ectx.LookupEncoder(val.Type())
	if err != nil {
		return err
	}
	return enc.EncodeValue(ectx, vw, val)
}

// DecodeValue decodes BSON value to Protobuf type wrapper value
func (e *wrapperValueCodec) DecodeValue(ectx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	val = reflect.Indirect(val)
	val = val.Field(0)
	enc, err := ectx.LookupDecoder(val.Type())
	if err != nil {
		return err
	}
	return enc.DecodeValue(ectx, vr, val)
}

// timestampCodec is codec for Protobuf Timestamp
type timestampCodec struct {
}

// EncodeValue encodes Protobuf Timestamp value to BSON value
func (e *timestampCodec) EncodeValue(ectx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	v := val.Interface().(types.Timestamp)
	t, err := types.TimestampFromProto(&v)
	enc, err := ectx.LookupEncoder(timeType)
	if err != nil {
		return err
	}
	return enc.EncodeValue(ectx, vw, reflect.ValueOf(t.In(time.UTC)))
}

// DecodeValue decodes BSON value to Timestamp value
func (e *timestampCodec) DecodeValue(ectx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	enc, err := ectx.LookupDecoder(timeType)
	if err != nil {
		return err
	}
	var t time.Time
	if err = enc.DecodeValue(ectx, vr, reflect.ValueOf(&t).Elem()); err != nil {
		return err
	}
	ts, err := types.TimestampProto(t.In(time.UTC))
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(*ts))
	return nil
}

// objectIDCodec is codec for Protobuf ObjectId
type objectIDCodec struct {
}

// EncodeValue encodes Protobuf ObjectId value to BSON value
func (e *objectIDCodec) EncodeValue(ectx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	v := val.Interface().(pmongo.ObjectId)
	// Create primitive.ObjectId from string
	id, err := primitive.ObjectIDFromHex(v.Value)
	if err != nil {
		return err
	}
	enc, err := ectx.LookupEncoder(objectIDPrimitiveType)
	if err != nil {
		return err
	}
	return enc.EncodeValue(ectx, vw, reflect.ValueOf(id))
}

// DecodeValue decodes BSON value to ObjectId value
func (e *objectIDCodec) DecodeValue(ectx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	enc, err := ectx.LookupDecoder(objectIDPrimitiveType)
	if err != nil {
		return err
	}
	var id primitive.ObjectID
	if err = enc.DecodeValue(ectx, vr, reflect.ValueOf(&id).Elem()); err != nil {
		return err
	}
	oid := *pmongo.NewObjectId(id)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(oid))
	return nil
}

// Register registers Google protocol buffers types codecs
func Register(rb *bsoncodec.RegistryBuilder) *bsoncodec.RegistryBuilder {
	return rb.RegisterCodec(boolValueType, wrapperValueCodecRef).
		RegisterCodec(bytesValueType, wrapperValueCodecRef).
		RegisterCodec(doubleValueType, wrapperValueCodecRef).
		RegisterCodec(floatValueType, wrapperValueCodecRef).
		RegisterCodec(int32ValueType, wrapperValueCodecRef).
		RegisterCodec(int64ValueType, wrapperValueCodecRef).
		RegisterCodec(stringValueType, wrapperValueCodecRef).
		RegisterCodec(uint32ValueType, wrapperValueCodecRef).
		RegisterCodec(uint64ValueType, wrapperValueCodecRef).
		RegisterCodec(timestampType, timestampCodecRef).
		RegisterCodec(objectIDType, objectIDCodecRef).
		RegisterCodec(tListValueType, listValueCodecRef).
		RegisterCodec(tValue_StringValue, valueCodecRef).
		RegisterCodec(tValue_NullValue, valueCodecRef).
		RegisterCodec(tValue_NumberValue, valueCodecRef).
		RegisterCodec(tValue_BoolValue, valueCodecRef).
		RegisterCodec(tValue_StructValue, valueCodecRef).
		RegisterCodec(tValue_ListValue, valueCodecRef).
		RegisterCodec(tNullValue, nullValueCodecRef)

}
