package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogFiltering_for_global_debug_level(t *testing.T) {
	filterString := "service_a:warn,service_b:Error"
	next, logs := observer.New(zapcore.DebugLevel)
	core, _ := NewFilteringCore(next, filterString)

	loggerA := zap.New(core).Named("service_a")
	loggerB := zap.New(core).Named("service_b")
	loggerC := zap.New(core).Named("service_c")

	loggerA.Error("from-a")
	loggerA.Warn("from-a")
	loggerA.Debug("from-a")


	loggerB.Error("from-b")
	loggerB.Warn("from-b")
	loggerB.Debug("from-b")

	loggerC.Error("from-c")
	loggerC.Warn("from-c")
	loggerC.Debug("from-c")


	logLinesA := logs.FilterMessage("from-a").All()
	assert.Equal(t, 2, len(logLinesA))

	logLinesB := logs.FilterMessage("from-b").All()
	assert.Equal(t, 1, len(logLinesB))

	logLinesC := logs.FilterMessage("from-c").All()
	assert.Equal(t, 3, len(logLinesC))
}


func TestLogFiltering_for_global_error_level(t *testing.T) {
	filterString := "service_a:warn,service_b:debug"
	next, logs := observer.New(zapcore.ErrorLevel)
	core, _ := NewFilteringCore(next, filterString)

	loggerA := zap.New(core).Named("service_a")
	loggerB := zap.New(core).Named("service_b")
	loggerC := zap.New(core).Named("service_c")

	loggerA.Error("from-a")
	loggerA.Warn("from-a")
	loggerA.Debug("from-a")

	loggerB.Error("from-b")
	loggerB.Warn("from-b")
	loggerB.Debug("from-b")

	loggerC.Error("from-c")
	loggerC.Warn("from-c")
	loggerC.Debug("from-c")

	logLinesA := logs.FilterMessage("from-a").All()
	assert.Equal(t, 1, len(logLinesA))

	logLinesB := logs.FilterMessage("from-b").All()
	assert.Equal(t, 1, len(logLinesB))

	logLinesC := logs.FilterMessage("from-c").All()
	assert.Equal(t, 1, len(logLinesC))
}


func TestLogFiltering_errors(t *testing.T) {
	filterString := "service_a:warn,service_b:debug"
	next, logs := observer.New(zapcore.ErrorLevel)
	core, _ := NewFilteringCore(next, filterString)

	loggerA := zap.New(core).Named("service_a")
	loggerB := zap.New(core).Named("service_b")
	loggerC := zap.New(core).Named("service_c")

	loggerA.Error("from-a")
	loggerA.Warn("from-a")
	loggerA.Debug("from-a")

	loggerB.Error("from-b")
	loggerB.Warn("from-b")
	loggerB.Debug("from-b")

	loggerC.Error("from-c")
	loggerC.Warn("from-c")
	loggerC.Debug("from-c")

	logLinesA := logs.FilterMessage("from-a").All()
	assert.Equal(t, 1, len(logLinesA))

	logLinesB := logs.FilterMessage("from-b").All()
	assert.Equal(t, 1, len(logLinesB))

	logLinesC := logs.FilterMessage("from-c").All()
	assert.Equal(t, 1, len(logLinesC))
}

