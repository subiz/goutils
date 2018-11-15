package conv

import "fmt"

var t__, f__ = true, false
var True, False = &t__, &f__

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

func AmpInt32(i int32) *int32 { return &i }

func AmpI64(i int64) *int64 { var ix = int64(i); return &ix }

func B(b bool) *bool { return &b }

func PI32(i int) *int32 {
	i32 := int32(i)
	return &i32
}

func PI64(i int) *int64 {
	i64 := int64(i)
	return &i64
}

func F32(f float32) *float32 {
	return &f
}

func F64(f float64) *float64 {
	return &f
}
