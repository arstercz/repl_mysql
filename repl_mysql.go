package main

import (
	"github.com/arstercz/go-mysql/replication"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/client"
	"golang.org/x/net/context"
	"github.com/arstercz/goconfig"
	_ "github.com/davecgh/go-spew/spew"
	"os"
	"regexp"
	"time"
	"flag"
	"fmt"
)

type SQLInfo struct {
	Type           replication.EventType
	Host           string
	Port           int
	Schema         string
	Table          string
	Timestamp      string
	Executiontime  uint32
	Binlogname     string
	Logpos         uint32
	Eventsize      uint32
	Query          string
}

var (
	section  = "replication"
	host     = "localhost"
	port     = int64(3306)
	username = "user_repl"
	password = ""
	database = ""
	table    = ""
	binlog   = ""
	serverid = int64(99999)
	pos      = int64(0)
	rowevent = false
)

func TimeFormat(t uint32) string {
	const time_format = "2006-01-02T15:04:05"
	return time.Unix(int64(t), 0).Format(time_format)
}

func LogOut(info SQLInfo) {
	fmt.Fprintf(os.Stdout, "\n\n")
	fmt.Fprintf(os.Stdout, "Time: %s\n", info.Timestamp)
	fmt.Fprintf(os.Stdout, "Type: %s\n", info.Type)
	fmt.Fprintf(os.Stdout, "Host: %s, Port: %d\n", info.Host, info.Port)
	fmt.Fprintf(os.Stdout, "Schema: %s\n", info.Schema)
	if info.Table != "" {
		fmt.Fprintf(os.Stdout, "Table: %s\n", info.Table)
	}
	fmt.Fprintf(os.Stdout, "Binlog: %s, Logpos: %d, Eventsize: %d\n", info.Binlogname, info.Logpos, info.Eventsize)
	if info.Query != "" {
		fmt.Fprintf(os.Stdout, "Query: %s\n", info.Query)
	}
}

