package analytics

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
)

type AnalyticsConfiguration struct {
	RequestAnalitics       chan bool
	ReadAnalyticsData      chan *ShiroxyAnalytics
	TriggerInterval        int
	stop                   chan bool
	changeTriggerInterval  chan time.Duration
	lock                   *sync.RWMutex
	latestShiroxyAnalytics *ShiroxyAnalytics
	collectingAnalitics    bool
}

type ShiroxyAnalytics struct {
	TotalDomain            int
	TotalCertSize          int
	TotalCerts             int
	TotalUnSecuredDomains  int
	TotalSecuredDomains    int
	TotalFailedSSLAttempts int
	TotalSuccessSSLAttemps int
	TotalUser              int
	Memory_ALLOC           int
	Memory_TOTAL_ALLOC     int
	Memory_SYS             int
	CPU_Usage              float64
	GC_Count               int
	Metadata               map[string]interface{}
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func StartAnalytics(triggerInterval time.Duration) (*AnalyticsConfiguration, error) {
	analyticsConfiguration := &AnalyticsConfiguration{
		RequestAnalitics:      make(chan bool, 1),
		ReadAnalyticsData:     make(chan *ShiroxyAnalytics),
		stop:                  make(chan bool, 1),
		changeTriggerInterval: make(chan time.Duration),
	}

	ticker := time.NewTicker(triggerInterval)

	go func() {
		analyticsConfiguration.lock.Lock()
		var collectingAnalitics bool = true
		analyticsConfiguration.lock.Unlock()
		for {
			select {
			case <-ticker.C:
				if !collectingAnalitics {
					analyticsConfiguration.collectingAnalitics = true
					shiroxyAnalytics := ShiroxyAnalytics{}

					var memStat runtime.MemStats
					runtime.ReadMemStats(&memStat)

					shiroxyAnalytics.GC_Count = int(memStat.NumGC)
					shiroxyAnalytics.Memory_SYS = int(bToMb(memStat.Sys))
					shiroxyAnalytics.Memory_ALLOC = int(bToMb(memStat.Alloc))
					shiroxyAnalytics.Memory_TOTAL_ALLOC = int(bToMb(memStat.TotalAlloc))

					p, err := process.NewProcess(int32(os.Getpid()))
					if err != nil {
						// return nil, err
					}

					percent, err := p.CPUPercent()
					if err != nil {
						// return nil, err
					}
					shiroxyAnalytics.CPU_Usage = percent
					analyticsConfiguration.latestShiroxyAnalytics = &shiroxyAnalytics
					analyticsConfiguration.ReadAnalyticsData <- &shiroxyAnalytics

					analyticsConfiguration.lock.Lock()
					analyticsConfiguration.collectingAnalitics = false
					analyticsConfiguration.lock.Unlock()
				}
			case <-analyticsConfiguration.RequestAnalitics:
				if !collectingAnalitics {
					analyticsConfiguration.lock.Lock()
					analyticsConfiguration.collectingAnalitics = true
					analyticsConfiguration.lock.Unlock()

					shiroxyAnalytics := ShiroxyAnalytics{}

					var memStat runtime.MemStats
					runtime.ReadMemStats(&memStat)

					shiroxyAnalytics.GC_Count = int(memStat.NumGC)
					shiroxyAnalytics.Memory_SYS = int(bToMb(memStat.Sys))
					shiroxyAnalytics.Memory_ALLOC = int(bToMb(memStat.Alloc))
					shiroxyAnalytics.Memory_TOTAL_ALLOC = int(bToMb(memStat.TotalAlloc))

					p, err := process.NewProcess(int32(os.Getpid()))
					if err != nil {
						// return nil, err
					}

					percent, err := p.CPUPercent()
					if err != nil {
						// return nil, err
					}
					shiroxyAnalytics.CPU_Usage = percent
					analyticsConfiguration.latestShiroxyAnalytics = &shiroxyAnalytics
					analyticsConfiguration.ReadAnalyticsData <- &shiroxyAnalytics

					analyticsConfiguration.lock.Lock()
					analyticsConfiguration.collectingAnalitics = true
					analyticsConfiguration.lock.Unlock()
				}

			case newDuration := <-analyticsConfiguration.changeTriggerInterval:
				ticker.Stop()
				ticker = time.NewTicker(newDuration)
			case <-analyticsConfiguration.stop:
				ticker.Stop()
			}
		}
	}()

	return analyticsConfiguration, nil
}

func (a *AnalyticsConfiguration) UpdateTriggerInterval(triggerInterval time.Duration) error {
	a.changeTriggerInterval <- triggerInterval
	return nil
}

func (a *AnalyticsConfiguration) RequestAnalytics(forced bool) (*ShiroxyAnalytics, error) {
	// a.lock.RLock()
	// collectionAnalytics := a.collectingAnalitics
	// a.lock.RUnlock()

	// if collectionAnalytics {
	// 	shiroxyAnalytics := <-a.ReadAnalyticsData
	// 	return shiroxyAnalytics, nil
	// } else {
	// 	return a.latestShiroxyAnalytics, nil
	// }

	return a.latestShiroxyAnalytics, nil
}

// function to capture CPU profile
func captureCPUProfile(duration time.Duration) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := pprof.StartCPUProfile(&buf); err != nil {
		return nil, err
	}
	time.Sleep(duration) // profile for a duration
	pprof.StopCPUProfile()
	return &buf, nil
}

// function to capture and print memory profile
func captureAndPrintMemProfile() error {
	var buf bytes.Buffer
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(&buf); err != nil {
		return err
	}
	fmt.Println(buf.String()) // print the memory profile data
	return nil
}
