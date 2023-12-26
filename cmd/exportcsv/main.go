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

	baseCsvDir := os.Getenv("CSV_DIRECTORY")
	if baseCsvDir == "" {
		log.Fatal("CSV_DIRECTORY environment variable is not set.")
	}

	// DBネームを仕様したい場合はコメントアウト（スプシにインポートする際に環境変数変更する必要あり）
	//newDbDirPath := filepath.Join(baseCsvDir, dbInfo.Name)

	// 出力ディレクトリの作成
	_, err = createUniqueDir(baseCsvDir)
	if err != nil {
		log.Fatalf("Could not create unique directory: %v", err)
	}

	for _, table := range dbInfo.Tables {
		// CSVファイルのパス
		csvPath := fmt.Sprintf("%s/%s.csv", baseCsvDir, table.Name)

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
			"COMMENT",
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
				isForeign,
				col.ForeignKeyTable,
				col.ForeignKeyColumn,
				col.Comment,
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

// createUniqueDir は指定されたベースディレクトリに対してユニークなディレクトリを作成します。
func createUniqueDir(baseDir string) (string, error) {
	dir := baseDir
	for i := 1; ; i++ {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// ディレクトリが存在しない場合、ディレクトリを作成
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return "", err
			}
			return dir, nil
		}
		// ディレクトリが存在する場合、新しい名前を生成
		dir = fmt.Sprintf("%s_%d", baseDir, i)
	}
}
