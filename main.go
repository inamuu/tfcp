package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	defaultBufferLines = 5000
)

var (
	// Warningセクションの開始と終了を検出するための正規表現
	warningStartRegex = regexp.MustCompile(`^╷$`)
	warningEndRegex   = regexp.MustCompile(`^╵$`)

	// "Refreshing state..." を検出するための正規表現
	refreshingStateRegex = regexp.MustCompile(`^\s*.*: Refreshing state\.\.\.`)
)

func main() {
	var bufferLines int
	var fromStdin bool
	flag.IntVar(&bufferLines, "lines", defaultBufferLines, "ターミナル履歴を遡る行数（履歴モード時）")
	flag.BoolVar(&fromStdin, "stdin", false, "標準入力から読み取る")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "使用方法:\n")
		fmt.Fprintf(os.Stderr, "  terraform plan | tfcp           # パイプライン経由で使用\n")
		fmt.Fprintf(os.Stderr, "  tfcp                            # ターミナル履歴から検索\n")
		fmt.Fprintf(os.Stderr, "  tfcp -stdin                     # 標準入力から読み取り\n")
		fmt.Fprintf(os.Stderr, "\nフラグ:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	var input string
	var err error

	// 標準入力がパイプされているかチェック
	if isPipeInput() || fromStdin {
		input, err = readFromStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: 標準入力の読み取りに失敗しました: %v\n", err)
			os.Exit(1)
		}
	} else {
		// ターミナル履歴を読み取る
		input, err = getRecentTerminalOutput(bufferLines)
		if err != nil {
			fmt.Fprintf(os.Stderr, "エラー: ターミナル出力の取得に失敗しました: %v\n", err)
			fmt.Fprintf(os.Stderr, "以下のコマンドでTerraformの出力をパイプしてください:\n")
			fmt.Fprintf(os.Stderr, "  terraform plan | tfcp\n")
			fmt.Fprintf(os.Stderr, "または、-stdinフラグを使用して手動で貼り付けてください:\n")
			fmt.Fprintf(os.Stderr, "  tfcp -stdin\n")
			os.Exit(1)
		}
	}

	// Terraform結果をフィルタリング
	filtered := filterTerraformOutput(input)

	if len(strings.TrimSpace(filtered)) == 0 {
		fmt.Fprintf(os.Stderr, "有効なTerraformの出力が見つかりませんでした\n")
		os.Exit(1)
	}

	// pbcopyにコピー
	err = copyToPbcopy(filtered)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: pbcopyへのコピーに失敗しました: %v\n", err)
		os.Exit(1)
	}

	filteredLines := strings.Split(strings.TrimSpace(filtered), "\n")
	fmt.Printf("Terraformの出力 (%d行) をクリップボードにコピーしました\n", len(filteredLines))
}

// 標準入力がパイプされているかチェック
func isPipeInput() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// 最近のターミナル出力を取得（簡易版）
func getRecentTerminalOutput(lines int) (string, error) {
	// macOSでOSAスクリプトを使ってTerminalアプリの内容を取得する試み
	script := fmt.Sprintf(`
		tell application "Terminal"
			if (count of windows) > 0 then
				get contents of selected tab of front window
			else
				error "No terminal windows open"
			end if
		end tell
	`)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("AppleScriptでターミナル内容を取得できませんでした: %w", err)
	}

	// 出力の最後の指定行数を取得
	outputLines := strings.Split(string(output), "\n")
	start := 0
	if len(outputLines) > lines {
		start = len(outputLines) - lines
	}

	return strings.Join(outputLines[start:], "\n"), nil
}

// 標準入力から読み取る
func readFromStdin() (string, error) {
	if !isPipeInput() {
		fmt.Println("Terraformの出力を以下に貼り付けてください（Ctrl+Dで終了）:")
	}
	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Terraformの出力をフィルタリング
func filterTerraformOutput(input string) string {
	lines := strings.Split(input, "\n")
	var result []string
	var inWarning bool

	for _, line := range lines {
		// Warningセクションの開始を検出
		if warningStartRegex.MatchString(line) {
			inWarning = true
			continue
		}

		// Warningセクションの終了を検出
		if warningEndRegex.MatchString(line) && inWarning {
			inWarning = false
			continue
		}

		// Warningセクション内の場合はスキップ
		if inWarning {
			continue
		}

		// "Refreshing state..." をスキップ
		if refreshingStateRegex.MatchString(line) {
			continue
		}

		// その他の行を保持
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// pbcopyにコピー
func copyToPbcopy(text string) error {
	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
