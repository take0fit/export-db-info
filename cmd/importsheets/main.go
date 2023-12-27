package main

import (
	"context"
	"encoding/csv"
	"export-db-info/internal/google_internal"
	"export-db-info/internal/model/google_model"
	"export-db-info/pkg/lib/google"
	"fmt"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	ctx := context.Background()
	// Google Sheets APIの認証情報
	serviceAccountKeyFile := os.Getenv("GOOGLE_SERVICE_ACCOUNT_KEY_FILE")

	// Google Sheets APIクライアントの初期化
	sheSrv, err := google.InitializeSheetsClient(ctx, serviceAccountKeyFile)
	if err != nil {
		log.Fatalf("Error initializing Google Sheets client: %v", err)
	}

	// CSVファイルが保存されているディレクトリのパス
	csvDir := os.Getenv("CSV_DIRECTORY")
	files, err := ioutil.ReadDir(csvDir)
	if err != nil {
		log.Fatalf("Unable to read directory: %v", err)
	}

	lastPass := filepath.Base(csvDir)

	spreadsheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: lastPass,
		},
	}

	// スプレッドシートの作成
	createdSpreadsheet, err := sheSrv.Spreadsheets.Create(spreadsheet).Do()
	if err != nil {
		log.Fatalf("Unable to create spreadsheet. %v", err)
	}
	spreadsheetId := createdSpreadsheet.SpreadsheetId

	// インデックスページ（通常は最初のシート）のIDを取得
	var indexSheetId int64
	if len(createdSpreadsheet.Sheets) > 0 {
		indexSheetId = createdSpreadsheet.Sheets[0].Properties.SheetId
	}

	// 共有設定のリクエスト
	permission := &drive.Permission{
		Type:         "user",
		Role:         "writer",
		EmailAddress: os.Getenv("GOOGLE_SERVICE_ACCOUNT_EMAIL"), // 共有するユーザーのメールアドレス
	}

	// Google Drive APIクライアントの初期化
	drvSrv, err := google.InitializeDriveClient(ctx, serviceAccountKeyFile)

	_, err = drvSrv.Permissions.Create(spreadsheetId, permission).Do()
	if err != nil {
		log.Fatalf("Unable to create permission: %v", err)
	}

	log.Println("Spreadsheet shared successfully.")

	// シート名とIDのマッピングを格納する変数
	sheetMappings := make(map[string]int64)

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".csv" {
			// CSVファイルの読み込み
			f, err := os.Open(filepath.Join(csvDir, file.Name()))
			if err != nil {
				log.Printf("Unable to read csv file: %v", err)
				continue
			}
			defer f.Close()

			r := csv.NewReader(f)
			records, err := r.ReadAll()
			if err != nil {
				log.Printf("Unable to parse csv file: %v", err)
				continue
			}

			tableName := strings.TrimSuffix(file.Name(), ".csv")

			newSheet := &sheets.SheetProperties{
				Title: tableName,
			}
			addSheetRequest := &sheets.AddSheetRequest{
				Properties: newSheet,
			}
			batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
				Requests: []*sheets.Request{{
					AddSheet: addSheetRequest,
				}},
			}

			resp, err := sheSrv.Spreadsheets.BatchUpdate(spreadsheetId, batchUpdateRequest).Do()
			if err != nil {
				log.Printf("Unable to create new sheet: %v", err)
				continue
			}

			// 新しいシートのIDを取得
			var newSheetId int64
			if len(resp.Replies) > 0 && resp.Replies[0].AddSheet != nil {
				newSheetId = resp.Replies[0].AddSheet.Properties.SheetId
				sheetMappings[tableName] = newSheetId
			} else {
				log.Fatal("Failed to get the new sheet ID")
			}

			fmt.Println(newSheetId)

			var requests []*sheets.Request

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 0, EndRow: 2, StartCol: 0, EndCol: 3}, // セルの範囲オプションを設定
					true,     // セルの結合
					"CENTER", // 水平方向の配置
					"MIDDLE", // 垂直方向の配置
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25}, // 背景色
					&sheets.Color{Red: 1, Green: 1, Blue: 1},          // テキスト色
					"テーブル仕様書",                                         // セルに挿入するテキスト
					&sheets.TextFormat{FontSize: 10, Bold: true},      // テキストフォーマット
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 0, EndRow: 1, StartCol: 3, EndCol: 5},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"テーブル論理名",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 1, EndRow: 2, StartCol: 3, EndCol: 5},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"テーブル物理名",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 0, EndRow: 1, StartCol: 5, EndCol: 10},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					nil,
					"",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 1, EndRow: 2, StartCol: 5, EndCol: 10},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					nil,
					tableName,
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 0, EndRow: 1, StartCol: 10, EndCol: 11},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"作成者",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 0, EndRow: 1, StartCol: 11, EndCol: 12},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					nil,
					"",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 0, EndRow: 1, StartCol: 12, EndCol: 13},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"修正者",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 0, EndRow: 1, StartCol: 13, EndCol: 14},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					nil,
					"",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 1, EndRow: 2, StartCol: 10, EndCol: 11},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"作成日",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 1, EndRow: 2, StartCol: 11, EndCol: 12},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					nil,
					"",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 1, EndRow: 2, StartCol: 12, EndCol: 13},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"修正日",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 1, EndRow: 2, StartCol: 13, EndCol: 14},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					nil,
					"",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 2, EndRow: 4, StartCol: 0, EndCol: 2},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"内容説明",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 2, EndRow: 4, StartCol: 2, EndCol: 14},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					nil,
					"",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 0, EndCol: 1},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"No",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 1, EndCol: 4},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"カラム名",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 4, EndCol: 5},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"型",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 5, EndCol: 6},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"主キー",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 6, EndCol: 7},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"NULL",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 7, EndCol: 8},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"unique",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 8, EndCol: 9},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"index",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 9, EndCol: 10},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"外部キー",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 10, EndCol: 11},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"外部キーテーブル",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 11, EndCol: 12},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"外部キーカラム",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			requests = append(
				requests,
				google_internal.CreateSheetLayoutRequest(
					newSheetId,
					&google_model.RangeOption{StartRow: 6, EndRow: 7, StartCol: 12, EndCol: 14},
					true,
					"CENTER",
					"MIDDLE",
					&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
					&sheets.Color{Red: 1, Green: 1, Blue: 1},
					"コメント",
					&sheets.TextFormat{FontSize: 10, Bold: false},
				)...,
			)

			for ri, record := range records {

				if ri == 0 {
					continue
				}

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 0, EndCol: 1},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						strconv.Itoa(ri),
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 1, EndCol: 4},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[0],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 4, EndCol: 5},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[1],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 5, EndCol: 6},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[2],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 6, EndCol: 7},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[3],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 7, EndCol: 8},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[4],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 8, EndCol: 9},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[5],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 9, EndCol: 10},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[6],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 10, EndCol: 11},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[7],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 11, EndCol: 12},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[8],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)

				requests = append(
					requests,
					google_internal.CreateSheetLayoutRequest(
						newSheetId,
						&google_model.RangeOption{StartRow: int64(ri) + 6, EndRow: int64(ri) + 7, StartCol: 12, EndCol: 14},
						true,
						"CENTER",
						"MIDDLE",
						&sheets.Color{Red: 1, Green: 1, Blue: 1},
						&sheets.Color{Red: 0.25, Green: 0.25, Blue: 0.25},
						record[9],
						&sheets.TextFormat{FontSize: 10, Bold: false},
					)...,
				)
			}

			batchUpdateRequestForLayout := &sheets.BatchUpdateSpreadsheetRequest{
				Requests: requests,
			}

			_, err = sheSrv.Spreadsheets.BatchUpdate(spreadsheetId, batchUpdateRequestForLayout).Do()
			if err != nil {
				log.Printf("Unable to create new sheet: %v", err)
				continue
			}

			// 3秒間待機
			time.Sleep(3 * time.Second)
		}
	}

	// インデックスページにシート名とリンクを追加するリクエストを作成
	var indexRequests []*sheets.Request
	rowIndex := 0
	for sheetName, sheetId := range sheetMappings {
		indexRequests = append(indexRequests, createIndexEntryRequest(sheetName, sheetId, rowIndex, indexSheetId)...)
		rowIndex++
	}

	// インデックスページの更新を実行
	_, err = sheSrv.Spreadsheets.BatchUpdate(spreadsheetId, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: indexRequests,
	}).Do()
	if err != nil {
		log.Fatalf("Unable to update index page: %v", err)
	}
}

