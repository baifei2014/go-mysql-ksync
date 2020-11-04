package service

import (
	"sync"
	"time"

	"github.com/baifei2014/go-mysql-ksync/conf"
	"github.com/baifei2014/go-mysql-ksync/library/log"
	"github.com/juju/errors"
	"github.com/siddontang/go-mysql/client"
	"github.com/siddontang/go-mysql/mysql"
)

type masterInfo struct {
	c *conf.MasterInfoConfig
	l sync.RWMutex

	addr string

	binName      string `toml:"bin_name"`
	binPos       uint32 `toml:"bin_pos"`
	lastSaveTime time.Time
}

func loadMasterInfo(c *conf.MasterInfoConfig) (*masterInfo, error) {
	m := &masterInfo{c: c}
	conn, err := client.Connect(c.Addr, c.User, c.Password, c.DBName)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer conn.Close()
	if m.c.Timeout > 0 {
		conn.SetDeadline(time.Now().Add(m.c.Timeout * time.Second))
	}
	r, err := conn.Execute("SELECT addr,bin_name,bin_pos FROM master_info WHERE addr=?", c.Addr)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if r.RowNumber() == 0 {
		if m.c.Timeout > 0 {
			conn.SetDeadline(time.Now().Add(m.c.Timeout * time.Second))
		}
		if _, err = conn.Execute("INSERT INTO master_info (addr,bin_name,bin_pos) VALUE (?,'',0)", c.Addr); err != nil {
			return nil, errors.Trace(err)
		}
	} else {
		m.addr, _ = r.GetStringByName(0, "addr")
		m.binName, _ = r.GetStringByName(0, "bin_name")
		bpos, _ := r.GetIntByName(0, "bin_pos")
		m.binPos = uint32(bpos)
	}
	return m, nil
}

func (m *masterInfo) Save(pos mysql.Position, force bool) error {
	n := time.Now()
	if !force && n.Sub(m.lastSaveTime) < 2*time.Second {
		return nil
	}
	conn, err := client.Connect(m.c.Addr, m.c.User, m.c.Password, m.c.DBName)
	if err != nil {
		log.Error("db master info client error(%v)", err)
		return errors.Trace(err)
	}
	defer conn.Close()

	if m.c.Timeout > 0 {
		conn.SetDeadline(time.Now().Add(m.c.Timeout * time.Second))
	}
	if _, err = conn.Execute("UPDATE master_info SET bin_name=?,bin_pos=? WHERE addr=?", pos.Name, pos.Pos, m.addr); err != nil {
		log.Error("db save master info error(%v)", err)
		return errors.Trace(err)
	}
	m.lastSaveTime = n
	return nil
}

func (m *masterInfo) Pos() mysql.Position {
	var pos mysql.Position
	m.l.RLock()
	pos.Name = m.binName
	pos.Pos = m.binPos
	m.l.RUnlock()
	return pos
}
