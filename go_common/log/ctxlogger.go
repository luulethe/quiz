package log

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

// Fields The field of ctxlogger
type Fields map[string]interface{}

// SugarFields The sugar field of ctxlogger
type SugarFields map[string]string

// usedMap record whether a k has been taken, if so the new k would be k',k”,k”' and so on
type usedMap map[string]bool

// Level Level of ctxlogger is exactly the level of zapcore
type Level = zapcore.Level

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel Level = zapcore.DebugLevel

	// InfoLevel is the default logging priority.
	InfoLevel Level = zapcore.InfoLevel

	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel Level = zapcore.WarnLevel

	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel Level = zapcore.ErrorLevel

	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel Level = zapcore.DPanicLevel

	// PanicLevel logs a message, then panics.
	PanicLevel Level = zapcore.PanicLevel

	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel Level = zapcore.FatalLevel
)

type FileConfig map[Level][]string

// Configuration for logging
type Config struct {
	// Level set log level
	Level Level
	// EncodeLogsAsJSON makes the log framework log JSON
	EncodeLogsAsJSON bool
	// FileLoggingEnabled makes the framework log to a file
	// the fields below can be skipped if this value is false!
	FileLoggingEnabled bool
	// ConsoleLoggingEnabled makes the framewor log to console
	ConsoleLoggingEnabled bool
	// CallerEnabled makes the caller log to a file
	CallerEnabled bool
	// CallerSkip increases the number of callers skipped by caller
	CallerSkip int
	// Directory to log to to when filelogging is enabled
	Directory string
	// FileConfig is the config of file to store log
	FileConfig FileConfig
	// MaxSize the max size in MB of the logfile before it's rolled
	MaxSize int
	// MaxBackups the max number of rolled files to keep
	MaxBackups int
	// MaxAge the max age in days to keep a logfile
	MaxAge int
	// SkipRedirectStdLog makes the log framework skip redirecting standard error log
	SkipRedirectStdLog bool
}

// LogLevel The atomic log level for zap logger, can be configued after new logger
var logLevel = zap.NewAtomicLevel()

// Default holds the logger returned by Logger when there is no logger in
// the context. If replacing Default with a new Logger then consider
// using &LogLevel as the LevelEnabler so that SetLevel can still be used
// to dynamically change the logging level.
var Default = NewCtxLogger()

// DefaultContext The defaultContext for ctxlogger
var DefaultContext = WithLogger(context.Background(), Default)

// SetDefaultContext Set the defaultContext
func SetDefaultContext(ctx context.Context) {
	DefaultContext = ctx
}

// NewCtxLogger Return a ctxlogger based
func NewCtxLogger() map[Level]*zap.Logger {
	return map[Level]*zap.Logger{
		DebugLevel: zap.New(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zapcore.EncoderConfig{
					MessageKey:  defaultMessageKey,
					LevelKey:    defaultLevelKey,
					TimeKey:     defaultTimeKey,
					EncodeLevel: zapcore.LowercaseLevelEncoder,
					EncodeTime:  zapcore.ISO8601TimeEncoder,
				}),
				os.Stderr,
				&logLevel,
			),
		),
	}
}

// loggerKey holds the context key used for loggers.
type loggerKey struct{}

// usedMapKey holds the used map for this context.
type usedMapKey struct{}

// fieldKey holds the context key to a map of fields in loggers.
type fieldKey struct{}

// WithLogger returns a new context derived from ctx that
// is associated with the given logger.
func WithLogger(ctx context.Context, loggers map[Level]*zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, loggers)
}

// CopyLogger Copy the logger from srcCtx to dstCtx, return the dstCtx with logger
func CopyLogger(srcCtx context.Context, dstCtx context.Context) context.Context { //nolint:golint
	return context.WithValue(context.WithValue(dstCtx, loggerKey{}, Logger(srcCtx)),
		fieldKey{}, srcCtx.Value(fieldKey{}))
}

// Flush flush the buffer to disk
func Flush(ctx context.Context) {
	for _, logger := range Logger(ctx) {
		_ = logger.Sync()
	}
}

// DFlush flush the default buffer to disk
func DFlush() {
	for _, logger := range Logger(DefaultContext) {
		_ = logger.Sync()
	}
}

