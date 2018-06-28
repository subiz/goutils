package grpc

import (
	"bitbucket.org/subiz/header/lang"
	"fmt"
	json "github.com/pquerna/ffjson/ffjson"
	"hash/crc32"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

// Err respresent a error
type Error struct {
	Description  string
	DebugMessage string
	Code         lang.T
	Hash         string
	Class        int
	Stack        string
}

var crc32q = crc32.MakeTable(0xD5828281)

func (me *Error) Error() string {
	b, err := json.Marshal(me)
	if err != nil {
		return "#ERRX " + err.Error() + "(" + strings.Join(strings.Split(string(getStack()), "\n"), "|") + ")"
	}
	return "#ERR " + string(b)
}

func FromError(err string) *Error {
	if !strings.HasPrefix(err, "#ERR ") {
		return &Error{
			Description: fmt.Sprintf("%v", err),
			Stack:       strings.Join(strings.Split(string(getStack()), "\n")[2:], "\n"),
			Code:        lang.T_internal_error,
			Class:       InternalErrClass,
		}
	}
	e := &Error{}
	if err := json.Unmarshal([]byte(err[len("#ERR "):]), e); err != nil {
		return &Error{
			Description: fmt.Sprintf("%v", err),
			Stack:       strings.Join(strings.Split(string(getStack()), "\n")[2:], "\n"),
			Code:        lang.T_internal_error,
			Class:       InternalErrClass,
		}
	}
	return e
}

// InternalErrDesc abc
const (
	UnknownErrorDesc     = "unknown error"
	InputInvalidErrDesc  = "user input invalid error"
	InternalErrDesc      = "internal error"
	InternalErrClass     = 500
	InputInvalidErrClass = 400
)

func New500(code lang.T, v ...interface{}) error {
	var format, message string
	if len(v) == 0 {
		format = ""
	} else {
		var ok bool
		format, ok = v[0].(string)
		if !ok {
			format = "%v"
		} else {
			v = v[1:]
		}
	}

	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	stack := getStack()
	// convert error to internal error
	return &Error{
		Hash:        fmt.Sprintf("%08x", crc32.Checksum(stack, crc32q)),
		Description:  message,
		Stack:        fmt.Sprintf("%s", debug.Stack()),
		Code:         code,
		Class:        InternalErrClass,
	}
}

func To400(err error, code lang.T, v ...interface{}) error {
	var format, message string
	if len(v) == 0 {
		format = ""
	} else {
		var ok bool
		format, ok = v[0].(string)
		if !ok {
			format = "%v"
		} else {
			v = v[1:]
		}
	}

	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	// convert error to internal error
	return &Error{
		Description: message,
		//DebugMessage: message,
		Stack: fmt.Sprintf("%s", debug.Stack()),
		Code:  code,
		Class: InputInvalidErrClass,
	}
}

func New400(code lang.T, v ...interface{}) error {
	var format, message string
	if len(v) == 0 {
		format = ""
	} else {
		var ok bool
		format, ok = v[0].(string)
		if !ok {
			format = "%v"
		} else {
			v = v[1:]
		}
	}

	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	stack := getStack()
	// convert error to internal error
	return &Error{
		Description: message,
		Hash:        fmt.Sprintf("%08x", crc32.Checksum(stack, crc32q)),
		Stack:       fmt.Sprintf("%s", stack),
		Code:        code,
		Class:       InputInvalidErrClass,
	}
}

func ToError(r interface{}) error {
	if r == nil {
		return nil
	}
	perr, ok := r.(*Error)
	if ok {
		return perr
	}
	return &Error{
		Description:  fmt.Sprintf("%v", r),
		DebugMessage: "",
		Stack:        getMinifiedStack(),
		Code:         lang.T_internal_error,
		Class:        InternalErrClass,
	}
}

/*
// TODO: remove this
func NewApikeyRequiredInterceptor(apikey string) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (_ interface{}, err error) {
		cred := ExtractCredFromCtx(ctx)
		if cred == nil {
			panic(&Error{
				Description: "wrong api key",
				DebugMessage: "wrong api key: " + apikey,
				Code: "invalid_apikey",
				Class: InputInvalidErrClass,
			})
		}
		return handler(ctx, req)
	})
}
*/
func isInternalErr(e error) bool {
	perr, ok := e.(*Error)
	return ok && perr.Class == InternalErrClass
}

func getMinifiedStack() string {
	stack := ""
	for i := 3; i < 90; i++ {
		_, fn, line, _ := runtime.Caller(i)
		if fn == "" {
			break
		}
		hl := false // highlight
		if strings.Contains(fn, "bitbucket.org/subiz") {
			hl = true
		}
		var split = strings.Split(fn, string(os.PathSeparator))
		var n int

		if len(split) >= 2 {
			n = len(split) - 2
		} else {
			n = len(split)
		}
		fn = strings.Join(split[n:], string(os.PathSeparator))
		if hl {
			stack += fmt.Sprintf("\n→ %s:%d", fn, line)
		} else {
			stack += fmt.Sprintf("\n→ %s:%d", fn, line)
		}
	}
	return stack
}

func E400(err error, code lang.T, v ...interface{}) *Error {
	return erro(InputInvalidErrClass, err, code, v...)
}

func E500(err error, code lang.T, v ...interface{}) *Error {
	return erro(InternalErrClass, err, code, v...)
}

func erro(class int, err error, code lang.T, v ...interface{}) *Error {
	if err == nil {
		return nil
	}

	var format, message string
	if len(v) == 0 {
		format = ""
	} else {
		var ok bool
		format, ok = v[0].(string)
		if !ok {
			format = "%v"
		} else {
			v = v[1:]
		}
	}

	if len(v) > 0 {
		message = fmt.Sprintf(format, v...)
	} else {
		message = format
	}

	if perr, ok := err.(*Error); ok {
		if message != "" {
			message = "\n" + message
		}
		return &Error{
			Description:  perr.Description,
			Hash:         perr.Hash,
			DebugMessage: perr.DebugMessage + message,
			Stack:        perr.Stack,
			Code:         code,
			Class:        class,
		}
	}

	stack := getStack()
	return &Error{
		Hash:         fmt.Sprintf("%08x", crc32.Checksum(stack, crc32q)),
		Description:  fmt.Sprintf("%v", err),
		DebugMessage: message,
		Stack:        strings.Join(strings.Split(string(stack), "\n")[2:], "\n"),
		Code:         code,
		Class:        class,
	}
}

func getStack() []byte {
	s := string(debug.Stack())
	sp := strings.Split(s, "\n")
	sp = sp[9:]
	out := ""
	for i := range sp {
		if i%2 == 1 {
			f := removelastplussign(strings.TrimSpace(sp[i]))
			f = splitLineNumber(f)
			out += f + "\n"
		}
	}
	return []byte(out)
}

func removelastplussign(s string) string {
	split := strings.Split(s, " ")
	if len(split) < 2 {
		return s
	}
	if !strings.HasPrefix(split[len(split)-1], "+0x") {
		return s
	}
	return strings.Join(split[0:len(split)-1], " ")
}

func splitLineNumber(s string) string {
	split := strings.Split(s, ":")
	if len(split) < 2 {
		return s
	}

	line := split[len(split)-1]
	return strings.Join(split[0:len(split)-1], ":") + " " + line
}
