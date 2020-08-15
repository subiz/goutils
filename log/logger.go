package log

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/subiz/goutils/map"
	"github.com/subiz/header"
	compb "github.com/subiz/header/common"
	"github.com/subiz/kafka"
)

type key int

const (
	trace_id key = 2108439
)

type Logger struct {
	hostname string
	pub      *kafka.Publisher
	tags     []string
	service  string
	tm       cmap.Map
}

func clearTimer(tm cmap.Map) { // after one day
	for {
		for _, k := range tm.Keys() {
			t, ok := tm.Get(k)
			if !ok {
				continue
			}
			tim := t.(time.Time)
			if time.Since(tim) > 24*time.Hour {
				tm.Remove(k)
			}
		}
		time.Sleep(1 * time.Hour)
	}
}

func newLogger() *Logger {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = err.Error()
	}
	tm := cmap.New(32)
	go clearTimer(tm)
	return &Logger{tm: tm, tags: []string{}, hostname: hostname}
}

var logger = newLogger()

func (l *Logger) Log(persist bool, ctx context.Context, level compb.Level, v ...interface{}) {
	if len(v) < 1 || (len(v) == 1 && v[0] == nil) {
		return
	}

	var t interface{} = v
	if len(v) == 1 {
		t = v[0]
	}

	message, err := json.Marshal(t)
	if err != nil {
		message = []byte(fmt.Sprintf("%v", t))
	}

	if err, ok := t.(error); ok {
		message = []byte(err.Error())
	}

	echo(t)

	if l.pub == nil {
		return
	}

	// only publish to kafka if persistent is required
	if !persist {
		return
	}
	log := &compb.Log{
		Level:       level.String(),
		Created:     time.Now().UnixNano(),
		Message:     message,
		TraceId:     GetTrace(ctx),
		Tags:        l.tags,
		ServiceName: l.service,
	}
	log.Ctx = &compb.Context{SubTopic: header.E_LogLogRequested.String()}
	l.pub.PublishAsync(header.E_LogRequested.String(), log, -1, GetTrace(ctx))
}

func (l *Logger) log(persist bool, level compb.Level, v ...interface{}) {
	if len(v) == 0 {
		return
	}

	ctx, ok := v[0].(context.Context)
	if ok {
		l.Log(persist, ctx, level, v[1:]...)
	} else {
		l.Log(persist, nil, level, v...)
	}
}

func (l *Logger) Error(ctx context.Context, v ...interface{}) {
	l.Log(true, ctx, compb.Level_error, v...)
}

func (l *Logger) Tags(tags ...string) Logger {
	for _, tag := range tags {
		if !inArray(tag, l.tags) {
			l.tags = append(l.tags, tag)
		}
	}

	return *l
}

func GetStack() []byte {
	s := string(debug.Stack())
	lines := strings.Split(strings.TrimSpace(s), "\n")
	lines = lines[3:] // ignore unnecessary lines
	out := ""
	for i, line := range lines {
		if i%2 == 1 { // filter lines contains file path
			f := removeLastPlusSign(strings.TrimSpace(line))
			f = splitLineNumber(f)
			out += f + "\n"
		}
	}
	return []byte(out)
}

func removeLastPlusSign(s string) string {
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
	return strings.Join(split[0:len(split)-1], ":") + ":" + line
}

// Log print anything to stdout
func echo(v ...interface{}) {
	format := strings.Repeat("%v ", len(v))
	message := fmt.Sprintf(format, v...)
	fmt.Printf("%s %s\n", getCaller(), message)
}

// find outside caller
func getCaller() string {
	_, currentFile, currentLine, _ := runtime.Caller(0)
	for i := 0; i <= 20; i++ {
		_, file, line, _ := runtime.Caller(i)
		if file != currentFile {
			return chopPath(file) + ":" + strconv.Itoa(line)
		}
	}

	return chopPath(currentFile) + " " + strconv.Itoa(currentLine)
}

func inArray(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func Config(brokers []string, serviceName string) {
	if len(brokers) > 0 {
		logger.pub = kafka.NewPublisher(brokers)
	}
	logger.service = serviceName
}

func Trace(ctx context.Context, traceid string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, trace_id, traceid)
}

func GetTrace(ctx context.Context) string {
	if ctx == nil {
		return "<no trace>"
	}
	traceid, _ := ctx.Value(trace_id).(string)
	return traceid
}

func Log(ctx context.Context, level compb.Level, v ...interface{}) {
	logger.Log(false, ctx, level, v...)
}

func Info(v ...interface{}) { logger.log(false, compb.Level_info, v...) }

func Warn(v ...interface{}) { logger.log(false, compb.Level_warning, v...) }

func Fatal(v ...interface{}) { logger.log(true, compb.Level_fatal, v...) }

func Debug(v ...interface{}) { logger.log(false, compb.Level_debug, v...) }

func Error(v ...interface{}) {
	if len(v) == 0 || v[0] == nil {
		return
	}
	logger.log(true, compb.Level_error, v...)
}

func Panic(v ...interface{}) { logger.log(true, compb.Level_panic, v...) }

func Errorf(ctx context.Context, format string, v ...interface{}) {
	logger.Error(ctx, fmt.Sprintf(format, v...))
}

func Time(key string) { logger.tm.Set(key, time.Now()) }

func TimeCheck(key string, labels ...string) {
	t, ok := logger.tm.Get(key)
	if !ok {
		Info("[TIMER] missing key " + key)
		return
	}

	last := ""
	l, ok := logger.tm.Get("+systemcheckmark" + key)
	if ok {
		last = fmt.Sprintf(". +%v", time.Since(l.(time.Time)))
	}
	logger.tm.Set("+systemcheckmark"+key, time.Now())

	Info(fmt.Sprintf("[TIMER] %s %v: %v%s", key, labels, time.Since(t.(time.Time)), last))
}

func TimeEnd(key string) {
	TimeCheck(key)
	logger.tm.Remove(key)
	logger.tm.Remove("+systemcheckmark" + key)
}

func Logf(ctx context.Context, level compb.Level, format string, v ...interface{}) {
	logger.Log(false, ctx, level, fmt.Sprintf(format, v...))
}

func chopPath(path string) string {
	defaultPaths := []string{
		"/src/bitbucket.org/subiz/",
		"/src/git.subiz.net/",
		"/src/github.com/subiz/",
	}
	for _, p := range defaultPaths {
		i := strings.LastIndex(path, p)
		if i >= 0 {
			return path[i+len(p):]
		}
	}
	return path
}

func Assert(a, b interface{}) {
	aa, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	if string(aa) != string(bb) {
		Errorf(context.Background(), "must equal, got: %s, expected: %s", string(aa), string(bb))
	}
}

func NotAssert(a, b interface{}) {
	aa, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	if string(aa) == string(bb) {
		Errorf(context.Background(), "must not equal, got: %s, not expected: %s", string(aa), string(bb))
	}
}
