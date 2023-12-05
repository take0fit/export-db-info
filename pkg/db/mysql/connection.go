package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

// Connect はMySQLデータベースへの接続を確立します。
func Connect(username, password, hostname, port, dbname string) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname)
	fmt.Println(dsn)
	return sql.Open("mysql", dsn)
}
