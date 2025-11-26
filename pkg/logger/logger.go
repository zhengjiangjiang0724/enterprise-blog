package logger

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(level, logFile string) error {
	// 解析日志级别
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// 配置时间格式
	zerolog.TimeFieldFormat = time.RFC3339

	// 如果指定了日志文件，则同时输出到文件和控制台
	if logFile != "" {
		// 确保日志目录存在
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			return err
		}

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}

		// 多写入器：同时输出到文件和控制台
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		multi := zerolog.MultiLevelWriter(consoleWriter, file)
		log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	} else {
		// 仅输出到控制台
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	return nil
}

func GetLogger() zerolog.Logger {
	return log.Logger
}

