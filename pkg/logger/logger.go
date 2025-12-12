package logger

import (
	"fmt"
	"log/syslog"
	"shiroxy/pkg/models"
	"time"
)

func StartLogger(logConfig *models.Logging) (*Logger, error) {

	logger := Logger{}
	if logConfig != nil {
		logger.logConfig = logConfig
		if logConfig.EnableRmote {
			err := logger.startSyslogServer(logConfig.RemoteBind.Host, logConfig.RemoteBind.Port)
			if err != nil {
				return nil, err
			}
		}
	}

	return &logger, nil
}

type Logger struct {
	sysLogger *syslog.Writer
	logConfig *models.Logging
}

func (l *Logger) InjectLogConfig(logConfig *models.Logging) error {
	if logConfig != nil {
		l.logConfig = logConfig
		if logConfig.EnableRmote {
			err := l.startSyslogServer(logConfig.RemoteBind.Host, logConfig.RemoteBind.Port)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *Logger) LogError(content any, packageName, moduleName string) {
	l.handleLog(content, "error", packageName, moduleName)
}

func (l *Logger) LogSuccess(content any, packageName, moduleName string) {
	l.handleLog(content, "success", packageName, moduleName)
}

func (l *Logger) LogWarning(content any, packageName, moduleName string) {
	l.handleLog(content, "warning", packageName, moduleName)
}

func (l *Logger) Log(content any, packageName, moduleName string) {
	l.handleLog(content, "", packageName, moduleName)
}

func (l *Logger) handleLog(content any, event string, packageName string, moduleName string) error {
	currentTime := time.Now()
	formattedDateTime := currentTime.Format("02/01/2006 15:04:05")

	formattedLog := "[" + formattedDateTime + "] [" + packageName + "] [" + moduleName + "] => " + content.(string)

	if l.logConfig != nil {
		if l.logConfig.Enable {
			fmt.Print("\n[" + formattedDateTime + "]")
			fmt.Print(" [" + packageName + "]")
			fmt.Print(" [" + moduleName + "]")
			fmt.Print(" => ")
			if event == "success" {
				GreenPrint(content)
			} else if event == "error" {
				RedPrint(content)
			} else if event == "warning" {
				YellowPrint(content)
			} else {
				fmt.Printf("%v", content)
			}
		}

		if l.logConfig.EnableRmote {
			if event == "success" {
				err := l.sysLogger.Notice(formattedLog)
				if err != nil {
					return err
				}
			} else if event == "error" {
				err := l.sysLogger.Err(formattedLog)
				if err != nil {
					return err
				}
			} else if event == "warning" {
				err := l.sysLogger.Warning(formattedLog)
				if err != nil {
					return err
				}
			} else {
				err := l.sysLogger.Info(formattedLog)
				if err != nil {
					return err
				}
			}
		}
	} else {
		BluePrint("\n[" + formattedDateTime + "]")
		BluePrint(" [" + packageName + "]")
		BluePrint(" [" + moduleName + "]")
		CyanPrint(" => ")
		switch event {
		case "success":
			GreenPrint(content)
		case "error":
			RedPrint(content)
		case "warning":
			YellowPrint(content)
		default:
			fmt.Printf("%v", content)
		}
	}

	return nil
}

func (l *Logger) startSyslogServer(host string, port string) error {
	// Connect to the syslog server
	logURL := host + ":" + port
	syslogger, err := syslog.Dial("udp", logURL, syslog.LOG_INFO, "log")
	if err != nil {
		return err
	}
	l.sysLogger = syslogger
	return nil
}
