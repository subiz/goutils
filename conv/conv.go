package conv

import "fmt"

func S(s interface{}) *string {
	if s == nil {
		return S("")
	}
	switch v := s.(type) {
	case []byte:
		b := string(v)
		return &b
	case string:
		return &v
	case fmt.Stringer:
		str := v.String()
		return &str
	default:
		str := fmt.Sprintf("%v", v)
		return &str
	}
}

func I32(i int32) *int32 { return &i }

func I64(i int64) *int64 { return &i }

func B(b bool) *bool { return &b }

func PI32(i int) *int32 { return I32(int32(i)) }

func PI64(i int) *int64 { return I64(int64(i)) }

func F32(f float32) *float32 { return &f }

func F64(f float64) *float64 { return &f }