// createIndexPageRequest は、インデックスページを初期化するためのリクエストを作成します
func createIndexPageRequest() *sheets.Request {
	return &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: "インデックス",
			},
		},
	}
}

// createIndexEntryRequest は、インデックスページにシート名とリンクを追加するためのリクエストを作成します
func createIndexEntryRequest(sheetName string, sheetId int64, rowIndex int, indexSheetId int64) []*sheets.Request {
	// シート名をインデックスページに追加
	//appendSheetNameRequest := &sheets.Request{
	//	UpdateCells: &sheets.UpdateCellsRequest{
	//		Start: &sheets.GridCoordinate{
	//			SheetId:     indexSheetId, // インデックスページのID
	//			RowIndex:    int64(rowIndex),
	//			ColumnIndex: 0, // シート名の列
	//		},
	//		Rows: []*sheets.RowData{
	//			{
	//				Values: []*sheets.CellData{
	//					{
	//						UserEnteredValue: &sheets.ExtendedValue{
	//							StringValue: &sheetName,
	//						},
	//					},
	//				},
	//			},
	//		},
	//		Fields: "userEnteredValue",
	//	},
	//}

	formulaValue := fmt.Sprintf("=HYPERLINK(\"#gid=%d\",\"%s\")", sheetId, sheetName)

	// シートへのリンクを追加（ハイパーリンクの形式で）
	appendLinkRequest := &sheets.Request{
		UpdateCells: &sheets.UpdateCellsRequest{
			Start: &sheets.GridCoordinate{
				SheetId:     indexSheetId, // インデックスページのID
				RowIndex:    int64(rowIndex),
				ColumnIndex: 1, // リンクの列
			},
			Rows: []*sheets.RowData{
				{
					Values: []*sheets.CellData{
						{
							UserEnteredValue: &sheets.ExtendedValue{
								FormulaValue: &formulaValue,
							},
						},
					},
				},
			},
			Fields: "userEnteredValue",
		},
	}

	// 両方のリクエストを組み合わせて返す
	return []*sheets.Request{appendLinkRequest}
}
