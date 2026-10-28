# 開発・ビルド手順（Docker不要）

WindowsにGo 1.24以降とMicrosoft Edge WebView2 Runtimeを用意し、プロジェクトのルートで実行します。初回の依存パッケージ取得には接続が必要です。通常プレイには接続不要です。

```powershell
go mod download
go test ./... -count=1
go vet ./...
go build -tags "desktop,production" -ldflags "-H windowsgui" -o card-game-lounge.exe .
.\card-game-lounge.exe
```

フロントエンドは素のHTML・CSS・JavaScriptなので、npmビルドはありません。Wails CLIを利用する場合はwails.jsonを使えます。

## 構成

- frontend/index.html、suite.js、suite.css：統合トップとゲーム切替。
- frontend/poker、blackjack、baccarat：各ゲームの画面。ゲームごとに同一オリジンのフレームで表示し、親画面のWailsバインディングへ接続。
- poker_app.go、blackjack_app.go、baccarat_app.go：独立したGoの状態管理。フレームを閉じてもアプリ終了までは状態を保持。
- internal/poker、blackjack、baccarat：ルールと自動判断。
- frontend/shared、asset-map.json、assets.go：重複する画像・音声を1回だけ埋め込み、元の各ゲームのURLから読み出す仕組み。

子画面の名前はgame.htmlです。Wailsのランタイム自動挿入を親のindex.htmlだけに限定し、二重の通信処理を避けています。親のGo API呼び出しなので、外部サーバーを起動する必要はありません。

設定キーはcardlounge.*です。ゲーム間でキャラクターと音量・速度を共有し、各ゲームへ入り直す際に反映します。Windowsの画面サイズ設定はユーザー設定フォルダのCardGameLounge/window.jsonに保存します。

トップの挨拶はfrontend/lobby-greeting.jsで制御し、各人の直前のお誘いを除いてランダムに選びます。ゲーム開始・女性の切替・タブの非表示時に古い音声を止めます。途中で別の女性を選んでも声は重なりません。起動時はMISAKIを選択し、画面の表示後に自動で挨拶します。音声OFF・音量0の設定は尊重します。一覧へ戻っただけでは起動時の挨拶を繰り返しません。

## 現行の画像・音声更新

5人の声はdocs/five-cast-voice-profiles.json、全原稿はscripts/five-cast-voice-plan.json、生成はscripts/build-five-cast-voices.pyで管理します。既存AivisSpeechを起動し、librosa/numpyが利用できるPythonで生成します。古いVOICEVOX生成スクリプトは資料です。

録音はfrontend/cast-voiceへ共有保存し、asset-map.jsonでゲーム別の仮想URLへ割り当てます。各ゲームの通常・お祝いマニフェストは各assets/voice*直下、トップはlobby-voice/manifest.json。追加画像はcast-imagesへ保存します。更新後は全Goテスト・UI検証・EXE再ビルド・SHA256SUMS.txtの更新を行います。

## UIテスト

PlaywrightとChromeを用意します。

```powershell
go test -c -o build/ui-test.exe .
$env:PLAYWRIGHT_MODULE = 'Playwrightモジュールの絶対パス'
node scripts/ui-check.cjs
```

標準入力のテスト専用ブリッジで本物のGoゲームエンジンをブラウザへ接続します。テストexeは通常版には不要です。

scripts/native-check.cjsは完成exeのWebView2へローカルデバッグ接続して確認する任意の開発用スクリプトです。実行時だけ9237番のデバッグポートを指定し、通常起動時には使いません。納品時は作業環境のWebView2起動エラーにより未完了です。

## 現行の声とウィンドウ

声は参照版の人物別設定、MISAKIは長さを維持した1.10倍の音程補正。標準再生速度1倍、挨拶も同倍率。詳細は5名対応・音声更新-20261028.md。

internal/windowstateのOptimalでモニターを選び、従来の通常サイズの90%を中央に表示します。起動時は最大化しません。DPI、負のモニター座標、取り外し対応は維持します。

## タイトル画面

frontend/suite-title.jsとsuite-title.css、lounge-title-five.pngで構成します。画像を読み込み、画面が表示されてから5秒でゲーム一覧へ進みます。ボタン・Enter・Space・Escapeでも進めます。遷移は一度だけ実行し、MISAKIの自動挨拶はタイトル終了後に開始します。背景画像は全体が見えるようcontainで表示し、縦長画面でも3人が切れないようにしています。node scripts/title-check.cjsで自動遷移・ボタン遷移・音声開始タイミングを確認できます。

## ホームボタン（2026-10-18）
各ゲーム上部の共通ヘッダーに「ホーム」を表示し、既存のleave処理でゲーム一覧へ戻る。配札・精算中の移動制限、音声停止、アプリ内の進行保持を維持する。

## ラウンジBGM
frontend/lobby-music.jsがTOP・一覧の曲を管理し、lobby-greeting.jsの発話通知で減音します。曲の再生成はpython scripts/build-lobby-bgm.py、検証はnode scripts/top-bgm-check.cjs。各ゲームへ移動すると親の曲を停止し、ホーム復帰時に再開します。
