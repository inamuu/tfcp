# tfcp - Terraform Copy Tool

Terraformのplan/apply結果から不要な出力を除去し、クリーンな結果をクリップボードにコピーするGoツールです。

## 機能

- Terraformのplan/apply結果をフィルタリング
- "Refreshing state..."行の除去
- Warning セクションの除去
- フィルタリング後の結果をpbcopyでクリップボードにコピー
- パイプライン、ターミナル履歴、標準入力の3つの入力方法をサポート

## インストール

```bash
go build -o tfcp main.go
# /usr/local/bin などのPATHが通った場所に配置
sudo mv tfcp /usr/local/bin/
```

## 使用方法

### 1. パイプライン経由（推奨）

```bash
terraform plan | tfcp
terraform apply | tfcp
```

### 2. ターミナル履歴から自動取得（macOS Terminal.appのみ）

```bash
# terraform plan を実行した直後に
tfcp

# カスタム行数指定
tfcp -lines 10000
```

### 3. 標準入力から手動入力

```bash
tfcp -stdin
# Terraformの出力を貼り付けて Ctrl+D で終了
```

## オプション

- `-lines N`: ターミナル履歴を遡る行数（デフォルト: 5000）
- `-stdin`: 標準入力から読み取りモード
- `-h`, `-help`: ヘルプ表示

## フィルタリング対象

### 除去される項目

1. **Refreshing state 行**
   ```
   aws_instance.example: Refreshing state...
   ```

2. **Warning セクション**
   ```
   ╷
   │ Warning: Argument is deprecated
   │
   │   with module.iam.aws_iam_role.lambda_sidekiq_monitoring,
   │   on modules/iam/lambda.tf line 293, in resource "aws_iam_role" "lambda_sidekiq_monitoring":
   │  293: resource "aws_iam_role" "lambda_sidekiq_monitoring" {
   │
   │ inline_policy is deprecated. Use the aws_iam_role_policy resource instead.
   ╵
   ```

### 保持される項目

- Plan: の出力
- リソースの変更内容（+, -, ~, -/+）
- Apply complete! メッセージ
- エラーメッセージ
- その他の重要な出力

## 要件

- macOS (pbcopyコマンドが必要)
- Go 1.21以上（ビルド時）

## 例

```bash
# Terraformプランを実行してクリーンな結果をコピー
terraform plan | tfcp

# 実行結果
Terraformの出力 (45行) をクリップボードにコピーしました
```

コピーされた内容は警告や不要な出力が除去され、重要な変更内容のみが含まれます。 