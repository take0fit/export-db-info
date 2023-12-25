package google

import (
	"context"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// InitializeGoogleClient はGoogle APIクライアントを初期化します。
func InitializeGoogleClient(ctx context.Context, credentialsFilePath, apiScope string) (*http.Client, error) {
	b, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(b, apiScope)
	if err != nil {
		return nil, err
	}

	return config.Client(ctx), nil
}

// InitializeSheetsClient はGoogle Sheets APIクライアントを初期化します。
func InitializeSheetsClient(ctx context.Context, credentialsFilePath string) (*sheets.Service, error) {
	client, err := InitializeGoogleClient(ctx, credentialsFilePath, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}

	return sheets.NewService(ctx, option.WithHTTPClient(client))
}

// InitializeDriveClient はGoogle Drive APIクライアントを初期化します。
func InitializeDriveClient(ctx context.Context, credentialsFilePath string) (*drive.Service, error) {
	client, err := InitializeGoogleClient(ctx, credentialsFilePath, drive.DriveScope)
	if err != nil {
		return nil, err
	}

	return drive.NewService(ctx, option.WithHTTPClient(client))
}
