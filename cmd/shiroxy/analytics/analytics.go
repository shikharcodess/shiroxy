package analytics

import (
	"os"
	"runtime"
	"shiroxy/pkg/logger"
	"sync"
	"time"

	"github.com/shirou/gopsutil/process"
)

type AnalyticsConfiguration struct {
	RequestAnalytics       chan bool
	ReadAnalyticsData      chan *ShiroxyAnalytics
	TriggerInterval        int
	stop                   chan bool
	changeTriggerInterval  chan time.Duration
	lock                   *sync.RWMutex
	latestShiroxyAnalytics *ShiroxyAnalytics
	collectingAnalytics    bool
}

type ShiroxyAnalytics struct {
	TotalDomain             int                    `json:"total_domain"`
	TotalCertSize           int                    `json:"total_cert_size"`
	TotalCerts              int                    `json:"total_certs"`
	TotalUnSecuredDomains   int                    `json:"total_unsecured_domains"`
	TotalSecuredDomains     int                    `json:"total_secured_domain"`
	TotalFailedSSLAttempts  int                    `json:"total_failed_ssl_attempts"`
	TotalSuccessSSLAttempts int                    `json:"total_success_ssl_attempts"`
	TotalUser               int                    `json:"total_user"`
	Memory_ALLOC            int                    `json:"memory_alloc"`
	Memory_TOTAL_ALLOC      int                    `json:"memory_total_alloc"`
	Memory_SYS              int                    `json:"memory_sys"`
	CPU_Usage               float64                `json:"cpu_usage"`
	GC_Count                int                    `json:"gc_count"`
	Metadata                map[string]interface{} `json:"metadata"`
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func StartAnalytics(triggerInterval time.Duration, logHandler *logger.Logger, wg *sync.WaitGroup) (*AnalyticsConfiguration, error) {
	analyticsConfiguration := &AnalyticsConfiguration{
		RequestAnalytics:      make(chan bool, 1),
		ReadAnalyticsData:     make(chan *ShiroxyAnalytics, 1),
		stop:                  make(chan bool, 1),
		changeTriggerInterval: make(chan time.Duration),
		lock:                  &sync.RWMutex{},
	}

	ticker := time.NewTicker(triggerInterval)

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ticker.C:
				analyticsConfiguration.lock.RLock()
				if !analyticsConfiguration.collectingAnalytics {
					analyticsConfiguration.lock.RUnlock()
					analyticsConfiguration.collectAnalytics(logHandler)
				} else {
					analyticsConfiguration.lock.RUnlock()
				}
			case <-analyticsConfiguration.RequestAnalytics:
				analyticsConfiguration.lock.RLock()
				if !analyticsConfiguration.collectingAnalytics {
					analyticsConfiguration.lock.RUnlock()
					analyticsConfiguration.collectAnalytics(logHandler)
				} else {
					analyticsConfiguration.lock.RUnlock()
				}
			case newDuration := <-analyticsConfiguration.changeTriggerInterval:
				ticker.Stop()
				ticker = time.NewTicker(newDuration)
			case <-analyticsConfiguration.stop:
				ticker.Stop()
				return
			}
		}
	}()

	return analyticsConfiguration, nil
}

func (a *AnalyticsConfiguration) collectAnalytics(logHandler *logger.Logger) {
	a.lock.Lock()
	a.collectingAnalytics = true
	a.lock.Unlock()

	shiroxyAnalytics := ShiroxyAnalytics{}

	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)

	shiroxyAnalytics.GC_Count = int(memStat.NumGC)
	shiroxyAnalytics.Memory_SYS = int(bToMb(memStat.Sys))
	shiroxyAnalytics.Memory_ALLOC = int(bToMb(memStat.Alloc))
	shiroxyAnalytics.Memory_TOTAL_ALLOC = int(bToMb(memStat.TotalAlloc))

	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		logHandler.LogError(err.Error(), "Analytics", "")
	}

	percent, err := p.CPUPercent()
	if err != nil {
		logHandler.LogError(err.Error(), "Analytics", "")
	}
	shiroxyAnalytics.CPU_Usage = percent

	a.lock.Lock()
	a.latestShiroxyAnalytics = &shiroxyAnalytics
	a.collectingAnalytics = false
	a.lock.Unlock()

	select {
	case a.ReadAnalyticsData <- &shiroxyAnalytics:
	default:
	}
}

func (a *AnalyticsConfiguration) UpdateTriggerInterval(triggerInterval time.Duration) error {
	a.changeTriggerInterval <- triggerInterval
	return nil
}

func (a *AnalyticsConfiguration) ReadAnalytics(forced bool) (*ShiroxyAnalytics, error) {
	if forced {
		a.RequestAnalytics <- true
	}

	a.lock.RLock()
	defer a.lock.RUnlock()

	return a.latestShiroxyAnalytics, nil
}

func (a *AnalyticsConfiguration) StopAnalytics() {
	a.stop <- true
}
