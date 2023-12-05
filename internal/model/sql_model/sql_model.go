package sql_model

// DB はデータベース全体の情報を保持します。
type DB struct {
	Name   string   // データベース名
	Tables []*Table // データベースに含まれるテーブルのスライス
}

// Table はデータベースのテーブル情報を表します。
type Table struct {
	Name    string    // テーブル名
	Columns []*Column // テーブルのカラム情報
}

// Column はデータベースのカラム情報を表します。
type Column struct {
	Name             string // カラム名
	Type             string // データ型
	IsNullable       bool   // NULL値を許容するか
	Default          string // デフォルト値
	Comment          string // コメント
	IsPrimaryKey     bool   // プライマリーキーかどうか
	IsUnique         bool   // ユニーク制約があるかどうか
	IsIndexed        bool   // インデックスが貼られているか
	IsForeign        bool   // インデックスが貼られているか
	ForeignKeyTable  string // 外部キーとして参照しているテーブル名
	ForeignKeyColumn string // 外部キーとして参照しているテーブルのカラム名
}