func main() {
	conf := flag.String("conf", "", "configure file.")
	s := flag.String("section", "replication", "configure section.")
	h := flag.String("host", "localhost", "the mysql master server address.")
	P := flag.Int64("port", 3306, "the mysql master server port.")
	u := flag.String("user", "user_repl", "replicate user")
	p := flag.String("pass", "", "replicate user password")
	i := flag.Int64("serverid", 99999, "unique server id in the replication")
	f := flag.String("binlog", "", "replicate from this binlog file")
	n := flag.Int64("pos", 0, "replicate from this position which in the binlog file")
	d := flag.String("database", "", "only replicate the database.")
	t := flag.String("table", "", "only replicate the table.")
	r := flag.Bool("rowevent", false, "whether print row event change")

	flag.Parse()
	rowevent = *r

	if len(*conf) <= 0 {
		host = *h
		port = *P
		username = *u
		password = *p
		serverid = *i
		binlog = *f
		pos = *n
		database = *d
		table = *t
	} else {

		section = *s
		c, err := goconfig.ReadConfigFile(*conf)
		host, err = c.GetString(section, "host")
		port, err = c.GetInt64(section, "port")
		username, err = c.GetString(section, "user")
		password, err = c.GetString(section, "pass")
		binlog, err = c.GetString(section, "binlog")
		pos, err = c.GetInt64(section, "pos")
		serverid, err = c.GetInt64(section, "serverid")
		database, err = c.GetString(section, "database")
		if err != nil {
			fmt.Fprintf(os.Stderr, "readconfigfile err: " + err.Error())
			os.Exit(1)
		}
	}
	if serverid == 0 {
		serverid = *i
	}
	if (password == "") {
		fmt.Fprintf(os.Stderr, "[ERROR] must set password!\n\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}
	TypeCheck := map[replication.EventType]bool {
		replication.QUERY_EVENT: true,
		replication.TABLE_MAP_EVENT: true,
		replication.WRITE_ROWS_EVENTv0: true,
		replication.UPDATE_ROWS_EVENTv0: true,
		replication.DELETE_ROWS_EVENTv0: true,
		replication.WRITE_ROWS_EVENTv1: true,
		replication.UPDATE_ROWS_EVENTv1: true,
		replication.DELETE_ROWS_EVENTv1: true,
		replication.WRITE_ROWS_EVENTv2: true,
		replication.UPDATE_ROWS_EVENTv2: true,
		replication.DELETE_ROWS_EVENTv2: true,
	}


	replcfg := replication.BinlogSyncerConfig{
		ServerID: uint32(serverid),
		Flavor: "mysql",
		Host: host,
		Port: uint16(port),
		User: username,
		Password: password,
	}

	if binlog == "" || pos == 0 {
		c, err := client.Connect(fmt.Sprintf("%s:%d", host, port), username, password, "")
		if err != nil {
			fmt.Printf("connect master error: %v\n", err)
			os.Exit(2)
		}
		rs, err := c.Execute("SHOW MASTER STATUS")
		if err != nil {
			fmt.Printf("get master status error: %v\n", err)
			os.Exit(3)
		}
		binlog, _ = rs.GetString(0, 0)
		pos, _ = rs.GetInt(0, 1)
	}


	syncer := replication.NewBinlogSyncer(replcfg)
	streamer, err := syncer.StartSync(mysql.Position{Name: binlog, Pos: uint32(pos)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] streamer error %s\n", err)
		os.Exit(1)
	}
	var tableid uint64 = 0
	var tabletmp string
	for {
		event, _ := streamer.GetEvent(context.Background())
		if ! TypeCheck[event.Header.EventType] {
			continue
		}
		meta, _ := event.Event.GetMeta()

		if database != "" && meta.Schema != "" && database != meta.Schema {
			continue
		}
		eventinfo := SQLInfo{}
		eventinfo.Type = replication.EventType(event.Header.EventType)
		eventinfo.Host = host
		eventinfo.Port = int(port)
		eventinfo.Timestamp = TimeFormat(event.Header.Timestamp)
		eventinfo.Binlogname = binlog
		eventinfo.Logpos = uint32(event.Header.LogPos)
		eventinfo.Eventsize = event.Header.EventSize

		switch eventinfo.Type {
			case replication.QUERY_EVENT:
				eventinfo.Schema = meta.Schema
				eventinfo.Table = meta.Table
				eventinfo.Query = meta.Query
				if len(table) > 0 {
					matchstring := fmt.Sprintf("(?i:(\\s+|.)(%s|`%s`)\\s+)", table, table)
					matched, err := regexp.MatchString(matchstring, meta.Query)
					if matched && err == nil {
						eventinfo.Table = table
						LogOut(eventinfo)
					}
				} else {
					LogOut(eventinfo)
				}

			case replication.ROTATE_EVENT:
				binlog = meta.Binlog

			case replication.TABLE_MAP_EVENT:
				eventinfo.Schema = meta.Schema
				eventinfo.Table = meta.Table
				tabletmp = meta.Table
				tableid = meta.TableID
				if len(table) > 0 {
					if table == meta.Table {
						LogOut(eventinfo)
					}
				} else {
					LogOut(eventinfo)
				}

			case replication.WRITE_ROWS_EVENTv0, replication.UPDATE_ROWS_EVENTv0, replication.DELETE_ROWS_EVENTv0,
				replication.WRITE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv1,
				replication.WRITE_ROWS_EVENTv2, replication.UPDATE_ROWS_EVENTv2, replication.DELETE_ROWS_EVENTv2:
				if tableid == meta.TableID {
					if len(table) > 0 {
						if table == tabletmp {
							fmt.Fprintf(os.Stdout, "== %s ==\n", replication.EventType(event.Header.EventType))
							if rowevent {
								event.Event.Dump(os.Stdout)
							}
						}
					} else {
						fmt.Fprintf(os.Stdout, "== %s ==\n", replication.EventType(event.Header.EventType))
							if rowevent {
								event.Event.Dump(os.Stdout)
							}
					}
				}
				tableid = 0
				tabletmp = ""

			default:
		}
	}

}
