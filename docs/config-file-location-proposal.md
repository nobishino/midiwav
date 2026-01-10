# 設定ファイルの配置場所に関する提案

## 課題
Issue #7 のコメント（#3733671954）で提起された「WindowsまたはMacで動かすので、設定ファイルの適切な置き場所を決めたい」について検討した結果を報告します。

## 推奨する設定ファイル配置場所

### 1. プラットフォーム別の標準的な配置場所

#### Windows
```
%APPDATA%\midiwav\config.toml
```
例: `C:\Users\<Username>\AppData\Roaming\midiwav\config.toml`

#### macOS
```
~/Library/Application Support/midiwav/config.toml
```
または
```
~/.config/midiwav/config.toml
```

#### Linux（参考）
```
~/.config/midiwav/config.toml
```

### 2. 推奨される実装方法

#### オプション A: XDG Base Directory 仕様に準拠（推奨）

Goの標準的なライブラリ `github.com/adrg/xdg` を使用することで、クロスプラットフォーム対応が容易になります。

**メリット:**
- Windows、macOS、Linuxで自動的に適切な場所を選択
- 業界標準に準拠
- メンテナンスが容易

**実装例:**
```go
import "github.com/adrg/xdg"

configPath, err := xdg.ConfigFile("midiwav/config.toml")
```

これにより以下のように解決されます:
- Windows: `%APPDATA%\midiwav\config.toml`
- macOS: `~/Library/Application Support/midiwav/config.toml`
- Linux: `~/.config/midiwav/config.toml`

#### オプション B: カスタム実装

標準ライブラリのみで実装する場合:

```go
import (
    "os"
    "path/filepath"
    "runtime"
)

func getConfigPath() (string, error) {
    var baseDir string
    
    switch runtime.GOOS {
    case "windows":
        baseDir = os.Getenv("APPDATA")
    case "darwin": // macOS
        home, _ := os.UserHomeDir()
        baseDir = filepath.Join(home, "Library", "Application Support")
    default: // Linux等
        configDir := os.Getenv("XDG_CONFIG_HOME")
        if configDir == "" {
            home, _ := os.UserHomeDir()
            configDir = filepath.Join(home, ".config")
        }
        baseDir = configDir
    }
    
    return filepath.Join(baseDir, "midiwav", "config.toml"), nil
}
```

### 3. 設定ファイルの検索順序（推奨）

柔軟性を高めるため、以下の順序で設定ファイルを検索することを推奨します:

1. コマンドライン引数で指定されたパス: `--config /path/to/config.toml`
2. 環境変数: `MIDIWAV_CONFIG=/path/to/config.toml`
3. カレントディレクトリ: `./config.toml`
4. ユーザー設定ディレクトリ（上記で説明した場所）

これにより、開発時やテスト時の利便性を保ちつつ、本番環境での適切な配置も可能になります。

### 4. 設定ファイルの形式

Issue #7 で提案されているTOML形式は適切な選択です:

```toml
[[target]]
dir = "/abs/path/to/dir1"
discord_webhook_url = "https://discord.com/api/webhooks/XXX/YYY"

[[target]]
dir = "/abs/path/to/dir2"
recursive = true
# discord_webhook_url を省略すると Discord へ投稿しません。
```

TOMLのメリット:
- 人間が読み書きしやすい
- Goでは `github.com/BurntSushi/toml` などの成熟したライブラリがある
- 配列やテーブルの表現が直感的

## 実装の推奨手順

1. `github.com/adrg/xdg` と `github.com/BurntSushi/toml` を依存関係に追加
2. 設定ファイル構造体を定義
3. 設定ファイル読み込み機能を実装
4. 既存の環境変数による設定を後方互換性のために維持（環境変数 > 設定ファイルの優先順位）

## 結論

**推奨: オプション A（xdg ライブラリの使用）**

理由:
- クロスプラットフォーム対応が自動的に行われる
- 業界標準に準拠しており、ユーザーにとって予測可能
- コードがシンプルで保守しやすい
- 多くのGoアプリケーションで採用されている実績がある

この提案により、WindowsとmacOSの両方で適切な設定ファイルの配置が実現できます。
