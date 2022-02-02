package osc

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
