package mysql_internal

import (
	"database/sql"
	"export-db-info/internal/model/sql_model"
	"export-db-info/pkg/db/mysql"
	"os"
)

var (
	dbHost     = os.Getenv("DB_HOST")
	dbPort     = os.Getenv("DB_PORT")
	dbName     = os.Getenv("DB_DATABASE")
	dbUser     = os.Getenv("DB_USERNAME")
	dbPassword = os.Getenv("DB_PASSWORD")
)

func GetDatabaseInfo() (*sql_model.DB, error) {
	// データベースに接続
	db, err := mysql.Connect(dbUser, dbPassword, dbHost, dbPort, dbName)
	if err != nil {
		return nil, err
	}

	// 各テーブルとそのカラム情報を取得
	tables, err := getTables(db)
	if err != nil {
		return nil, err
	}

	return &sql_model.DB{Name: dbName, Tables: tables}, nil
}

func getTables(db *sql.DB) ([]*sql_model.Table, error) {
	var tables []*sql_model.Table

	// テーブル一覧の取得
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, err
		}

		// カラム情報の取得
		columns, err := getColumns(db, tableName)
		if err != nil {
			return nil, err
		}

		tables = append(tables, &sql_model.Table{Name: tableName, Columns: columns})
	}

	return tables, nil
}

func getColumns(db *sql.DB, tableName string) ([]*sql_model.Column, error) {
	var columns []*sql_model.Column

	// カラム情報の取得
	query := `
    SELECT COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, COLUMN_COMMENT, 
           COLUMN_KEY
    FROM information_schema.COLUMNS 
    WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
    ORDER BY ORDINAL_POSITION
    `
	rows, err := db.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		col := new(sql_model.Column)
		var isNullable, columnKey string
		var defaultVal sql.NullString
		err := rows.Scan(&col.Name, &col.Type, &isNullable, &defaultVal, &col.Comment, &columnKey)
		if err != nil {
			return nil, err
		}

		// sql.NullStringの値をチェック
		if defaultVal.Valid {
			col.Default = defaultVal.String
		} else {
			col.Default = "NULL" // NULLの場合、デフォルト値を空文字列に設定
		}
		col.IsNullable = isNullable == "YES"
		col.IsPrimaryKey = columnKey == "PRI"
		// ユニーク制約の確認（データベース固有のクエリが必要）
		col.IsUnique, err = checkColumnIsUnique(db, tableName, col.Name)
		if err != nil {
			return nil, err
		}
		// インデックスの存在確認（データベース固有のクエリが必要）
		col.IsIndexed, err = checkColumnIsIndexed(db, tableName, col.Name)
		if err != nil {
			return nil, err
		}
		// 外部キーの存在確認（データベース固有のクエリが必要）
		col.IsForeign, err = checkColumnIsForeignKey(db, tableName, col.Name)
		if err != nil {
			return nil, err
		}

		if col.IsForeign {
			// 外部キー情報の取得
			fkTable, fkColumn, err := getForeignKeyInfo(db, tableName, col.Name)
			if err != nil {
				// エラー処理
				return nil, err
			}
			col.ForeignKeyTable = fkTable
			col.ForeignKeyColumn = fkColumn
		}

		columns = append(columns, col)
	}

	return columns, nil
}

func checkColumnIsUnique(db *sql.DB, tableName, columnName string) (bool, error) {
	var count int
	query := `
    SELECT COUNT(*)
    FROM information_schema.TABLE_CONSTRAINTS AS tc
    JOIN information_schema.KEY_COLUMN_USAGE AS kcu ON tc.CONSTRAINT_NAME = kcu.CONSTRAINT_NAME
    WHERE tc.TABLE_SCHEMA = DATABASE() AND tc.TABLE_NAME = ? AND kcu.COLUMN_NAME = ? AND tc.CONSTRAINT_TYPE = 'UNIQUE'
    `
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func checkColumnIsIndexed(db *sql.DB, tableName, columnName string) (bool, error) {
	var count int
	query := `
    SELECT COUNT(*)
    FROM information_schema.STATISTICS
    WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
    `
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func checkColumnIsForeignKey(db *sql.DB, tableName, columnName string) (bool, error) {
	var count int
	query := `
        SELECT COUNT(*)
        FROM information_schema.KEY_COLUMN_USAGE
        WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = ?
        AND COLUMN_NAME = ?
        AND REFERENCED_TABLE_NAME IS NOT NULL
    `
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func getForeignKeyInfo(db *sql.DB, tableName, columnName string) (string, string, error) {
	var fkTable, fkColumn string
	query := `
        SELECT REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME
        FROM information_schema.KEY_COLUMN_USAGE
        WHERE TABLE_SCHEMA = DATABASE()
        AND TABLE_NAME = ?
        AND COLUMN_NAME = ?
        AND REFERENCED_TABLE_NAME IS NOT NULL
    `
	err := db.QueryRow(query, tableName, columnName).Scan(&fkTable, &fkColumn)
	if err != nil {
		return "", "", err
	}
	return fkTable, fkColumn, nil
}
