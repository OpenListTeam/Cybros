package memberdump

import (
	"encoding/base64"
	"reflect"

	"github.com/gotd/td/tdp"
)

type tlObject interface {
	TypeInfo() tdp.Type
}

func marshalTLObject(object tlObject) map[string]any {
	if object == nil {
		return nil
	}

	value := reflect.ValueOf(object)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return nil
	}

	info := object.TypeInfo()
	if info.Null {
		return map[string]any{
			"@type": info.Name,
		}
	}

	for value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
	}

	result := map[string]any{
		"@type": info.Name,
	}
	for _, field := range info.Fields {
		if field.Null {
			continue
		}

		fieldValue := value.FieldByName(field.Name)
		result[field.SchemaName] = marshalTLValue(fieldValue)
	}

	return result
}

func marshalTLValue(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}
	if value.CanInterface() {
		object, ok := value.Interface().(tlObject)
		if ok {
			return marshalTLObject(object)
		}
	}

	for value.Kind() == reflect.Interface || value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		if value.CanInterface() {
			object, ok := value.Interface().(tlObject)
			if ok {
				return marshalTLObject(object)
			}
		}
		value = value.Elem()
	}

	if value.Kind() == reflect.Struct && value.CanAddr() {
		object, ok := value.Addr().Interface().(tlObject)
		if ok {
			return marshalTLObject(object)
		}
	}

	switch value.Kind() {
	case reflect.Bool:
		return value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint()
	case reflect.Float32, reflect.Float64:
		return value.Float()
	case reflect.String:
		return value.String()
	case reflect.Slice:
		if value.Type().Elem().Kind() == reflect.Uint8 {
			return base64.RawURLEncoding.EncodeToString(value.Bytes())
		}

		items := make([]any, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			items = append(items, marshalTLValue(value.Index(i)))
		}
		return items
	case reflect.Array:
		items := make([]any, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			items = append(items, marshalTLValue(value.Index(i)))
		}
		return items
	default:
		if value.CanInterface() {
			return value.Interface()
		}
		return nil
	}
}
