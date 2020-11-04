package conf

import (
	"flag"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/baifei2014/go-mysql-ksync/infoc"
	"github.com/baifei2014/go-mysql-ksync/library/database/sql"
	"github.com/baifei2014/go-mysql-ksync/library/log"
	"github.com/siddontang/go-mysql/canal"
)

var (
	confPath string
	Conf     = &Config{}
)

type Config struct {
	Monitor *Monitor
	// master info
	MasterInfo *MasterInfoConfig
	// db
	DB *sql.Config

	CanalInfo *CanalConfig

	Log *log.Config
}

// Monitor wechat monitor
type Monitor struct {
	User   string
	Token  string
	Secret string
}

type CanalConfig struct {
	Instances []*InsConfig `toml:"instance"`
}

type InsConfig struct {
	*canal.Config
	MasterInfo *MasterInfoConfig `toml:"masterinfo"`
	Sources    []SourceConfig    `toml:"source"`
}

type MasterInfoConfig struct {
	Addr     string        `toml:"addr"`
	DBName   string        `toml:"dbName"`
	User     string        `toml:"user"`
	Password string        `toml:"password"`
	Charset  string        `toml:"charset"`
	Timeout  time.Duration `toml:"timeout"`
}

type SourceConfig struct {
	Schema string   `toml:"schema"`
	Tables []string `toml:"tables"`
}

type Addition struct {
	PrimaryKey []string `toml:"primarykey"` // kafka msg key
	OmitField  []string `toml:"omitfield"`  // field will be ignored in table
}

type CTable struct {
	PrimaryKey []string `toml:"primarykey"` // kafka msg key
	OmitField  []string `toml:"omitfield"`  // field will be ignored in table
	OmitAction []string `toml:"omitaction"` // action will be ignored in table
	Name       string   `toml:"name"`       // table name support regular expression
	Tables     []string
}

type Database struct {
	Schema   string        `toml:"schema"`
	Infoc    *infoc.Config `toml:"infoc"`
	CTables  []*CTable     `toml:"table"`
	TableMap map[string]*Addition
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")

}

func Init() error {
	return local()
}

func local() (err error) {
	_, err = toml.DecodeFile(confPath, &Conf)
	return
}
