package protobuf

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// MessageHandler handles protobuf message operations
type MessageHandler struct {
	registry *protoregistry.Files
}

// LoadProtoFile reads a protobuf file and registers its messages
func (h *MessageHandler) LoadProtoFile(protoPath string) error {
	// TODO: Implement proto file loading and registration
	return nil
}

// CreateMessage creates a new protobuf message of the specified type
func (h *MessageHandler) CreateMessage(messageType string) (proto.Message, error) {
	mt, err := protoregistry.GlobalTypes.FindMessageByName(protoreflect.FullName(messageType))
	if err != nil {
		return nil, fmt.Errorf("failed to find message type: %v", err)
	}
	return mt.New().Interface(), nil
}

// SerializeMessage serializes a protobuf message to bytes
func (h *MessageHandler) SerializeMessage(message proto.Message) ([]byte, error) {
	return proto.Marshal(message)
}

// DeserializeMessage deserializes bytes into a protobuf message
func (h *MessageHandler) DeserializeMessage(messageType string, data []byte) (proto.Message, error) {
	message, err := h.CreateMessage(messageType)
	if err != nil {
		return nil, err
	}
	if err := proto.Unmarshal(data, message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}
	return message, nil
}

// SetField sets a field value in a protobuf message
func (h *MessageHandler) SetField(message proto.Message, fieldName string, value interface{}) error {
	reflection := message.ProtoReflect()
	field := reflection.Descriptor().Fields().ByName(protoreflect.Name(fieldName))
	if field == nil {
		return fmt.Errorf("field %s not found", fieldName)
	}

	switch field.Kind() {
	case protoreflect.StringKind:
		reflection.Set(field, protoreflect.ValueOfString(value.(string)))
	case protoreflect.Int32Kind:
		reflection.Set(field, protoreflect.ValueOfInt32(value.(int32)))
	case protoreflect.Int64Kind:
		reflection.Set(field, protoreflect.ValueOfInt64(value.(int64)))
	case protoreflect.Uint32Kind:
		reflection.Set(field, protoreflect.ValueOfUint32(value.(uint32)))
	case protoreflect.Uint64Kind:
		reflection.Set(field, protoreflect.ValueOfUint64(value.(uint64)))
	case protoreflect.BoolKind:
		reflection.Set(field, protoreflect.ValueOfBool(value.(bool)))
	case protoreflect.FloatKind:
		reflection.Set(field, protoreflect.ValueOfFloat32(value.(float32)))
	case protoreflect.DoubleKind:
		reflection.Set(field, protoreflect.ValueOfFloat64(value.(float64)))
	case protoreflect.BytesKind:
		reflection.Set(field, protoreflect.ValueOfBytes(value.([]byte)))
	default:
		return fmt.Errorf("unsupported field kind: %v", field.Kind())
	}
	return nil
}

// GetField gets a field value from a protobuf message
func (h *MessageHandler) GetField(message proto.Message, fieldName string) (interface{}, error) {
	reflection := message.ProtoReflect()
	field := reflection.Descriptor().Fields().ByName(protoreflect.Name(fieldName))
	if field == nil {
		return nil, fmt.Errorf("field %s not found", fieldName)
	}

	value := reflection.Get(field)
	switch field.Kind() {
	case protoreflect.StringKind:
		return value.String(), nil
	case protoreflect.Int32Kind:
		return value.Int(), nil
	case protoreflect.Int64Kind:
		return value.Int(), nil
	case protoreflect.Uint32Kind:
		return value.Uint(), nil
	case protoreflect.Uint64Kind:
		return value.Uint(), nil
	case protoreflect.BoolKind:
		return value.Bool(), nil
	case protoreflect.FloatKind:
		return value.Float(), nil
	case protoreflect.DoubleKind:
		return value.Float(), nil
	case protoreflect.BytesKind:
		return value.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported field kind: %v", field.Kind())
	}
}
