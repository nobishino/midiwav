# Issue #7 Comment #3733671954への回答

## 質問
「WindowsまたはMacで動かすので、設定ファイルの適切な置き場所を決めたい。」

## 推奨案

### 設定ファイルの配置場所

**Windows:**
```
%APPDATA%\midiwav\config.toml
```

**macOS:**
```
~/Library/Application Support/midiwav/config.toml
```

### 実装方法の推奨

**推奨: `github.com/adrg/xdg` ライブラリを使用**

このライブラリを使用すると、以下のコードで自動的にプラットフォーム適切なパスが得られます:

```go
import "github.com/adrg/xdg"

configPath, err := xdg.ConfigFile("midiwav/config.toml")
```

**理由:**
- クロスプラットフォーム対応が自動的に行われる
- 業界標準（XDG Base Directory仕様）に準拠
- コードがシンプルで保守しやすい
- 多くのGoアプリケーションで採用実績がある

### 柔軟性のための設定ファイル検索順序

1. コマンドライン引数: `--config /path/to/config.toml`
2. 環境変数: `MIDIWAV_CONFIG`
3. カレントディレクトリ: `./config.toml`
4. ユーザー設定ディレクトリ（上記の推奨場所）

これにより、開発時とテスト時の利便性を保ちつつ、本番環境での適切な配置も可能になります。

### 実装の優先順位

1. 既存の環境変数（`MIDIWAV_DIR`、`MIDIWAV_DISCORD_WEBHOOK`）による設定を後方互換性のために維持
2. 環境変数が設定されていない場合に設定ファイルを読み込む
3. 将来的に環境変数は非推奨として、設定ファイルに移行

## 詳細

より詳細な分析と実装例については、以下のドキュメントを参照してください:
- [設定ファイル配置場所の提案（日本語）](docs/config-file-location-proposal.md)
- [Configuration File Location Proposal (English)](docs/config-file-location-proposal.en.md)
