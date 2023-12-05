package sheets

// InitializeClient はGoogle Sheets APIクライアントを初期化します。
func InitializeClient(serviceAccountKeyFile string) (*sheets.Service, error) {
	// クライアントの初期化ロジック
	// ...
}

// ImportCSVToSheets はCSVファイルをGoogleスプレッドシートにインポートします。
func ImportCSVToSheets(client *sheets.Service, csvDir string) error {
	// CSVインポートのロジック
	// ...
}
