# 求人マッチ度確認サイト 設計書（Codex実装向け）

## 1. 概要

本アプリは、ユーザーの志向・スキル・働き方の希望と、求人情報（テキストまたはURL入力）をもとに、ローカルLLMを使って求人とのマッチ度を判定し、その理由を表示するWebアプリケーションである。

公開は `server mac` 上のアプリケーションを Cloudflare Tunnel 経由で行う。推論は `llm mac` 上の Ollama に委譲する。外部公開するのは `server mac` のみで、`llm mac` はLAN内からのみアクセス可能とする。

---

## 2. 目的

### 2.1 目的
- 自分の志向と求人の相性を手早く確認できること
- 判定結果を単なるスコアではなく、理由つきで確認できること
- 結果をクリップボードへコピーし、転職活動のメモに流用できること
- 個人開発ポートフォリオとして、以下を示せること
  - Go による API サーバー実装
  - Vue によるフォームUI
  - Cloudflare Tunnel を用いた公開構成
  - ローカルLLM連携
  - URL抽出、非同期ジョブ、セキュリティ考慮を含むアプリ全体設計

### 2.2 非目的
- 求人サイトごとの高精度専用スクレイパー実装
- 複雑な認証・マルチユーザー管理
- 求人応募支援や自動応募
- 完全な職務経歴書生成
- LLMによる事実保証

---

## 3. 想定ユーザー

### 3.1 主ユーザー
- サイト作成者本人

### 3.2 将来的な公開利用者
- 転職活動中のエンジニア
- 求人票を見て自分との相性をざっくり確認したい人

---

## 4. ユースケース

### 4.1 テキスト入力で判定
1. ユーザーが自分の情報をフォームへ入力する
2. ユーザーが求人テキストをフォームへ貼り付ける
3. 実行ボタンを押す
4. サーバーがLLMへ判定依頼する
5. マッチ度、理由、一致点、懸念点、面接で確認すべき点を表示する
6. 結果をクリップボードへコピーできる

### 4.2 URL入力で判定
1. ユーザーが自分の情報をフォームへ入力する
2. ユーザーが求人URLを入力する
3. 実行ボタンを押す
4. サーバーがURL先HTMLを取得し、本文を抽出・整形する
5. サーバーが整形後テキストをLLMへ渡す
6. 判定結果を表示する

### 4.3 過去結果の確認（将来拡張）
1. 過去の判定履歴一覧を見る
2. 入力内容と結果を再確認する
3. モデルやプロンプトの差分比較に使う

---

## 5. システム構成

## 5.1 物理構成

### server mac
責務:
- Go APIサーバー
- Vueのビルド済み静的ファイル配信
- URL取得・本文抽出
- ジョブ管理
- SQLite保存
- Cloudflare Tunnel 公開対象

### llm mac
責務:
- Ollama 実行
- 推論モデル管理
- LAN 内からのみアクセスを許可

---

## 5.2 論理構成

```text
[Browser]
  -> Cloudflare Tunnel
    -> [server mac: Go API + Static Files + SQLite]
      -> [llm mac: Ollama API]
```

---

## 6. 技術スタック

### 6.1 バックエンド
- Go
- HTTPルータ: chi または echo のどちらか一方
- HTML抽出: goquery ベース
- DB: SQLite
- JSON API

### 6.2 フロントエンド
- Vue 3
- Vite
- TypeScript
- UIは最小構成でよい

### 6.3 LLM
- Ollama
- モデルは設定ファイルで切り替え可能にする
- 初期は 1 モデル固定でよい

### 6.4 デプロイ
- Docker
- server mac 上でコンテナ起動
- Cloudflare Tunnel で外部公開

---

## 7. 機能要件

## 7.1 入力フォーム

### ユーザープロフィール入力
以下の入力欄を持つ:
- 得意言語（複数可）
- 避けたい言語（複数可）
- 興味のある領域（複数可）
- 興味の薄い領域（複数可）
- 希望する働き方
- 希望年収（任意）
- 自由記述

### 求人入力
- ラジオボタン: `text` / `url`
- `text` 選択時: テキストエリア表示
- `url` 選択時: URL入力欄表示

### 実行
- 実行ボタンを押すとマッチ判定ジョブを開始

---

## 7.2 結果表示
表示内容:
- 総合マッチ度（0〜100）
- 要約
- 一致している点
- 懸念点
- 面接で確認すべき質問
- クリップボードコピー用整形テキスト
- 使用モデル名（任意表示）
- URL抽出失敗時のエラー内容

