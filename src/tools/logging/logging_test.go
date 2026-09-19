// Copyright 2026 Nicolas Piganeau. All Rights Reserved.
// See LICENSE file for full licensing details.

package logging

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hexya-erp/hexya/src/tools/exceptions"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// logFile is the file to which the base logger writes during the tests
var logFile string

// TestMain sets up the log file used by all the tests of this package.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "hexya-logging-test")
	if err != nil {
		panic(err)
	}
	logFile = filepath.Join(dir, "hexya.log")
	res := m.Run()
	os.RemoveAll(dir)
	os.Exit(res)
}

// readLog returns the content of the test log file
func readLog(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(logFile)
	assert.Nil(t, err)
	return string(data)
}

func TestUninitializedLogger(t *testing.T) {
	// We use a standalone logger hierarchy so that this test does not depend
	// on the initialization state of the base logger.
	l := (&zapLogger{}).New("module", "uninitialized")
	t.Run("Logging with an uninitialized logger should not fail", func(t *testing.T) {
		assert.NotPanics(t, func() {
			l.Debug("debug")
			l.Info("info")
			l.Warn("warn")
			l.Error("error")
		})
	})
	t.Run("Syncing an uninitialized logger should return an error", func(t *testing.T) {
		assert.NotNil(t, l.Sync())
	})
	t.Run("Panic should panic even with an uninitialized logger", func(t *testing.T) {
		assert.PanicsWithValue(t, "boom\n\tkey : value\n", func() {
			l.Panic("boom", "key", "value")
		})
	})
}

func TestInitialize(t *testing.T) {
	viper.Set("Debug", false)
	viper.Set("LogLevel", "debug")
	viper.Set("LogStdout", false)
	viper.Set("LogFile", logFile)
	assert.NotPanics(t, Initialize)
	assert.Contains(t, readLog(t), "Hexya Starting...")
}

func TestLoggerLevels(t *testing.T) {
	l := GetLogger("test-levels")
	l.Debug("debug message", "key", "value")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")
	assert.Nil(t, l.Sync())
	content := readLog(t)
	assert.Contains(t, content, "debug message")
	assert.Contains(t, content, "info message")
	assert.Contains(t, content, "warn message")
	assert.Contains(t, content, "error message")
	assert.Contains(t, content, "test-levels")
}

func TestChildLoggers(t *testing.T) {
	child := GetLogger("parent").New("sub", "child")
	child.Info("child message")
	assert.Nil(t, child.Sync())
	content := readLog(t)
	assert.Contains(t, content, "child message")
	assert.Contains(t, content, "child")
}

func TestPanicLogger(t *testing.T) {
	assert.PanicsWithValue(t, "panic message\n\tkey : value\n", func() {
		GetLogger("test-panic").Panic("panic message", "key", "value")
	})
	assert.Contains(t, readLog(t), "panic message")
}

func TestLogPanicData(t *testing.T) {
	err := LogPanicData("something went wrong")
	assert.NotNil(t, err)
	uError, ok := err.(exceptions.UserError)
	assert.True(t, ok)
	assert.Equal(t, "something went wrong", uError.Message)
	assert.Contains(t, uError.Debug, "something went wrong")
	assert.Contains(t, uError.Debug, "logging_test.go")
	assert.Contains(t, uError.Debug, "TestLogPanicData")
	assert.Contains(t, readLog(t), "Hexya panicked")
}

func TestStackHelpers(t *testing.T) {
	t.Run("source returns the trimmed line or ??? when out of range", func(t *testing.T) {
		lines := [][]byte{[]byte("  first  "), []byte("second")}
		assert.Equal(t, "first", string(source(lines, 1)))
		assert.Equal(t, "second", string(source(lines, 2)))
		assert.Equal(t, "???", string(source(lines, 0)))
		assert.Equal(t, "???", string(source(lines, 3)))
	})
	t.Run("function returns ??? for an unknown PC", func(t *testing.T) {
		assert.Equal(t, "???", string(function(0)))
	})
	t.Run("stack returns the current stack trace", func(t *testing.T) {
		assert.Contains(t, string(stack(0)), "logging_test.go")
	})
}

func TestLogForGin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LogForGin(GetLogger("gin")))
	router.GET("/ok", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	router.GET("/notfound", func(c *gin.Context) {
		c.String(http.StatusNotFound, "not found")
	})
	router.GET("/error", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.String(http.StatusInternalServerError, "error")
	})

	for _, path := range []string{"/ok", "/notfound", "/error"} {
		resp := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		assert.NotPanics(t, func() { router.ServeHTTP(resp, req) })
	}
	content := readLog(t)
	assert.Contains(t, content, "/ok")
	assert.Contains(t, content, "HTTP Error")
	assert.True(t, strings.Contains(content, assert.AnError.Error()))
}
