/*
currently sqlite3 and mysql are supported
*/
package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

const (
	DRIVER_SQLITE3 = "sqlite3"
	DRIVER_MYSQL   = "mysql"
)

type SqlDb struct {
	driver string
	dsn    string
	debug  bool
	logger *log.Logger
	db     *sql.DB
}

func NewSqlDb(driver, dsn string, logger *log.Logger) *SqlDb {
	this := new(SqlDb)
	this.driver = driver
	this.dsn = dsn
	this.logger = logger
	this.debug = false

	// conn to db
	var err error
	this.db, err = sql.Open(this.driver, this.dsn)
	this.checkError(err, dsn)

	return this
}

func (this SqlDb) String() string {
	return fmt.Sprintf("%s[%s]", this.driver, this.dsn)
}

func (this *SqlDb) SetDebug(d bool) {
	this.debug = d
}

// sets the maximum number of connections in the idle connection pool
func (this *SqlDb) SetMaxIdleConns(n int) {
	this.db.SetMaxIdleConns(n)
}

func (this *SqlDb) checkError(err error, sql string) {
	if err != nil {
		panic(fmt.Sprintf("%s: %s, %s", this, err.Error(), sql))
	}
}

func (this *SqlDb) CreateDb(createTableSql string) {
	if this.debug {
		this.logger.Println(createTableSql)
	}

	var err error
	_, err = this.db.Exec(createTableSql)
	this.checkError(err, createTableSql)

	if this.driver == DRIVER_SQLITE3 {
		// performance tuning for sqlite3
		// http://www.sqlite.org/cvstrac/wiki?p=DatabaseIsLocked
		this.db.Exec("PRAGMA synchronous = OFF")
		this.db.Exec("PRAGMA journal_mode = MEMORY")
		this.db.Exec("PRAGMA read_uncommitted = true")
	}
}

func (this *SqlDb) Query(query string, args ...interface{}) *sql.Rows {
	if this.debug {
		this.logger.Printf("%s, args=%+v\n", query, args)
	}

	rows, err := this.db.Query(query, args...)
	this.checkError(err, query)

	return rows
}

func (this *SqlDb) QueryRow(query string, args ...interface{}) *sql.Row {
	if this.debug {
		this.logger.Printf("%s, args=%+v\n", query, args)
	}

	return this.db.QueryRow(query, args...)
}

func (this *SqlDb) ExecSql(query string, args ...interface{}) (afftectedRows int64) {
	if this.debug {
		this.logger.Printf("%s, args=%+v\n", query, args)
	}

	res, err := this.db.Exec(query, args...)
	this.checkError(err, query)

	afftectedRows, err = res.RowsAffected()
	this.checkError(err, query)

	return
}

func (this *SqlDb) Prepare(query string) *sql.Stmt {
	if this.debug {
		this.logger.Println(query)
	}

	r, err := this.db.Prepare(query)
	this.checkError(err, query)
	return r
}

func (this *SqlDb) Close() error {
	return this.db.Close()
}

func (this *SqlDb) Db() *sql.DB {
	return this.db
}
