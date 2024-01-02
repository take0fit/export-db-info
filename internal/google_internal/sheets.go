package google_internal

import (
	"export-db-info/internal/model/google_model"
	"fmt"
	"google.golang.org/api/sheets/v4"
)

// CreateSheetLayoutRequest は、指定されたレイアウトオプションに基づいてスプレッドシートのリクエストを生成します。
func CreateSheetLayoutRequest(
	sheetId int64,
	rangeOption *google_model.RangeOption,
	merge bool,
	hAlign string,
	vAlign string,
	bgColor *sheets.Color,
	textColor *sheets.Color,
	text string,
	link string,
	textFormat *sheets.TextFormat,
) []*sheets.Request {
	var requests []*sheets.Request

	// セルの結合
	if merge {
		requests = append(requests, &sheets.Request{
			MergeCells: &sheets.MergeCellsRequest{
				Range:     createGridRange(sheetId, rangeOption),
				MergeType: "MERGE_ALL",
			},
		})
	}

	// テキストのフォーマット設定を更新
	if textFormat == nil {
		textFormat = &sheets.TextFormat{}
	}
	textFormat.ForegroundColor = textColor

	// セルのフォーマット設定
	cellFormat := &sheets.CellFormat{
		BackgroundColor:     bgColor,
		HorizontalAlignment: hAlign,
		VerticalAlignment:   vAlign,
		TextFormat:          textFormat,
	}

	// テキスト内容の設定
	var cellValue *sheets.ExtendedValue
	if text != "" {
		if link != "" {
			// ハイパーリンク付きのテキスト
			formulaValue := fmt.Sprintf("=HYPERLINK(\"%s\", \"%s\")", link, text)
			cellValue = &sheets.ExtendedValue{
				FormulaValue: &formulaValue,
			}
		} else {
			// 通常のテキスト
			cellValue = &sheets.ExtendedValue{StringValue: &text}
		}
	}

	requests = append(requests, &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: createGridRange(sheetId, rangeOption),
			Cell: &sheets.CellData{
				UserEnteredFormat: cellFormat,
				UserEnteredValue:  cellValue,
			},
			Fields: "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment,textFormat),userEnteredValue",
		},
	})

	// 枠線の設定
	border := &sheets.Border{
		Style: "SOLID",
		Width: 1,
		Color: &sheets.Color{Red: 0, Green: 0, Blue: 0}, // 黒色
	}

	requests = append(requests, &sheets.Request{
		UpdateBorders: &sheets.UpdateBordersRequest{
			Range:           createGridRange(sheetId, rangeOption),
			Top:             border,
			Bottom:          border,
			Left:            border,
			Right:           border,
			InnerHorizontal: border,
			InnerVertical:   border,
		},
	})

	return requests
}

func createGridRange(sheetId int64, option *google_model.RangeOption) *sheets.GridRange {
	return &sheets.GridRange{
		SheetId:          sheetId,
		StartRowIndex:    option.StartRow,
		EndRowIndex:      option.EndRow,
		StartColumnIndex: option.StartCol,
		EndColumnIndex:   option.EndCol,
	}
}
