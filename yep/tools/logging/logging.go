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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-stack/stack"
	"github.com/inconshreveable/log15"
	"github.com/spf13/viper"
)

var log log15.Logger

func init() {
	log = log15.New()
}

// Initialize starts the base logger used by all YEP components
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
	log.Info("Yep Starting...")
}

// GetLogger returns a context logger for the given module
func GetLogger(moduleName string) log15.Logger {
	l := log.New("module", moduleName)
	l.SetHandler(log15.CallerFuncHandler(l.GetHandler()))
	return l
}

// LogAndPanic is a helper function for logging an error message on
// the given logger and then panic with the same error message.
func LogAndPanic(log log15.Logger, msg string, ctx ...interface{}) {
	caller := stack.Caller(1)
	ctx = append(ctx, "caller", fmt.Sprintf("%+n", caller))
	log.Error(msg, ctx...)

	fullMsg := fmt.Sprintf("%s, %v\n", msg, ctx)
	panic(fullMsg)
}

// LogPanicData logs the panic data with stacktrace and return an
// error with the panic message. This function is separated from
// LogAndPanic so that unwanted panics can still be logged with
// this function.
func LogPanicData(panicData interface{}) error {
	caller := stack.Caller(1)
	ctx := []interface{}{"caller", fmt.Sprintf("%+n", caller)}

	msg := fmt.Sprintf("%v", panicData)
	log.Error(msg, ctx...)

	trace := stack.Trace()
	var traceStr string
	for _, call := range trace {
		log.Debug(fmt.Sprintf("%+v", call))
		log.Debug(fmt.Sprintf(">> %n", call))
		traceStr += fmt.Sprintf("%+v\n>> %n\n", call, call)
	}

	fullMsg := fmt.Sprintf("%s, %v\n\n%s", msg, ctx, traceStr)
	return errors.New(fullMsg)
}

// Log15ForGin returns a gin.HandlerFunc (middleware) that logs requests using log15.
//
// Requests with errors are logged using log15.Error().
// Requests without errors are logged using log15.Info().
func Log15ForGin(logger log15.Logger) gin.HandlerFunc {
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
			"user-agent", c.Request.UserAgent(),
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
