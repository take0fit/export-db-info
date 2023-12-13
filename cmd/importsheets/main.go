package main

import (
	"context"
	"encoding/csv"
	"export-db-info/pkg/lib/google_sheets"
	"fmt"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {
	ctx := context.Background()
	// Google Sheets APIの認証情報
	serviceAccountKeyFile := os.Getenv("GOOGLE_SERVICE_ACCOUNT_KEY_FILE")

	// Google Sheets APIクライアントの初期化
	srv, err := google_sheets.InitializeSheetsClient(ctx, serviceAccountKeyFile)
	if err != nil {
		log.Fatalf("Error initializing Google Sheets client: %v", err)
	}

	// CSVファイルが保存されているディレクトリのパス
	csvDir := os.Getenv("CSV_DIRECTORY")
	files, err := ioutil.ReadDir(csvDir)
	if err != nil {
		log.Fatalf("Unable to read directory: %v", err)
	}

	spreadsheetId := "your_spreadsheet_id" // スプレッドシートID

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

			// 新しいシートの作成
			newSheet := &sheets.SheetProperties{
				Title: file.Name(),
			}
			addSheetRequest := &sheets.AddSheetRequest{
				Properties: newSheet,
			}

			sheetTitle := "テーブル仕様書" // 新しいシートのタイトル

			requests := []*sheets.Request{
				// シートのタイトルを設定
				{
					AddSheet: &sheets.AddSheetRequest{
						Properties: &sheets.SheetProperties{
							Title: sheetTitle,
						},
					},
				},
				// セルの結合: A1:C1
				{
					MergeCells: &sheets.MergeCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          0, // 新しいシートのIDが0であることを仮定
							StartRowIndex:    0,
							EndRowIndex:      1,
							StartColumnIndex: 0,
							EndColumnIndex:   3,
						},
						MergeType: "MERGE_ALL",
					},
				},
				// セルの結合: D1:F1
				{
					MergeCells: &sheets.MergeCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          0,
							StartRowIndex:    0,
							EndRowIndex:      1,
							StartColumnIndex: 3,
							EndColumnIndex:   6,
						},
						MergeType: "MERGE_ALL",
					},
				},
				// セルの結合: G1:K1
				{
					MergeCells: &sheets.MergeCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          0,
							StartRowIndex:    0,
							EndRowIndex:      1,
							StartColumnIndex: 6,
							EndColumnIndex:   11,
						},
						MergeType: "MERGE_ALL",
					},
				},
				// セルの背景色の設定: A1:K1
				{
					RepeatCell: &sheets.RepeatCellRequest{
						Range: &sheets.GridRange{
							SheetId:          0,
							StartRowIndex:    0,
							EndRowIndex:      2, // 2行目までの範囲で背景色を設定
							StartColumnIndex: 0,
							EndColumnIndex:   11,
						},
						Cell: &sheets.CellData{
							UserEnteredFormat: &sheets.CellFormat{
								BackgroundColor: &sheets.Color{
									Red:   0.0,
									Green: 0.0,
									Blue:  0.0,
								},
								HorizontalAlignment: "CENTER", // 中央揃え
								TextFormat: &sheets.TextFormat{
									ForegroundColor: &sheets.Color{
										Red:   1.0,
										Green: 1.0,
										Blue:  1.0,
									},
									FontSize: 12,
									Bold:     true,
								},
							},
						},
						Fields: "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
					},
				},
				// A1:C1にテキスト「テーブル仕様書」を挿入
				{
					UpdateCells: &sheets.UpdateCellsRequest{
						Start: &sheets.GridCoordinate{
							SheetId:     0,
							RowIndex:    0, // 1行目
							ColumnIndex: 0, // A列
						},
						Rows: []*sheets.RowData{
							{
								Values: []*sheets.CellData{
									{
										UserEnteredValue: &sheets.ExtendedValue{
											StringValue: &sheetTitle,
										},
									},
								},
							},
						},
						Fields: "userEnteredValue",
					},
				},
				// D1:F1にテキスト「テーブル」を挿入
				{
					UpdateCells: &sheets.UpdateCellsRequest{
						Start: &sheets.GridCoordinate{
							SheetId:     0,
							RowIndex:    0, // 1行目
							ColumnIndex: 3, // D列
						},
						Rows: []*sheets.RowData{
							{
								Values: []*sheets.CellData{
									{
										UserEnteredValue: &sheets.ExtendedValue{
											StringValue: &sheetTitle,
										},
									},
								},
							},
						},
						Fields: "userEnteredValue",
					},
				},
				{
					AddSheet: addSheetRequest,
				},
			}

			batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
				Requests: requests,
			}
			resp, err := srv.Spreadsheets.BatchUpdate(spreadsheetId, batchUpdateRequest).Do()
			if err != nil {
				log.Printf("Unable to create new sheet: %v", err)
				continue
			}

			// 新しいシートのIDを取得
			var newSheetId int64
			if len(resp.Replies) > 0 && resp.Replies[0].AddSheet != nil {
				newSheetId = resp.Replies[0].AddSheet.Properties.SheetId
			} else {
				log.Fatal("Failed to get the new sheet ID")
			}

			fmt.Println(newSheetId)

			// データの書き込み
			var vr sheets.ValueRange
			for _, record := range records {
				// record（[]string）を[]interface{}に変換
				var interfaceRecord []interface{}
				for _, field := range record {
					interfaceRecord = append(interfaceRecord, field)
				}

				vr.Values = append(vr.Values, interfaceRecord)
			}
			_, err = srv.Spreadsheets.Values.
				Update(spreadsheetId, newSheet.Title+"!A1", &vr).
				ValueInputOption("USER_ENTERED").
				Do()
			if err != nil {
				log.Printf("Unable to write data to new sheet: %v", err)
				continue
			}
		}
	}
}
