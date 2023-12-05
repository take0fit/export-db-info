package main

import (
	"encoding/csv"
	"export-db-info/internal/db/mysql_internal"
	"fmt"
	"log"
	"os"
)

func main() {
	// .env ファイルから環境変数をロード
	//err := godotenv.Load("../../.env")
	//if err != nil {
	//	log.Fatal("Error loading .env file")
	//}

	dbInfo, err := mysql_internal.GetDatabaseInfo()
	if err != nil {
		log.Fatalf("faild get db info: %v", err)
	}

	// 出力ディレクトリの作成
	if err := os.Mkdir(dbInfo.Name, 0755); err != nil {
		log.Fatalf("Could not create directory: %v", err)
	}

	for _, table := range dbInfo.Tables {
		// CSVファイルのパス
		csvPath := fmt.Sprintf("%s/%s.csv", dbInfo.Name, table.Name)

		// CSVファイルの作成
		csvFile, err := os.Create(csvPath)
		if err != nil {
			log.Fatalf("Could not create CSV file for table %s: %v", table.Name, err)
		}

		// CSVライターの作成
		writer := csv.NewWriter(csvFile)

		// ヘッダー行の書き込み
		headers := []string{
			"COLUMN_NAME",
			"COLUMN_TYPE",
			"IS_PRIMARY_KEY",
			"IS_NULLABLE",
			"IS_UNIQUE",
			"IS_INDEX",
			"IS_FOREiGN_KEY",
			"FOREiGN_KEY_TABLE",
			"FOREiGN_KEY_COLUMN",
		}
		if err := writer.Write(headers); err != nil {
			log.Fatalf("Could not write headers to CSV for table %s: %v", table.Name, err)
		}

		for _, col := range table.Columns {
			isPrimaryKey := "×"
			isNullable := "×"
			isUnique := "×"
			isIndexed := "×"
			isForeign := "×"

			if col.IsNullable {
				isNullable = "○"
			}
			if col.IsPrimaryKey {
				isPrimaryKey = "○"
			}
			if col.IsUnique {
				isUnique = "○"
			}
			if col.IsIndexed {
				isIndexed = "○"
			}
			if col.IsForeign {
				isIndexed = "○"
			}

			record := []string{
				col.Name,
				col.Type,
				isPrimaryKey,
				isNullable,
				isUnique,
				isIndexed,
				isUnique,
				isForeign,
				col.ForeignKeyTable,
				col.ForeignKeyColumn,
			}

			if err := writer.Write(record); err != nil {
				log.Fatalf("Could not write row to CSV for column %s: %v", col.Name, err)
			}
		}
		// CSVライターをフラッシュ
		writer.Flush()

		// CSVファイルをクローズ
		csvFile.Close()
	}
}
