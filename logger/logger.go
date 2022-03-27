package logger

import (
	"io"
	"os"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// var zaplog *zap.SugaredLogger
func Init(loggerPath ...string) *zap.Logger {
	path := "./logs"
	if len(loggerPath) > 0 {
		path = loggerPath[0]
	}
	encoder := getEncoder()
	infoWrite := getLogWriter(path, "info", 7)
	errorWrite := getLogWriter(path, "error", 7)

	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})
	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(infoWrite), infoLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(errorWrite), warnLevel),
	)

	return zap.New(core, zap.AddCaller())
	//  zaplog = logger.Sugar()
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("[2006-01-02 15:04:05]"))
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = customTimeEncoder
	// encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(logPath, level string, save uint) io.Writer {
	logFullPath := path.Join(logPath, level)
	hook, err := rotatelogs.New(
		logFullPath+".%Y%m%d%H",                   // 没有使用go风格反人类的format格式
		rotatelogs.WithLinkName(logFullPath),      // 生成软链，指向最新日志文件
		rotatelogs.WithRotationCount(save),        // 文件最大保存份数
		rotatelogs.WithRotationTime(24*time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		panic(err)
	}
	return hook
}
