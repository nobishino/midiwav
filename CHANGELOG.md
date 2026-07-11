# Changelog

## [v0.1.4](https://github.com/nobishino/midiwav/compare/v0.1.3...v0.1.4) - 2026-07-11

- 4声体和声のMIDI禁則チェック機能を追加 by @nobishino in https://github.com/nobishino/midiwav/pull/23
- 調が既知の場合に和音記号（度数・種類・転回位置）を判定して表示する by @nobishino in https://github.com/nobishino/midiwav/pull/26
- 和音一覧にコードネーム表記を追加する by @nobishino in https://github.com/nobishino/midiwav/pull/28
- 和声分析のゴールデンファイルテストを追加する by @nobishino in https://github.com/nobishino/midiwav/pull/34
- WAV正規化で負のピークを考慮する by @nobishino in https://github.com/nobishino/midiwav/pull/35
- テンポマップを導入してテンポ処理を改善する by @nobishino in https://github.com/nobishino/midiwav/pull/37
- 和音一覧の声部をB-T-A-Sの順で表示する by @nobishino in https://github.com/nobishino/midiwav/pull/38
- 和声分析機能とwav変換機能を別パッケージに分離する by @nobishino in https://github.com/nobishino/midiwav/pull/39
- IV7の和音への対応、増6の和音への対応 by @nobishino in https://github.com/nobishino/midiwav/pull/40
- 増六の和音で第9音が省略された場合にV7と判定するよう修正 by @nobishino in https://github.com/nobishino/midiwav/pull/42
- 楽譜出力用の中間表現Scoreを追加する（#43 その1） by @nobishino in https://github.com/nobishino/midiwav/pull/44
- 和声分析結果のMusicXML生成を追加する（#43 その2） by @nobishino in https://github.com/nobishino/midiwav/pull/45
- 楽譜画像を生成してDiscordに添付する（#43 その3） by @nobishino in https://github.com/nobishino/midiwav/pull/46
- 楽譜出力に臨時記号（accidental）を反映する by @nobishino in https://github.com/nobishino/midiwav/pull/50
- 楽譜上の和音記号をコンパクトな表記にする by @nobishino in https://github.com/nobishino/midiwav/pull/51
- 楽譜冒頭のパート名「SATB」を表示しない by @nobishino in https://github.com/nobishino/midiwav/pull/52
- WAV書き出しをバッファリングして高速化する（#53 対応1） by @nobishino in https://github.com/nobishino/midiwav/pull/54
- 和声分析のみを実行する check サブコマンドを追加する by @nobishino in https://github.com/nobishino/midiwav/pull/55
- 和音記号ではなくコードネームを表示する by @nobishino in https://github.com/nobishino/midiwav/pull/57
- 楽譜上のコードネームを斜体でなく直立体で表示する by @nobishino in https://github.com/nobishino/midiwav/pull/58

## [v0.1.3](https://github.com/nobishino/midiwav/compare/v0.1.2...v0.1.3) - 2026-01-21
- Trim leading and trailing silence from WAV output by @Copilot in https://github.com/nobishino/midiwav/pull/20

## [v0.1.2](https://github.com/nobishino/midiwav/compare/v0.1.1...v0.1.2) - 2026-01-11
- Implement golden file testing for MIDI to WAV conversion by @Copilot in https://github.com/nobishino/midiwav/pull/9
- Add CI workflow for go test, go fmt, and go mod tidy verification by @Copilot in https://github.com/nobishino/midiwav/pull/11
- Remove unused code and commented-out blocks by @Copilot in https://github.com/nobishino/midiwav/pull/10
- tagpr導入: mainブランチに追従するリリースPRを作る by @nobishino in https://github.com/nobishino/midiwav/pull/13
- Add Copilot onboarding instructions for midiwav repository by @Copilot in https://github.com/nobishino/midiwav/pull/14
- Add TOML config file support for multiple target directories by @Copilot in https://github.com/nobishino/midiwav/pull/12
- nits: Refactor by @nobishino in https://github.com/nobishino/midiwav/pull/17

## [v0.1.1](https://github.com/nobishino/midiwav/compare/main-d3789e5908fab8e6c56f36cc8318aa00c4b6eb4f...v0.1.1) - 2026-01-10
- Goreleaserを使うようにする by @nobishino in https://github.com/nobishino/midiwav/pull/6
