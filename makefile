# .env ファイルから環境変数を読み込む
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# CSVエクスポートコマンドのビルドと実行
exportcsv:
	cd cmd/exportcsv && go build -o exportcsv
	./cmd/exportcsv/exportcsv

# Googleスプレッドシートへのインポートコマンドのビルドと実行
importsheets:
	cd cmd/importsheets && go build -o importsheets
	./cmd/importsheets/importsheets

# 両方のコマンドを実行するターゲット
all: exportcsv importsheets

.PHONY: exportcsv importsheets all