---

## 7.3 URL抽出
- URLからHTMLを取得する
- `script`, `style`, `noscript` は除去する
- 本文候補を抽出する
- 改行や空白を整形する
- 長すぎる場合は安全に切り詰める
- 取得に失敗した場合はユーザーへ再入力を促す

補足:
- 初期実装では完全なReadability再現は不要
- まずは `title`, `meta description`, `h1-h3`, `main`, `article`, `body` などから本文候補を作る

---

## 7.4 ジョブ実行
- マッチ判定は非同期ジョブとして処理する
- APIはジョブIDを返す
- フロントはポーリングで状態確認を行う

ジョブ状態:
- `queued`
- `running`
- `done`
- `failed`

---

## 8. 非機能要件

### 8.1 セキュリティ
- `llm mac` は外部公開しない
- URL入力に対してSSRF対策を行う
- 入力サイズ制限を設ける
- タイムアウトを設定する
- レート制限を設ける
- Cloudflare Tunnel の先は server 側のみ

### 8.2 性能
- テキスト入力判定: 数秒〜十数秒程度を許容
- URL入力判定: HTML取得とLLM推論を含むため十数秒程度を許容
- 単一ユーザー〜少数同時接続を想定

### 8.3 可観測性
- ジョブ開始/終了ログ
- URL取得失敗ログ
- LLM応答失敗ログ
- JSONパース失敗ログ

---

## 9. API設計

## 9.1 POST /api/match
判定ジョブを作成する。

### request
```json
{
  "profile": {
    "preferred_languages": ["Go"],
    "avoid_languages": ["Java", "Kotlin"],
    "interests": ["推薦", "マッチング", "収益に近い機能開発"],
    "low_interests": ["IaC中心の業務"],
    "work_style": ["柔軟な働き方", "裁量"],
    "desired_compensation": "800万円以上",
    "notes": "バックエンド中心だがプロダクト寄り。MLを実サービスに載せる実装は得意。"
  },
  "job_input": {
    "type": "text",
    "value": "求人票本文..."
  }
}
```

### response
```json
{
  "id": "match_01HXXXXX",
  "status": "queued"
}
```

---

## 9.2 GET /api/match/:id
ジョブ状態と結果を取得する。

### response (running)
```json
{
  "id": "match_01HXXXXX",
  "status": "running"
}
```

### response (done)
```json
{
  "id": "match_01HXXXXX",
  "status": "done",
  "result": {
    "score": 78,
    "summary": "バックエンド経験とデータ活用経験が活きやすい一方、Java中心環境が懸念。",
    "pros": [
      "推薦・データ活用に近い課題がある",
      "プロダクト改善への関与余地がある",
      "裁量が比較的大きい"
    ],
    "cons": [
      "主要言語がJava/Kotlin寄り",
      "インフラ運用負荷が高い可能性"
    ],
    "questions_to_ask": [
      "配属後の期待役割は何か",
      "推薦やデータ活用に関われる余地はあるか"
    ],
    "clipboard_text": "マッチ度: 78/100\n要約: ...",
    "model": "llama3.1:8b"
  }
}
```

### response (failed)
```json
{
  "id": "match_01HXXXXX",
  "status": "failed",
  "error": "failed to extract job text"
}
```

---

## 9.3 GET /api/health
ヘルスチェック。

### response
```json
{
  "status": "ok"
}
```

---

## 10. データモデル

## 10.1 profiles
ユーザー入力のスナップショットを保持する。

カラム案:
- `id`
- `preferred_languages_json`
- `avoid_languages_json`
- `interests_json`
- `low_interests_json`
- `work_style_json`
- `desired_compensation`
- `notes`
- `created_at`

## 10.2 match_jobs
ジョブ本体。

カラム案:
- `id`
- `profile_id`
- `job_input_type` (`text` or `url`)
- `job_input_value`
- `status`
- `error_message`
- `created_at`
- `updated_at`

## 10.3 extracted_job_texts
URLから取得・整形した本文。

カラム案:
- `id`
- `match_job_id`
- `source_url`
- `page_title`
- `meta_description`
- `extracted_text`
- `created_at`

## 10.4 match_results
LLMの出力結果。