// updateUsedMap Update the usedMap for ctx, return a new context and new usedMap
func updateUsedMap(ctx context.Context, fields Fields) (context.Context, []zapcore.Field) {
	var used usedMap = make(usedMap)
	if usedMapVal := ctx.Value(usedMapKey{}); usedMapVal != nil {
		for k, v := range usedMapVal.(usedMap) {
			used[k] = v
		}
	}
	var storedFields Fields = make(Fields, len(fields))
	if stored := ctx.Value(fieldKey{}); stored != nil {
		for k, v := range stored.(Fields) {
			storedFields[k] = v
		}
	}
	newFields := make([]zapcore.Field, 0, len(fields))
	for k, v := range fields {
		newK := k
		for used[newK] {
			newK += "+"
		}
		switch val := v.(type) {
		case uint64:
			newFields = append(newFields, zap.String(newK, strconv.FormatUint(val, 10)))
		default:
			newFields = append(newFields, zap.Any(newK, v))
		}
		used[newK] = true
		storedFields[newK] = v
	}
	ctx = context.WithValue(ctx, usedMapKey{}, used)
	ctx = context.WithValue(ctx, fieldKey{}, storedFields)
	return ctx, newFields
}

// convertFields updates the usedMap for ctx, and converts ctxlogger field to zapcore.Field
func convertFields(ctx context.Context, fields Fields) []zapcore.Field {
	var used usedMap
	usedMapVal := ctx.Value(usedMapKey{})
	if usedMapVal == nil {
		used = make(usedMap)
	} else {
		used = usedMapVal.(usedMap)
	}

	tmpMap := make(usedMap)
	newFields := make([]zapcore.Field, 0, len(fields))
	for k, v := range fields {
		newK := k
		for used[newK] || tmpMap[newK] {
			newK += "+"
		}
		switch val := v.(type) {
		case uint64:
			newFields = append(newFields, zap.String(newK, strconv.FormatUint(val, 10)))
		default:
			newFields = append(newFields, zap.Any(newK, v))
		}
		tmpMap[newK] = true
	}
	return newFields
}

func GetFieldValue(ctx context.Context, key string) interface{} {
	fields := ctx.Value(fieldKey{})
	if fields == nil {
		return nil
	}
	return fields.(Fields)[key]
}

func GetSugarFields(ctx context.Context, keys ...string) SugarFields {
	sugar := make(SugarFields, len(keys))
	fields := ctx.Value(fieldKey{})
	if fields == nil {
		return sugar
	}
	for _, key := range keys {
		if field, ok := fields.(Fields)[key]; ok {
			sugar[key] = fmt.Sprintf("%v", field)
		}
	}
	return sugar
}

// WithFields returns a new context derived from ctx
// that has a logger that always logs the given fields.
func WithFields(ctx context.Context, fields Fields) context.Context {
	ctx, newFields := updateUsedMap(ctx, fields)
	newLoggers := map[Level]*zap.Logger{}
	for l, logger := range Logger(ctx) {
		newLoggers[l] = logger.With(newFields...)
	}
	return WithLogger(ctx, newLoggers)
}

// WithSugarFields returns a new context derived from ctx
func WithSugarFields(ctx context.Context, sugar SugarFields) context.Context {
	fields := make([]zap.Field, 0, len(sugar))
	for k, v := range sugar {
		fields = append(fields, zap.String(k, v))
	}
	newLoggers := map[Level]*zap.Logger{}
	for l, logger := range Logger(ctx) {
		newLoggers[l] = logger.With(fields...)
	}
	return WithLogger(ctx, newLoggers)
}

// Logger returns the logger associated with the given
// context. If there is no logger, it will return Default.
func Logger(ctx context.Context) map[Level]*zap.Logger {
	if ctx == nil {
		if ctx != DefaultContext {
			return Logger(DefaultContext)
		}
		return Default
	}
	if logger, _ := ctx.Value(loggerKey{}).(map[Level]*zap.Logger); logger != nil {
		return logger
	}
	if ctx != DefaultContext {
		return Logger(DefaultContext)
	}
	return Default
}

// LoggerForLevel returns the logger associated with the given
// context and level
func LoggerForLevel(ctx context.Context, level Level) *zap.Logger {
	loggerMap := Logger(ctx)
	for ; level >= DebugLevel; level-- {
		if logger, ok := loggerMap[level]; ok {
			return logger
		}
	}
	return Default[DebugLevel]
}

