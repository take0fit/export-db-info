package main

import (
	"log"
	"os"
	"project-root/pkg/sheets"
)

func main() {
	// Google Sheets APIの認証情報
	serviceAccountKeyFile := os.Getenv("GOOGLE_SERVICE_ACCOUNT_KEY_FILE")

	// CSVファイルが保存されているディレクトリのパス
	csvDir := os.Getenv("CSV_DIRECTORY")

	// Google Sheets APIクライアントの初期化
	client, err := sheets.InitializeClient(serviceAccountKeyFile)
	if err != nil {
		log.Fatalf("Error initializing Google Sheets client: %v", err)
	}

	// スプレッドシートへのインポート
	if err := sheets.ImportCSVToSheets(client, csvDir); err != nil {
		log.Fatalf("Error importing CSV to Sheets: %v", err)
	}
}