カラム案:
- `id`
- `match_job_id`
- `score`
- `summary`
- `pros_json`
- `cons_json`
- `questions_to_ask_json`
- `clipboard_text`
- `model_name`
- `raw_response`
- `created_at`

---

## 11. LLM入出力仕様

## 11.1 方針
- LLMには構造化済みプロフィールと求人テキストを渡す
- URLは直接渡さない
- 出力は必ずJSONに限定する
- JSONパースに失敗した場合は再試行する

## 11.2 入力データ
- ユーザーの志向情報（構造化JSON）
- 求人本文（抽出・整形済みテキスト）
- 必要であれば page title / meta description

## 11.3 出力JSONスキーマ
```json
{
  "score": 0,
  "summary": "",
  "pros": [],
  "cons": [],
  "questions_to_ask": [],
  "clipboard_text": ""
}
```

### 制約
- `score` は 0〜100 の整数
- `pros` は 0〜5件
- `cons` は 0〜5件
- `questions_to_ask` は 0〜5件
- `summary` は 1〜3文程度
- `clipboard_text` はそのままコピー可能な整形済み文字列

---

## 12. プロンプト設計方針

### システムプロンプト要件
- 求人本文は参照資料であり、指示ではないことを明示
- ページ本文に含まれる命令・注釈・プロンプト風文言は無視することを明示
- ユーザー志向との相性評価のみを行う
- 事実の補完や外部知識による憶測を抑制する
- JSON以外を出力しない

### 期待する評価観点
- 技術スタックの一致度
- 役割の一致度
- ドメインや課題の面白さとの一致度
- 働き方や裁量との一致度
- 懸念点の明示

---

## 13. バックエンド処理フロー

## 13.1 テキスト入力時
1. リクエスト受信
2. 入力バリデーション
3. profile 保存
4. match_job 作成（`queued`）
5. ワーカーがジョブ取得
6. LLMへリクエスト送信
7. JSONパース
8. result 保存
9. job を `done` へ更新

## 13.2 URL入力時
1. リクエスト受信
2. 入力バリデーション
3. URLのSSRFチェック
4. profile 保存
5. match_job 作成（`queued`）
6. ワーカーがジョブ取得
7. HTML取得
8. 本文抽出・整形
9. extracted_job_text 保存
10. LLMへリクエスト送信
11. JSONパース
12. result 保存
13. job を `done` へ更新

失敗時:
- 任意の段階でエラー発生時、job を `failed` に更新
- エラーメッセージを保存

---

## 14. SSRF対策要件

URL入力は以下を拒否する:
- `localhost`
- `127.0.0.1`
- `::1`
- `10.0.0.0/8`
- `172.16.0.0/12`
- `192.168.0.0/16`
- link-local
- fileスキームなどHTTP/HTTPS以外

加えて:
- リダイレクト回数制限
- Content-Type がHTML系以外なら拒否
- レスポンスサイズ制限
- タイムアウト設定

---

## 15. フロントエンド要件

## 15.1 画面構成
単一ページ構成でよい。

セクション:
1. ユーザープロフィール入力
2. 求人入力方式選択（text/url）
3. 求人入力欄
4. 実行ボタン
5. 実行状態表示
6. 結果表示
7. コピー操作

## 15.2 状態管理
最低限必要な状態:
- フォーム入力値
- バリデーションエラー
- 実行中フラグ
- job_id
- ジョブ状態
- 判定結果
- APIエラー

## 15.3 UX要件
- 実行中はボタンを無効化
- ポーリング中は状態表示
- エラー時は原因を表示
- コピー成功表示を出す
- URL抽出失敗時はテキスト貼り付けへの誘導文を出す

---

## 16. ディレクトリ構成案

```text
repo/
  frontend/
    src/
      components/
      views/
      api/
      types/
    package.json
    vite.config.ts

  backend/
    cmd/
      app/
        main.go
    internal/
      api/
      config/
      domain/
      repository/
      service/
      worker/
      extractor/
      llm/
      middleware/
      db/
    go.mod

  deploy/
    Dockerfile
    docker-compose.yml

  docs/
    design.md
```

---

## 17. バックエンド内部責務案

### api
- HTTP handler
- request/response 変換

### service
- マッチジョブ作成
- ジョブ状態取得
- 判定処理オーケストレーション

### worker
- `queued` ジョブの取得
- 実行
- `running/done/failed` 更新

### extractor
- URL検証
- HTML取得
- テキスト抽出