// Debug calls Logger(ctx).Debug(msg, fields...).
func Debug(ctx context.Context, msg string) {
	LoggerForLevel(ctx, DebugLevel).Debug(msg)
}

// Debugf calls Logger(ctx).Debug(msg, fields...).
func Debugf(ctx context.Context, template string, args ...interface{}) {
	LoggerForLevel(ctx, DebugLevel).Debug(fmt.Sprintf(template, args...))
}

// Debugff Transfer a format expression to json expression
// Example Data|Test|uid:%d|gid:%d|rid:%d -> {"msg": "Data|Test", "uid":uid, "gid":gid|"rid":rid}
func Debugff(ctx context.Context, template string, args ...interface{}) {
	msg, fields := transferFormatToMap(template, args...)
	LoggerForLevel(ctx, DebugLevel).Debug(msg, convertFields(ctx, fields)...)
}

// DebugWith calls Logger(ctx).Debug(msg, fields...).
func DebugWith(ctx context.Context, msg string, fields Fields) {
	if len(fields) > 0 {
		LoggerForLevel(ctx, DebugLevel).Debug(msg, convertFields(ctx, fields)...)
	} else {
		LoggerForLevel(ctx, DebugLevel).Debug(msg)
	}
}

// Info calls Logger(ctx).Info(msg, fields...).
func Info(ctx context.Context, msg string) {
	LoggerForLevel(ctx, InfoLevel).Info(msg)
}

// Infof Log a format message at the info level
func Infof(ctx context.Context, template string, args ...interface{}) {
	LoggerForLevel(ctx, InfoLevel).Info(fmt.Sprintf(template, args...))
}

// Infoff Transfer a format expression to json expression
// Example Data|Test|uid:%d|gid:%d|rid:%d -> {"msg": "Data|Test", "uid":uid, "gid":gid|"rid":rid}
func Infoff(ctx context.Context, template string, args ...interface{}) {
	msg, fields := transferFormatToMap(template, args...)
	LoggerForLevel(ctx, InfoLevel).Info(msg, convertFields(ctx, fields)...)
}

// InfoWith Log a message with fields at the info level
func InfoWith(ctx context.Context, msg string, fields Fields) {
	if len(fields) > 0 {
		LoggerForLevel(ctx, InfoLevel).Info(msg, convertFields(ctx, fields)...)
	} else {
		LoggerForLevel(ctx, InfoLevel).Info(msg)
	}
}

// Warn calls Logger(ctx).Warn(msg, fields...).
func Warn(ctx context.Context, msg string) {
	LoggerForLevel(ctx, WarnLevel).Warn(msg)
}

// Warnf Log a format message at the info level
func Warnf(ctx context.Context, template string, args ...interface{}) {
	LoggerForLevel(ctx, WarnLevel).Warn(fmt.Sprintf(template, args...))
}

// Warnff Log a format message at the info level
func Warnff(ctx context.Context, template string, args ...interface{}) {
	msg, fields := transferFormatToMap(template, args...)
	LoggerForLevel(ctx, WarnLevel).Warn(msg, convertFields(ctx, fields)...)
}

// WarnWith Log a message with fields at the info level
func WarnWith(ctx context.Context, msg string, fields Fields) {
	if len(fields) > 0 {
		LoggerForLevel(ctx, WarnLevel).Warn(msg, convertFields(ctx, fields)...)
	} else {
		LoggerForLevel(ctx, WarnLevel).Warn(msg)
	}
}

// Error calls Logger(ctx).Error(msg, fields...).
func Error(ctx context.Context, msg string) {
	LoggerForLevel(ctx, ErrorLevel).Error(msg)
}

// Errorf calls Logger(ctx).Error(msg, fields...).
func Errorf(ctx context.Context, template string, args ...interface{}) {
	LoggerForLevel(ctx, ErrorLevel).Error(fmt.Sprintf(template, args...))
}

// Errorff Transfer a format expression to json expression
// Example Data|Test|uid:%d|gid:%d|rid:%d -> {"msg": "Data|Test", "uid":uid, "gid":gid|"rid":rid}
func Errorff(ctx context.Context, template string, args ...interface{}) {
	msg, fields := transferFormatToMap(template, args...)
	LoggerForLevel(ctx, ErrorLevel).Error(msg, convertFields(ctx, fields)...)
}

