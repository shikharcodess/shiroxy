# Notes

```golang
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
```
