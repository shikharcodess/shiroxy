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

CPU, Memory

```golang
func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

printMemUsage()
	// Allocate memory and print usage
	a := make([]byte, 10*1024*1024) // 10 MiB
	_ = a
	printMemUsage()

	// CPU usage
	percent, _ := cpu.Percent(time.Second, false)
	fmt.Printf("CPU Percent: %v%%\n", percent[0]) // On multi-core systems, handle appropriately

	// Memory usage
	v, _ := mem.VirtualMemory()
	fmt.Printf("Memory Usage: Used %v MB, Total: %v MB, Usage: %f%%\n", v.Used/1024/1024, v.Total/1024/1024, v.UsedPercent)

	// counters, _ := net.IOCounters(true) // per network interface
	// for _, counter := range counters {
	// 	fmt.Printf("Interface: %v\nBytes Sent: %v, Bytes Recv: %v\n", counter.Name, counter.BytesSent, counter.BytesRecv)
	// }
```
