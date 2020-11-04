package service

import (
	"regexp"
	"sync"
	"time"

	"github.com/baifei2014/go-mysql-ksync/conf"
	"github.com/baifei2014/go-mysql-ksync/dao"
	"github.com/baifei2014/go-mysql-ksync/library/stat/prom"
	"github.com/siddontang/go/log"
)

var (
	stats       = prom.New().WithState("go_canal_counter", []string{"type", "addr", "scheme", "table", "action"})
	tblReplacer = regexp.MustCompile("[0-9]+") // NOTE: replace number of sub-table name to space
)

type Canal struct {
	dao       *dao.Dao
	instances map[string]*Instance
	insl      sync.RWMutex
	tidbInsl  sync.RWMutex
	err       error
	errCount  int64
	lastErr   time.Time
}

func NewCanal(config *conf.Config) (c *Canal) {
	c = &Canal{}
	c.instances = make(map[string]*Instance, len(config.CanalInfo.Instances))
	for _, insc := range config.CanalInfo.Instances {
		ins, err := NewInstance(insc)
		if err != nil {
			log.Errorf("new instance error(%+v)", err)
		}
		c.insl.Lock()
		c.instances[insc.Addr] = ins
		c.insl.Unlock()
		if err == nil {
			go ins.Start()
			log.Infof("start canal instance(%s) success", ins.String())
		}
	}
	return
}

func (c *Canal) Close() {
	c.insl.RLock()
	defer c.insl.RUnlock()
	for _, ins := range c.instances {
		ins.Close()
		log.Infof("close canal instance(%s) success", ins.String())
	}
}