// ErrorWith calls Logger(ctx).Error(msg, fields...).
func ErrorWith(ctx context.Context, msg string, fields Fields) {
	if len(fields) > 0 {
		LoggerForLevel(ctx, ErrorLevel).Error(msg, convertFields(ctx, fields)...)
	} else {
		LoggerForLevel(ctx, ErrorLevel).Error(msg)
	}
}

// Fatal calls Logger(ctx).Fatal(msg, fields...).
func Fatal(ctx context.Context, msg string) {
	LoggerForLevel(ctx, FatalLevel).Fatal(msg)
}

// Fatalff calls Logger(ctx).Fatal(msg, fields...).
func Fatalff(ctx context.Context, template string, args ...interface{}) {
	msg, fields := transferFormatToMap(template, args...)
	LoggerForLevel(ctx, FatalLevel).Fatal(msg, convertFields(ctx, fields)...)
}

func transferFormatToMap(template string, args ...interface{}) (string, Fields) {
	fields := Fields{}
	strs := strings.Split(template, "|")
	msgField := ""
	id := 0
	for i := 0; i < len(strs); i++ {
		if !strings.Contains(strs[i], ":%") {
			if msgField == "" {
				msgField = strs[i]
			} else {
				msgField += fmt.Sprintf("|%s", strs[i])
			}
		} else {
			if id < len(args) {
				kv := strings.Split(strs[i], ":%")
				if len(kv) != 2 {
					continue
				}
				fields[kv[0]] = args[id]
				id++
			}
		}
	}
	return msgField, fields
}

// Configure sets up the logging Context
func Configure(ctx context.Context, config Config) (context.Context, error) {
	writersMap := map[Level][]zapcore.WriteSyncer{}
	if config.FileLoggingEnabled {
		var err error
		writersMap, err = newRollingFile(config)
		if err != nil {
			return ctx, err
		}
	}
	if config.ConsoleLoggingEnabled || len(writersMap) == 0 {
		writersMap[DebugLevel] = append(writersMap[DebugLevel], os.Stderr)
	}

	loggers := map[Level]*zap.Logger{}
	for l, writers := range writersMap {
		loggers[l] = newZapLogger(
			config.Level,
			config.EncodeLogsAsJSON,
			config.CallerEnabled,
			config.CallerSkip,
			zapcore.NewMultiWriteSyncer(writers...))
		if !config.SkipRedirectStdLog {
			zap.RedirectStdLog(loggers[l].WithOptions(zap.AddCallerSkip(-1 - config.CallerSkip)))
		}
	}
	return WithLogger(ctx, loggers), nil
}

func newRollingFile(config Config) (map[Level][]zapcore.WriteSyncer, error) {
	if err := os.MkdirAll(config.Directory, 0755); err != nil {
		return nil, err
	}

	result := map[Level][]zapcore.WriteSyncer{}
	for level, fileNames := range config.FileConfig {
		for _, fileName := range fileNames {
			result[level] = append(result[level], zapcore.AddSync(&lumberjack.Logger{
				Filename:   path.Join(config.Directory, fileName),
				MaxSize:    config.MaxSize,    // megabytes
				MaxAge:     config.MaxAge,     // days
				MaxBackups: config.MaxBackups, // files
			}))
		}
	}

	return result, nil
}

// TimeEncoder serializes a time.Time to an formatted string
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func newZapLogger(lvl Level, encodeAsJSON, enableCaller bool, callerSkip int, output zapcore.WriteSyncer) *zap.Logger {
	var encoder zapcore.Encoder
	encCfg := zapcore.EncoderConfig{
		TimeKey:  defaultTimeKey,
		LevelKey: defaultLevelKey,
		//ConsoleSeparator: defaultConsoleSeparator,
		NameKey:        "logger",
		CallerKey:      "caller_func",
		MessageKey:     defaultMessageKey,
		StacktraceKey:  "stack_trace",
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.NanosDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if encodeAsJSON {
		encoder = zapcore.NewJSONEncoder(encCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encCfg)
	}
	if enableCaller {
		return zap.New(zapcore.NewCore(encoder, output, zap.NewAtomicLevelAt(lvl)), zap.AddCaller(), zap.AddCallerSkip(callerSkip))
	}
	return zap.New(zapcore.NewCore(encoder, output, zap.NewAtomicLevelAt(lvl)))
}

const (
	defaultMessageKey       = "msg"
	defaultLevelKey         = "level"
	defaultTimeKey          = "@t"
	defaultConsoleSeparator = "|"
)
