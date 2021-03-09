package logger

import (
  "fmt"
  "github.com/pkg/errors"
  "go.uber.org/zap/zapcore"
  "strings"
)


type logFilteringSettings struct {
  serviceLevels map[string]zapcore.Level
}

func newSettings() logFilteringSettings {
  return logFilteringSettings{make(map[string]zapcore.Level)}
}

func (settings logFilteringSettings) Enabled(ent zapcore.Entry) bool {
  serviceLevel, exists := settings.serviceLevels[ent.LoggerName]
  if exists {
    return serviceLevel.Enabled(ent.Level)
  } else {
    return true
  }
}

type filteringCore struct {
  core zapcore.Core
  filter logFilteringSettings
}


func NewFilteringCore(next zapcore.Core, logFiltering string) (zapcore.Core, error) {
  filter, err := parseLogFilteringSetting(logFiltering)
  if err != nil {
    return nil, err
   //log.Fatalf("Failed to parse logging setting: %s", err)
  }
  return &filteringCore{next, filter}, nil
}

func parseLogFilteringSetting(logFiltering string) (logFilteringSettings, error) {
  settings := newSettings()
  parts := strings.Split(logFiltering, ",")
  for _, part := range parts {
    pair := strings.Split(part, ":")
    if len(pair) != 2 || pair[0] == "" || pair[1] == "" {
      return settings, errors.New(fmt.Sprintf("Invalid syntax for a logging setting: %s", part))
    }

    serviceName := pair[0]
    levelString := pair[1]

    if _, exists := settings.serviceLevels[serviceName]; exists {
      return settings, errors.New(fmt.Sprintf("Service: %s appears multiple times in the settings", serviceName ))
    }

    var level zapcore.Level
    err := level.UnmarshalText([]byte(levelString))

    if err != nil {
      return settings, errors.New(fmt.Sprintf("Invalid log level string (%s) for service: %s - %s", levelString, serviceName, err.Error()))
    }

    settings.serviceLevels[serviceName] = level
  }
  return settings, nil
}

func (c *filteringCore) With(fields []zapcore.Field) zapcore.Core {
  return &filteringCore{c.core.With(fields), c.filter}
}

func (c *filteringCore) Enabled(lvl zapcore.Level) bool {
  return true
}

func (c *filteringCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
  if !c.filter.Enabled(ent) {
    return ce
  }

  return c.core.Check(ent, ce)
}

func (c *filteringCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
  return c.core.Write(ent, fields)
}

func (c *filteringCore) Sync() error {
  return c.core.Sync()
}
