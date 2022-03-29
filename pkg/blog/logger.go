/**
@author 陈志银 1981330085@qq.com
@date 2021/11/23
*/

package blog

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ log.Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	log  *zap.Logger
	Sync func() error
}

// NewZapLogger return ZapLogger
func NewZapLogger(lumberJackLogger *lumberjack.Logger, level zap.AtomicLevel, opts ...zap.Option) *ZapLogger {
	encoder := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "name",
		CallerKey:      "file", //结构化（json）输出：打印日志的文件对应的Key
		FunctionKey:    "func",
		StacktraceKey:  "strace",
		EncodeLevel:    zapcore.CapitalLevelEncoder, //将日志级别转换成大写（INFO，WARN，ERROR等）
		EncodeCaller:   zapcore.ShortCallerEncoder,  //采用短文件路径编码输出
		EncodeTime:     zapcore.ISO8601TimeEncoder,  //输出的时间格式
		EncodeDuration: zapcore.MillisDurationEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		zapcore.NewMultiWriteSyncer(
			//zapcore.AddSync(os.Stdout),
			zapcore.AddSync(lumberJackLogger),
		), level)

	opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(3), zap.AddStacktrace(zapcore.ErrorLevel))

	zapLogger := zap.New(core, opts...)

	return &ZapLogger{log: zapLogger, Sync: zapLogger.Sync}
}

// Log Implementation of logger interface
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.log.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", keyvals))
		return nil
	}

	// Zap.Field is used when keyvals pairs appear
	var data []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case log.LevelDebug:
		l.log.Debug("", data...)
	case log.LevelInfo:
		l.log.Info("", data...)
	case log.LevelWarn:
		l.log.Warn("", data...)
	case log.LevelError:
		l.log.Error("", data...)
	}
	return nil
}
