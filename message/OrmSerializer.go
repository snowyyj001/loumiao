package message

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"gorm.io/gorm/schema"
	"reflect"
)

// ProtoSerializer proto serializer
type ProtoSerializer struct {
}

// Scan implements serializer interface
func (ProtoSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)
	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			return fmt.Errorf("failed to unmarshal proto value: %#v", dbValue)
		}

		err = proto.Unmarshal(bytes, fieldValue.Interface().(proto.Message))
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return
}

// Value implements serializer interface
func (ProtoSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	return proto.Marshal(fieldValue.(proto.Message))
}

func init() {
	schema.RegisterSerializer("proto", ProtoSerializer{})
}
