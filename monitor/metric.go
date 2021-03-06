package monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"runtime"

	"os"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxWork struct {
	Client   client.Client
	Address  string
	Username string
	Password string

	DatabaseName string
}

func NewInfluxWork(addr, username, password, databasename string) *InfluxWork {
	temp := &InfluxWork{}
	config := client.HTTPConfig{
		Addr:      addr,
		Username:  username,
		Password:  password,
		UserAgent: "Axon",
	}
	temp.DatabaseName = databasename

	var err error
	temp.Client, err = client.NewHTTPClient(config)
	if err != nil {
		return nil
	}

	return temp
}

func (in *InfluxWork) InitDatabase() error {
	createDbCmd := fmt.Sprintf("create database %s", in.DatabaseName)
	query := client.NewQuery(createDbCmd, "", "")
	if _, err := in.Client.Query(query); err != nil {
		return err
	}
	return nil
}

type Monitor struct {
	Os       string
	Arch     string
	Hostname string
	Program  string

	Goruntine int64
	CgoCall   int64
	MemStat   *runtime.MemStats

	SrvConnTotal  int64
	SrvConnPerMin int64

	Cpu int64

	Dur time.Duration
	sync.RWMutex

	worker *InfluxWork
}

func NewMonitor(addr, username, password, databasename string) *Monitor {
	temp := new(Monitor)
	temp.worker = NewInfluxWork(addr, username, password, databasename)
	temp.Dur = time.Second * 10
	return temp
}

type Collector interface {
	Reportor()
	Collection()
}

func (m *Monitor) Dog() {

	m.Arch = runtime.GOARCH
	if m.Arch == "" {
		m.Arch = "unknown"
	}
	m.Os = runtime.GOOS
	if m.Os == "" {
		m.Os = "unknown"
	}

	m.Hostname, _ = os.Hostname()
	m.Program = strings.TrimLeft(os.Args[0], "./")
	tick := time.NewTicker(m.Dur)
	defer func() {
		tick.Stop()
	}()
	for {
		select {
		case <-tick.C:
			m.Collection()
			m.Reportor()
		}
	}

}

func (m *Monitor) Collection() {
	m.RLock()
	defer m.RUnlock()
	m.Goruntine = int64(runtime.NumGoroutine())
	m.CgoCall = runtime.NumCgoCall()
	m.MemStat = new(runtime.MemStats)
	m.Cpu = int64(runtime.NumCPU())
	runtime.ReadMemStats(m.MemStat)
}

func (m *Monitor) ReportGCAndMem() {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  m.worker.DatabaseName,
		Precision: "s",
	})

	fields := map[string]interface{}{
		"cpu.number":       int(m.Cpu),
		"cgocall.number":   int(m.CgoCall),
		"goruntine.number": int(m.Goruntine),
		"mem.alloc":        int(m.MemStat.Alloc),
		"mem.buckhashsys":  int(m.MemStat.BuckHashSys),
		//"mem.bysize":        m.MemStat.BySize,
		"mem.frees":         int(m.MemStat.Frees),
		"mem.gc.sys":        int(m.MemStat.GCSys),
		"mem.gc.next":       int(m.MemStat.NextGC),
		"mem.gc.numforced":  int(m.MemStat.NumForcedGC),
		"mem.gc.last":       int(m.MemStat.LastGC),
		"mem.numgc":         int(m.MemStat.NumGC),
		"mem.heap.alloc":    int(m.MemStat.HeapAlloc),
		"mem.heap.idle":     int(m.MemStat.HeapIdle),
		"mem.heap.inuse":    int(m.MemStat.HeapInuse),
		"mem.heap.objects":  int(m.MemStat.HeapObjects),
		"mem.heap.released": int(m.MemStat.HeapReleased),
		"mem.heap.sys":      int(m.MemStat.HeapSys),

		"mem.lookups":      int(m.MemStat.Lookups),
		"mem.mallocs":      int(m.MemStat.Mallocs),
		"mem.mcache.inuse": int(m.MemStat.MCacheInuse),
		"mem.mcache.sys":   int(m.MemStat.MCacheSys),
		"mem.mspan.inuse":  int(m.MemStat.MSpanInuse),
		"mem.mspan.sys":    int(m.MemStat.MSpanSys),

		"mem.othersys":      int(m.MemStat.OtherSys),
		"mem.pause.totalns": int(m.MemStat.PauseTotalNs),
		"mem.stack.inuse":   int(m.MemStat.StackInuse),
		"mem.stack.sys":     int(m.MemStat.StackSys),
		"mem.sys":           int(m.MemStat.Sys),
		"mem.totalalloc":    int(m.MemStat.TotalAlloc),
	}
	tags := map[string]string{
		"os":      m.Os,
		"arch":    m.Arch,
		"host":    m.Hostname,
		"program": m.Program,
	}
	pt, err := client.NewPoint(
		"go",
		tags,
		fields,
		time.Now(),
	)
	fmt.Println(pt)
	bp.AddPoint(pt)

	err = m.worker.Client.Write(bp)
	if err != nil {
		fmt.Println("Error: ", err.Error())
	}
	return
}

func (m *Monitor) ReportProgram() {
	return
}

func (m *Monitor) Reportor() {
	go m.ReportGCAndMem()
	//go m.ReportProgram()
}
