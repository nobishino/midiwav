# Setup

## Configuration File

Create a TOML configuration file with the following format:

```toml
[[target]]
dir = "/abs/path/to/dir1"
discord_webhook_url = "https://discord.com/api/webhooks/XXX/YYY"

[[target]]
dir = "/abs/path/to/dir2"
recursive = true # trueの場合、配下のディレクトリを再帰的に検索します。デフォルトはfalseです。
# discord_webhook_url を省略すると Discord へ投稿しません。

# 楽譜画像の生成に使う外部コマンド（セクションごと省略可）。
# 4声体和声とみなせるMIDIに対して、楽譜（.musicxml とPNG画像）を生成・添付します。
# コマンドが見つからない場合、その段階までのファイル生成にとどめます
# （verovioなし → .musicxml のみ、rsvg-convertなし → SVGのまま）。
[notation]
verovio_path = "verovio"        # MusicXML→SVG。brew install verovio などで導入
svg2png_path = "rsvg-convert"   # SVG→PNG。brew install librsvg などで導入
```

## Running

```sh
midiwav -config /path/to/config.toml
```

## Package Layout

- `.` (main): CLI。ディレクトリ監視・設定・Discord投稿
- `harmony`: 4声体和声の分析・添削（禁則チェック、和音記号、コードネーム）
- `synth`: MIDIからの矩形波によるWAV合成

## Testing

Run tests:
```sh
go test ./...
```

Update golden files (`synth/testdata/*.wav` and `harmony/testdata/*.golden.txt`) with new output:
```sh
go test ./... -update
```

### 和声分析のバグ再現テストを追加する

和声分析（4声体和声の添削）の誤検出・見逃しを再現するには、問題のMIDIファイルを
`harmony/testdata/` に置くだけでテストケースになります。

1. MIDIファイルを `harmony/testdata/<name>_<key>.mid` として保存する。
   調はファイル名でドイツ語音名により指定します（例: `kadai3_es-moll.mid`）。
   調の指定がないファイルは、調に依存しない検査のみが実行されます。
2. `go test ./... -update` を実行して期待レポート `<name>_<key>.golden.txt` を生成する。
3. 生成されたレポートの内容が正しいことを確認してコミットする。
   （バグ再現の場合は、期待する正しいレポートに手で修正してからコミットすると、
   修正までテストが失敗し続けるリグレッションテストになります）