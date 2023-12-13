package google_sheets

import (
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
)

// InitializeSheetsClient はGoogle Sheets APIクライアントを初期化します。
func InitializeSheetsClient(ctx context.Context, credentialsFilePath string) (*sheets.Service, error) {
	b, err := ioutil.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}

	ts := config.TokenSource(ctx)

	srv, err := sheets.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}

	return srv, nil
}