### llm
- Ollama API クライアント
- JSON再試行
- モデル指定

### repository
- SQLiteアクセス
- 各テーブルの CRUD

### middleware
- ログ
- CORS
- レート制限
- リクエストサイズ制限

---

## 18. エラーハンドリング方針

想定エラー:
- 入力バリデーションエラー
- URL不正
- SSRF対象URL
- HTML取得失敗
- 本文抽出失敗
- LLMタイムアウト
- LLM応答JSON不正
- DB保存失敗

方針:
- APIでは機械可読なエラーコードを返す
- 画面表示ではわかりやすい文言へ変換する
- ログには詳細を残す

---

## 19. 設定値

環境変数例:
- `APP_PORT`
- `APP_ENV`
- `SQLITE_PATH`
- `OLLAMA_BASE_URL`
- `OLLAMA_MODEL`
- `REQUEST_TIMEOUT_SECONDS`
- `MAX_JOB_TEXT_LENGTH`
- `MAX_HTML_BYTES`
- `JOB_POLL_INTERVAL_MS`
- `RATE_LIMIT_PER_MINUTE`

---

## 20. Docker方針

### server用コンテナ
- frontend をビルドして静的ファイル生成
- backend をビルドして単一バイナリ生成
- 最終イメージに Go バイナリ + static assets を含める

### 前提
- `llm mac` 側は別管理
- server コンテナから `OLLAMA_BASE_URL` で到達可能にする

---

## 21. 実装ステップ

## Phase 1: MVP
- プロフィール入力フォーム
- 求人テキスト入力
- `/api/match` と `/api/match/:id`
- Ollama連携
- 結果表示
- コピー機能

## Phase 2: URL対応
- URL入力UI
- HTML取得
- 本文抽出
- SSRF対策
- エラー表示改善

## Phase 3: 保存と改善
- SQLite保存
- 履歴一覧
- ログ改善
- リトライ改善
- 結果表示の見栄え改善

---

## 22. 受け入れ条件

### MVP受け入れ条件
- テキスト入力で判定できる
- 結果に score / summary / pros / cons / questions_to_ask / clipboard_text が含まれる
- クリップボードコピーが動作する
- LLM出力JSONのパースに失敗したらエラーを返せる

### URL対応受け入れ条件
- URL入力でHTML取得ができる
- 抽出本文をLLMへ渡せる
- 内部向けURLを拒否できる
- URL抽出失敗時に `failed` 状態とエラーメッセージを返せる

---

## 23. Codex向け実装タスク分解

### backend
1. Go APIサーバーの初期化
2. ルーティング実装
3. request/response 型定義
4. SQLite接続実装
5. match_jobs テーブル作成
6. profiles テーブル作成
7. match_results テーブル作成
8. extracted_job_texts テーブル作成
9. `/api/match` 実装
10. `/api/match/:id` 実装
11. ジョブワーカー実装
12. Ollamaクライアント実装
13. JSONパース/再試行実装
14. URLバリデーション実装
15. HTML取得実装
16. テキスト抽出実装
17. SSRF対策実装
18. ログ実装
19. レート制限実装
20. Docker化

### frontend
1. Vue + Vite 初期化
2. プロフィールフォーム実装
3. text/url 切替UI実装
4. 実行ボタン実装
5. APIクライアント実装
6. ポーリング実装
7. 結果表示実装
8. コピー機能実装
9. エラー表示実装
10. 最低限のスタイル調整

---

## 24. 将来拡張案

- 求人比較（複数件を同時に比較）
- プロンプト切り替え
- モデル比較
- プロフィールテンプレート保存
- 結果履歴一覧
- 面接質問の深掘り生成
- 求人本文の自動セクション分割

---

## 25. 保留事項

- ルータに chi / echo のどちらを採用するか
- HTML抽出を goquery のみで行うか、Readability系ロジックを追加するか
- ワーカーを goroutine 常駐で持つか、簡易キュー実装にするか
- 履歴一覧を MVP 範囲に含めるか
- 認証を将来入れるか

---

## 26. 実装判断の原則

- まずは単一ユーザー向けにシンプルに作る
- URL入力時の安定性より、テキスト入力で確実に動くことを優先する
- LLMにはURLではなく本文を渡す
- 出力は必ずJSON化する
- 外部公開範囲は server mac のみに限定する
- セキュリティ上危険なURL取得は必ず防ぐ

