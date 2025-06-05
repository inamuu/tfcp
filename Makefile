.PHONY: build install test clean help

# デフォルトターゲット
help:
	@echo "利用可能なコマンド:"
	@echo "  make build   - tfcpツールをビルド"
	@echo "  make install - /usr/local/binにインストール"
	@echo "  make test    - テスト実行"
	@echo "  make clean   - ビルド成果物を削除"

# ビルド
build:
	go build -o tfcp main.go
	@echo "ビルド完了: ./tfcp"

# システムにインストール
install: build
	sudo mv tfcp /usr/local/bin/
	@echo "インストール完了: /usr/local/bin/tfcp"

# テスト実行
test: build
	@echo "=== tfcpツールのテスト ==="
	@echo "テストファイルを使用してフィルタリングをテスト中..."
	@cat test_terraform_output.txt | ./tfcp
	@echo ""
	@echo "=== フィルタリング結果 ==="
	@echo "クリップボードの内容:"
	@pbpaste

# クリーンアップ
clean:
	rm -f tfcp
	@echo "ビルド成果物を削除しました" 