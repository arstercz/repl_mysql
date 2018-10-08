# repl_mysql

`repl_mysql` is based on [replication](https://github.com/siddontang/go-mysql/replication) and can be used to replicate the MySQL table. Only support the following [event type](https://dev.mysql.com/doc/internals/en/binlog-event.html):
```
QUERY_EVENT
TABLE_MAP_EVENT
WRITE_ROWS_EVENTv0
UPDATE_ROWS_EVENTv0
DELETE_ROWS_EVENTv0
WRITE_ROWS_EVENTv1
UPDATE_ROWS_EVENTv1
DELETE_ROWS_EVENTv1
WRITE_ROWS_EVENTv2
UPDATE_ROWS_EVENTv2
DELETE_ROWS_EVENTv2
```

read more from [repl_mysql blog](https://arstercz.com/%E4%BD%BF%E7%94%A8-repl_mysql-%E7%9B%91%E6%8E%A7%E8%A1%A8%E6%95%B0%E6%8D%AE%E5%8F%98%E6%9B%B4/).

## How to use?

#### repl_mysql options

You must specify the value to the `-user` and `-pass` options, and the user must have `replication slave` privileges. the `-binlog` and `-pos` can be empty, which means use the current master status as the replication postition.

```
./repl_mysql -h
Usage of ./repl_mysql:
  -conf string
        configure file.
  -section string
        configure section. (default "replication") 
  -database string
        only replicate the database.
  -table string
        only replicate the table.
  -host string
        the mysql master server address. (default "localhost")
  -port int
        the mysql master server port. (default 3306)
  -user string
        replicate user (default "user_repl")
  -pass string
        replicate user password 
  -binlog string
        replicate from this binlog file
  -pos int
        replicate from this position which in the binlog file
  -rowevent
        whether print row event change
  -serverid int
        unique server id in the replication (default 99999)
```

##### use with configure file

```
[replication]
user = user_repl
pass = xxxxxxxx
host = 10.0.21.5
port = 3301
database = percona
table    = tri1
```
eg:
```
./repl_mysql -conf conn.conf -rowevent
......

Time: 2018-10-07T17:02:15
Type: TableMapEvent
Host: 10.3.254.106, Port: 3301
Schema: percona
Table: tri1
Binlog: mysql-bin.000035, Logpos: 3177, Eventsize: 53
== UpdateRowsEventV2 ==
  TableID: 143
  Flags: 1
  Column count: 2
  Values:
   +--
    0:10
    1:"arsterczxx"
   +--
    0:10
    1:"arsterczxxxxx"

```

#### only print the specified database message

```
# ./repl_mysql -host 10.0.21.5 -user user_repl -pass xxxxxx -port 3301 -database percona -rowevent -binlog mysql-bin.000035 -pos 3687
......
......

Time: 2018-10-08T10:22:53
Type: TableMapEvent
Host: 10.3.254.106, Port: 3301
Schema: percona
Table: test1
Binlog: mysql-bin.000035, Logpos: 3744, Eventsize: 57
== UpdateRowsEventV2 ==
  TableID: 197
  Flags: 1
  Column count: 4
  Values:
   +--
    0:5
    1:"flz1"
    2:201
    3:"2018-09-29 18:16:27"
   +--
    0:5
    1:"flz1"
    2:10301
    3:"2018-09-29 18:16:27"

Time: 2018-10-08T10:23:22
Type: QueryEvent
Host: 10.3.254.106, Port: 3301
Schema: percona
Binlog: mysql-bin.000035, Logpos: 3982, Eventsize: 85
Query: BEGIN


Time: 2018-10-08T10:23:22
Type: QueryEvent
Host: 10.3.254.106, Port: 3301
Schema: percona
Binlog: mysql-bin.000035, Logpos: 4109, Eventsize: 127
Query: update tri1 set name = "arstercz" where id = 10
```

#### only print the specified database and table message

```
./repl_mysql -host 10.0.21.5 -user user_repl -pass xxxxxx -port 3301 -database percona -table tri1 -rowevent -binlog mysql-bin.000035 -pos 3687
......
......

Time: 2018-10-08T10:23:22
Type: QueryEvent
Host: 10.3.254.106, Port: 3301
Schema: percona
Table: tri1
Binlog: mysql-bin.000035, Logpos: 4109, Eventsize: 127
Query: update tri1 set name = "arstercz" where id = 10
```

#### only print the specified table message

```
./repl_mysql -host 10.0.21.5 -user user_repl -pass xxxxxx -port 3301 -table tri1 -rowevent -binlog mysql-bin.000035 -pos 3687                  
......
......

Time: 2018-10-08T10:23:22
Type: QueryEvent
Host: 10.3.254.106, Port: 3301
Schema: percona
Table: tri1
Binlog: mysql-bin.000035, Logpos: 4109, Eventsize: 127
Query: update tri1 set name = "arstercz" where id = 10

Time: 2018-10-08T10:35:43
Type: QueryEvent
Host: 10.3.254.106, Port: 3301
Schema: percona2
Table: tri1
Binlog: mysql-bin.000035, Logpos: 4865, Eventsize: 129
Query: update tri1 set name = "arstercz" where id = 10
```
