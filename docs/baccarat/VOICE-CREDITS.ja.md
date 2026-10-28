# 音声・画像クレジット（2026-10-28）

制作：**aiota**

音声はAivisSpeechによるローカル合成。リバーシテスト版の採用モデル・スタイル・話速・抑揚に準拠し、カードゲーム用の原稿を収録しています。

| 人物 | 音声クレジット | スタイル | モデル |
|---|---|---|---|
| MISAKI | AivisSpeech: 凛音エル | ノーマル | [モデルと利用条件](https://hub.aivis-project.com/aivm-models/f5017410-fbb5-49e1-97cb-e785f42e15f5) |
| AYAKA | AivisSpeech: まお | ふつー | [モデルと利用条件](https://hub.aivis-project.com/aivm-models/a59cb814-0083-4369-8542-f51a29e72af7) |
| RINA | AivisSpeech: みちのくあいり | 喜び | [モデルと利用条件](https://hub.aivis-project.com/aivm-models/1b2830f4-8cf1-4184-a0d9-3a1bace3a844) |
| SHIZUE | AivisSpeech: まい | ノーマル | [モデルと利用条件](https://hub.aivis-project.com/aivm-models/e9339137-2ae3-4d41-9394-fb757a7e61e6) |
| KATHY | AivisSpeech: コハク | ノーマル | [モデルと利用条件](https://hub.aivis-project.com/aivm-models/22e8ed77-94fe-4ef2-871f-a86f94e9a579) |

各モデルの利用条件はACML 1.0。音声エンジン・モデル・辞書は配布せず、生成済み録音をEXEに同梱します。配布時は人物別クレジットとモデルごとの利用条件を保持してください。

MISAKIは参照版の再生補正に合わせ、長さを維持して音程を1.10倍に処理しています。標準再生速度は1倍。抑揚ONは参照値、OFFはその0.85倍で収録しています。利用者の音量・速度変更、ミュートは引き続き利用できます。

トップの挨拶は5人×3ゲームのお誘い×2話法の30録音。各ゲームには5人それぞれ通常40セリフとお祝い1セリフを収録。同一原稿・設定の録音は共有しています。

MISAKI・AYAKA・RINAの画像、ゲーム中のBGM・効果音は既存素材を維持。SHIZUE・KATHYの通常全身・顔・夏服は、ユーザー指定のC:/work/game-org/reversi-queens/assets/charactersから取り込んだ実写風素材です。新たな画像生成は行っていません。夏服のお祝いで水着には変更しません。

音声設定：five-cast-voice-profiles.json。原稿：../scripts/five-cast-voice-plan.json。生成結果：five-cast-voice-build.json。再生成：../scripts/build-five-cast-voices.py。

WailsなどGo依存パッケージのライセンスはbaccarat/THIRD-PARTY-NOTICES.mdとbaccarat/licensesを参照してください。旧VOICEVOX用生成スクリプトは過去の資料で、今回の音声の再生成には使用しません。

起動TOP・ラウンジBGMの最新情報は../TOP5人とBGM-20261028.mdおよび../CREDITS.ja.mdを参照。
