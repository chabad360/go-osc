package osc

type TypeTag rune

const (
	String  TypeTag = 's'
	Int32   TypeTag = 'i'
	Int64   TypeTag = 'h'
	Float32 TypeTag = 'f'
	Float64 TypeTag = 'd'
	Blob    TypeTag = 'b'
	TimeTag TypeTag = 't'
	Nil     TypeTag = 'N'
	True    TypeTag = 'T'
	False   TypeTag = 'F'
)

// ToTypeTag returns the OSC type tag for the given argument.
// Returns Null if the type is unsupported.
func ToTypeTag(arg interface{}) TypeTag {
	switch t := arg.(type) {
	case bool:
		if t {
			return True
		}
		return False
	case nil:
		return Nil
	case int32:
		return Int32
	case float32:
		return Float32
	case string:
		return String
	case []byte:
		return Blob
	case int64:
		return Int64
	case float64:
		return Float64
	case Timetag:
		return TimeTag
	default:
		return 0
	}
}

func GetTypeTag(i []interface{}) (string, error) {
	tt := make([]byte, len(i)+1)
	_, err := writeTypeTags(i, tt)
	if err != nil {
		return "", err
	}
	return string(tt), nil
}
