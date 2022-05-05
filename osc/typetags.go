package osc

import "fmt"

type TypeTag rune

const (
	TypeString  TypeTag = 's'
	TypeInt32   TypeTag = 'i'
	TypeInt64   TypeTag = 'h'
	TypeFloat32 TypeTag = 'f'
	TypeFloat64 TypeTag = 'd'
	TypeBlob    TypeTag = 'b'
	TypeTimeTag TypeTag = 't'
	TypeNil     TypeTag = 'N'
	TypeTrue    TypeTag = 'T'
	TypeFalse   TypeTag = 'F'
	TypeInvalid TypeTag = 0
)

type TypeError struct {
	typeString string
}

func (t *TypeError) Error() string {
	return fmt.Sprintf("osc: unsupported type: %s", t.typeString)
}

func NewTypeError(i interface{}) *TypeError {
	return &TypeError{fmt.Sprintf("%T", i)}
}

type TypeTagError struct {
	typeTag rune
}

func (t *TypeTagError) Error() string {
	return fmt.Sprintf("osc: unsupported TypeTag: %c", t.typeTag)
}

func NewTypeTagError(t rune) *TypeTagError {
	return &TypeTagError{t}
}

// ToTypeTag returns the OSC TypeTag for the given argument.
// Returns TypeInvalid if the argument type is unsupported.
func ToTypeTag(arg interface{}) TypeTag {
	switch t := arg.(type) {
	case bool:
		if t {
			return TypeTrue
		}
		return TypeFalse
	case nil:
		return TypeNil
	case int32:
		return TypeInt32
	case float32:
		return TypeFloat32
	case string:
		return TypeString
	case []byte:
		return TypeBlob
	case int64:
		return TypeInt64
	case float64:
		return TypeFloat64
	case Timetag:
		return TypeTimeTag
	default:
		return TypeInvalid
	}
}

// GetTypeTag returns the OSC TypeTag string for the given slice.
func GetTypeTag(i []interface{}) (string, error) {
	tt := make([]byte, len(i)+1)
	_, err := writeTypeTags(i, tt)
	if err != nil {
		return "", err
	}
	return string(tt), nil
}
