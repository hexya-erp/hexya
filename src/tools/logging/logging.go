// Copyright 2016 NDP Systèmes. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/src/tools/exceptions"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// log is the base logger of the framework
	log       = &zapLogger{}
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

// A Logger writes logs to a handler
type Logger interface {
	// Panic logs a error level message then panics
	Panic(msg string, ctx ...interface{})
	// Error logs an error level message
	Error(msg string, ctx ...interface{})
	// Warn logs a warning level message
	Warn(msg string, ctx ...interface{})
	// Info logs an information level message
	Info(msg string, ctx ...interface{})
	// Debug logs a debug level message. This may be very verbose
	Debug(msg string, ctx ...interface{})
	// New returns a child logger with the given context
	New(ctx ...interface{}) Logger
	// Sync the logger cache
	Sync() error
}

// zapLogger is an implementation of logger using Uber's zap library
type zapLogger struct {
	zap    *zap.SugaredLogger
	ctx    []interface{}
	parent *zapLogger
}

// Panic logs a error level message then panics
func (l *zapLogger) Panic(msg string, ctx ...interface{}) {
	if l.checkParent() {
		l.zap.Errorw(msg, ctx...)
	}
	panicData := msg + "\n"
	for i := 0; i < len(ctx); i += 2 {
		panicData += fmt.Sprintf("\t%v : %v\n", ctx[i], ctx[i+1])
	}
	panic(panicData)
}

// Error logs an error level message
func (l *zapLogger) Error(msg string, ctx ...interface{}) {
	if !l.checkParent() {
		return
	}
	l.zap.Errorw(msg, ctx...)
}

// Warn logs a warning level message
func (l *zapLogger) Warn(msg string, ctx ...interface{}) {
	if !l.checkParent() {
		return
	}
	l.zap.Warnw(msg, ctx...)
}

// Info logs an information level message
func (l *zapLogger) Info(msg string, ctx ...interface{}) {
	if !l.checkParent() {
		return
	}
	l.zap.Infow(msg, ctx...)
}

// Debug logs a debug level message. This may be very verbose
func (l *zapLogger) Debug(msg string, ctx ...interface{}) {
	if !l.checkParent() {
		return
	}
	l.zap.Debugw(msg, ctx...)
}

// Sync the logger cache
func (l *zapLogger) Sync() error {
	if !l.checkParent() {
		return errors.New("syncing a non-initialized logger")
	}
	return l.zap.Sync()
}

// New returns a child logger with the given context
func (l *zapLogger) New(ctx ...interface{}) Logger {
	return &zapLogger{
		ctx:    ctx,
		parent: l,
	}
}

// checkParent recursively looks for an ancestor with a valid zap logger backend.
//
// If one is found, all children zap loggers are instantiated and checkParent returns true.
// Otherwise, it returns false.
func (l *zapLogger) checkParent() bool {
	if l.zap != nil || l.parent == nil {
		return true
	}
	l.parent.checkParent()
	if l.parent.zap != nil {
		l.zap = l.parent.zap.With(l.ctx...)
		return true
	}
	return false
}

// Initialize starts the base logger used by all Hexya components
func Initialize() {
	logConfig := zap.NewProductionConfig()
	if viper.GetBool("Debug") {
		logConfig = zap.NewDevelopmentConfig()
	}
	logLevel := zap.NewAtomicLevel()
	err := logLevel.UnmarshalText([]byte(viper.GetString("LogLevel")))
	if err != nil {
		fmt.Printf("error while reading log level. Falling back to info. Error: %s\n", err.Error())
		logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	logConfig.Level = logLevel

	var outputPaths []string
	if viper.GetBool("LogStdout") {
		outputPaths = append(outputPaths, "stdout")
	}
	if path := viper.GetString("LogFile"); path != "" {
		outputPaths = append(outputPaths, path)
	}
	logConfig.OutputPaths = outputPaths

	plainLog, err := logConfig.Build()
	if err != nil {
		panic(err)
	}
	log.zap = plainLog.Sugar()

	log.Info("Hexya Starting...")
}

// GetLogger returns a context logger for the given module
func GetLogger(moduleName string) Logger {
	l := log.New("module", moduleName)
	return l
}

// LogPanicData logs the panic data with stacktrace and return an
// error with the panic message. This function is separated from
// LogAndPanic so that unwanted panics can still be logged with
// this function.
func LogPanicData(panicData interface{}) error {
	msg := fmt.Sprintf("%v", panicData)
	log.Error("Hexya panicked", "msg", msg)

	stackTrace := stack(1)
	fullMsg := fmt.Sprintf("%s\n\n%s", msg, stackTrace)
	return exceptions.UserError{
		Message: msg,
		Debug:   fullMsg,
	}
}

// stack returns a nicely formated stack frame, skipping skip frames
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // the returned data
	// As we loop, we open files and read them. These variables record the currently
	// loaded file.
	var lines [][]byte
	var lastFile string
	for i := skip; ; i++ { // Skip the expected number of frames
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		// Print this much at least.  If we can't find the source, it won't show.
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source returns a space-trimmed slice of the n'th line.
func source(lines [][]byte, n int) []byte {
	n-- // in stack trace, lines are 1-indexed but our array is 0-indexed
	if n < 0 || n >= len(lines) {
		return dunno
	}
	return bytes.TrimSpace(lines[n])
}

// function returns, if possible, the name of the function containing the PC.
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())
	// The name includes the path name to the package, which is unnecessary
	// since the file name is already included.  Plus, it has center dots.
	// That is, we see
	//	runtime/debug.*T·ptrmethod
	// and want
	//	*T.ptrmethod
	// Also the package path might contains dot (e.g. code.google.com/...),
	// so first eliminate the path prefix
	if lastslash := bytes.LastIndex(name, slash); lastslash >= 0 {
		name = name[lastslash+1:]
	}
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	name = bytes.Replace(name, centerDot, dot, -1)
	return name
}

// LogForGin returns a gin.HandlerFunc (middleware) that logs requests using Logger.
//
// Requests with errors are logged using log15.Error().
// Requests without errors are logged using log15.Info().
func LogForGin(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this value
		path := c.Request.URL.Path
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		status := c.Writer.Status()

		ctxLogger := logger.New(
			"status", status,
			"method", c.Request.Method,
			"path", path,
			"ip", c.ClientIP(),
			"latency", latency,
		)

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			ctxLogger.Error(c.Errors.String())
		} else if status >= 400 {
			ctxLogger.Warn("HTTP Error")
		} else {
			ctxLogger.Info("")
		}
	}
}
