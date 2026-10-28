# Gitに登録するファイル

このフォルダーは2026-10-28時点でGit未初期化。今回の作業は.gitignore作成のみで、git init・commit・pushは行っていません。

## 登録するもの

- .gitignore、README.md、AGENTS.md
- *.go（テストを含む）、go.mod、go.sum、wails.json
- internal/：ゲームルールなどのソース
- frontend/：HTML・CSS・JavaScript、asset-map.json、画像・音声・BGMすべて。GoのembedでEXEへ組み込むため必要
- scripts/：ビルド・素材生成・検証のスクリプトと設定
- docs/：仕様書・変更履歴・開発手順、素材ハッシュ・生成設定などのJSON。ローカル検証の画像と一部の実行結果は除外

## 登録しないもの

- EXE・ZIPなどの配布物、SHA256SUMS.txt（配布EXE用。GitHub Releasesなどへ添付）
- build/、.tools/、キャッシュ・依存パッケージ・一時ファイル
- 検証スクリーンショットや実戦操作ログなどの生成物
- .env、秘密鍵などの秘密情報
- ローカル整理用の案内・復元当時の管理ファイル

画像・音声を一括除外しないでください。frontendは現在約280MBあるため、リポジトリ全体の容量は大きくなります。単一ファイル100MiB以上がある場合は、GitHubへpushする前にGit LFSなどの対応が必要です。今回、100MiB以上のfrontend素材はありません。

既にGitで管理されているファイルは.gitignore追加だけでは管理解除されません。その場合のみgit rm --cachedで対象を指定します。作業フォルダーの実ファイルは残してください。

初回登録前にgit status --shortとgit diff --cachedを確認し、秘密情報や不要な配布物が含まれていないか見直します。素材を公開する場合は公開できる権利のあるものだけにしてください。
