package service

import (
	"fmt"
	"time"

	"github.com/baifei2014/go-mysql-ksync/conf"
	"github.com/baifei2014/go-mysql-ksync/library/log"

	"github.com/pkg/errors"
	"github.com/siddontang/go-mysql/canal"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
)

var ErrRuleNotExist = errors.New("rule is not exist")

// Instance canal instance
type Instance struct {
	*canal.Canal
	config *conf.InsConfig
	// one instance can have lots of target different by schema and table
	master *masterInfo

	targets []*Target

	// binlog latest timestamp, be used check delay
	latestTimestamp int64
	err             error
	closed          bool
}

// NewInstance new canal instance
func NewInstance(c *conf.InsConfig) (ins *Instance, err error) {
	// new instance
	ins = &Instance{config: c}
	// check and modify config
	if c.MasterInfo == nil {
		err = errors.New("no masterinfo config")
		ins.err = err
		return
	}

	if ins.master, err = loadMasterInfo(ins.config.MasterInfo); err != nil {
		ins.err = errors.Wrap(err, "init master info")
		return
	}
	ins.latestTimestamp = time.Now().Unix()
	// new canal
	if err = ins.newCanal(); err != nil {
		ins.err = errors.Wrap(err, "db init")
		return
	}
	// implement self as canal's event handler
	return
}

func (ins *Instance) newCanal() error {
	cfg := canal.NewDefaultConfig()
	cfg.Addr = ins.config.Addr
	cfg.User = ins.config.User
	cfg.Password = ins.config.Password
	cfg.Charset = ins.config.Charset
	cfg.ServerID = ins.config.ServerID

	cfg.Dump.DiscardErr = false

	for _, s := range ins.config.Sources {
		for _, t := range s.Tables {
			cfg.IncludeTableRegex = append(cfg.IncludeTableRegex, s.Schema+"\\."+t)
		}
	}

	var err error
	ins.Canal, err = canal.NewCanal(cfg)
	return errors.Wrap(err, "new canal")
}

func (ins *Instance) Check(ev *canal.RowsEvent) (ts []*Target) {
	for _, t := range ins.targets {
		if t.compare(ev.Table.Schema, ev.Table.Name, ev.Action) {
			ts = append(ts, t)
		}
	}
	return
}

func (ins *Instance) Start() {
	pos := ins.master.Pos()
	if pos.Name == "" || pos.Pos == 0 {
		var err error
		if pos, err = ins.Canal.GetMasterPos(); err != nil {
			log.Error("c.MasterPos error(%v)", err)
			ins.err = errors.Wrap(err, "canal get master pos when start")
			return
		}
	}
	ins.err = ins.Canal.RunFrom(pos)
}

func (ins *Instance) String() string {
	return ins.config.Addr
}

// OnRotate OnRotate
func (ins *Instance) OnRotate(re *replication.RotateEvent) error {
	log.Info("OnRotate binlog addr(%s) rotate binname(%s) pos(%d)", ins.config.Addr, re.NextLogName, re.Position)
	return nil
}

// OnDDL OnDDL
func (ins *Instance) OnDDL(pos mysql.Position, qe *replication.QueryEvent) error {
	log.Info("OnDDL binlog addr(%s) ddl binname(%s) pos(%d)", ins.config.Addr, pos.Name, pos.Pos)
	return nil
}

// OnXID OnXID
func (ins *Instance) OnXID(mysql.Position) error {
	return nil
}

//OnGTID OnGTID
func (ins *Instance) OnGTID(mysql.GTIDSet) error {
	return nil
}

// OnPosSynced OnPosSynced
func (ins *Instance) OnPosSynced(pos mysql.Position, force bool) error {
	return ins.master.Save(pos, force)
}

// OnRow send the envent to table
func (ins *Instance) OnRow(ev *canal.RowsEvent) error {
	for _, t := range ins.Check(ev) {
		t.send(ev)
	}
	// log.Infof("syncer_counter", ins.String(), ev.Table.Schema, tblReplacer.ReplaceAllString(ev.Table.Name, ""), ev.Action)
	// log.State("delay_syncer", ins.delay(), ins.String(), ev.Table.Schema, "", "")
	ins.latestTimestamp = time.Now().Unix()
	return nil
}

// Error returns instance error.
func (ins *Instance) Error() string {
	if ins.err == nil {
		return ""
	}
	return fmt.Sprintf("+%v", ins.err)
}

func (ins *Instance) delay() int64 {
	return time.Now().Unix() - ins.latestTimestamp
}
