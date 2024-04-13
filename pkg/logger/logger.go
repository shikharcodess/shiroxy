package logger

import (
	"log"

	"gopkg.in/mcuadros/go-syslog.v2"
)

func StartLogger(enable bool, remoteEnable bool) (*Logger, error) {
	logger := Logger{}
	if remoteEnable {
		logger.syslog_channel = make(syslog.LogPartsChannel)

		logger.startSyslogServer()
	}
	return nil, nil
}

type Logger struct {
	syslog_channel syslog.LogPartsChannel
}

func (l *Logger) LogError(content any) {

}

func (l *Logger) LogSuccess(content any) {

}

func (l *Logger) LogWarning(content any) {

}

func (l *Logger) Log(content any) {

}

func (l *Logger) log(content any) {

}

func (l *Logger) startSyslogServer() {
	handler := syslog.NewChannelHandler(l.syslog_channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	server.SetHandler(handler)

	// Start syslog server
	server.ListenUDP("0.0.0.0:514")
	server.ListenTCP("0.0.0.0:514")

	// Start syslog server in a separate goroutine
	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			log.Println("Received log:", logParts["content"].(string))
			// Here you can process or forward the log message as needed
		}
	}(l.syslog_channel)
}
