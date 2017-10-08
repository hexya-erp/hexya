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
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/hexya/tools/exceptions"
	"github.com/inconshreveable/log15"
	"github.com/spf13/viper"
)

var (
	// log is the base logger of the framework
	log       *Logger
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

func init() {
	log = NewLogger()
}

// A Logger writes logs to a handler
type Logger struct {
	log15.Logger
}

// NewLogger returns a pointer to a new Logger instance
func NewLogger(ctx ...interface{}) *Logger {
	return &Logger{
		Logger: log15.New(ctx...),
	}
}

// New returns a new Logger that has this logger's context plus the given context
func (l *Logger) New(ctx ...interface{}) *Logger {
	return &Logger{
		Logger: l.Logger.New(ctx...),
	}
}

// Panic logs as an error the given message and context and then panics.
func (l *Logger) Panic(msg string, ctx ...interface{}) {
	pc, _, _, _ := runtime.Caller(1)
	ctx = append(ctx, "caller", string(function(pc)))
	l.Error(msg, ctx...)

	fullMsg := fmt.Sprintf("%s, %v\n", msg, ctx)
	panic(exceptions.UserError{
		Message: msg,
		Debug:   fullMsg,
	})
}

// Initialize starts the base logger used by all Hexya components
func Initialize() {
	logLevel, err := log15.LvlFromString(viper.GetString("LogLevel"))
	if err != nil {
		log.Warn("Error while reading log level. Falling back to info", "error", err.Error())
		logLevel = log15.LvlInfo
	}

	stdoutHandler := log15.DiscardHandler()
	if viper.GetBool("LogStdout") {
		stdoutHandler = log15.StreamHandler(os.Stdout, log15.TerminalFormat())
	}

	fileHandler := log15.DiscardHandler()
	if path := viper.GetString("LogFile"); path != "" {
		fileHandler = log15.Must.FileHandler(path, log15.LogfmtFormat())
	}

	log.SetHandler(
		log15.LvlFilterHandler(
			logLevel,
			log15.MultiHandler(
				stdoutHandler,
				fileHandler,
			),
		),
	)
	log.Info("Hexya Starting...")
}

// GetLogger returns a context logger for the given module
func GetLogger(moduleName string) *Logger {
	l := log.New("module", moduleName)
	l.SetHandler(log15.CallerFuncHandler(l.GetHandler()))
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
	log.Error(fmt.Sprintf("Stack trace:\n%s", stackTrace))

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
func LogForGin(logger *Logger) gin.HandlerFunc {
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
