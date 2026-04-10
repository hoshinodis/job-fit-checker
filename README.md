# 求人マッチ度チェッカー

ユーザーの志向・スキル・働き方の希望と求人情報をもとに、ローカル LLM（Ollama）でマッチ度を判定し、スコア・理由・懸念点などを表示する Web アプリケーション。

## システム構成

```
[Browser]
  → Cloudflare Tunnel
    → [server mac: Go API + 静的ファイル + SQLite]
      → [llm mac: Ollama API]
```

- **server mac** — Go API サーバー + Vue フロントエンド配信 + SQLite
- **llm mac** — Ollama（LAN 内のみアクセス可）

## 技術スタック

| レイヤー | 技術 |
|---|---|
| バックエンド | Go 1.23, chi, SQLite, goquery |
| フロントエンド | Vue 3, Vite, TypeScript |
| LLM | Ollama |
| デプロイ | Docker, Cloudflare Tunnel |

---

## 起動方法

### 1. Docker で起動（推奨）

```bash
# リポジトリルートで実行
docker compose -f deploy/docker-compose.yml up --build -d
```

アプリは http://localhost:8080 でアクセスできる。

停止:

```bash
docker compose -f deploy/docker-compose.yml down
```

### 2. ローカルで起動（開発用）

#### 前提条件

- Go 1.23+
- Node.js 20+
- Ollama が起動済み（デフォルト: `http://localhost:11434`）

#### フロントエンドビルド

```bash
cd frontend
npm install
npm run build
# ビルド成果物は backend/static/ に出力される
```

#### バックエンド起動

```bash
cd backend
go run ./cmd/app/
```

アプリは http://localhost:8080 でアクセスできる。

#### フロントエンド開発サーバー（ホットリロード）

```bash
cd frontend
npm run dev
```

http://localhost:5173 で開発用サーバーが起動する。API は backend へプロキシされる。

---

## 環境変数

| 変数名 | 説明 | デフォルト値 |
|---|---|---|
| `APP_PORT` | サーバーポート | `8080` |
| `APP_ENV` | 実行環境 | `development` |
| `SQLITE_PATH` | SQLite ファイルパス | `data/job_fit.db` |
| `OLLAMA_BASE_URL` | Ollama API の URL | `http://localhost:11434` |
| `OLLAMA_MODEL` | 使用する LLM モデル | `llama3.1:8b` |
| `REQUEST_TIMEOUT_SECONDS` | リクエストタイムアウト（秒） | `120` |
| `MAX_JOB_TEXT_LENGTH` | 求人テキスト最大文字数 | `20000` |
| `MAX_HTML_BYTES` | HTML 取得最大バイト数 | `5242880` (5MB) |
| `JOB_POLL_INTERVAL_MS` | ポーリング間隔（ミリ秒） | `2000` |
| `RATE_LIMIT_PER_MINUTE` | レート制限（回/分） | `10` |
| `STATIC_DIR` | 静的ファイルディレクトリ | `static` |

---

## API

| メソッド | パス | 説明 |
|---|---|---|
| `POST` | `/api/match` | マッチ判定ジョブを作成 |
| `GET` | `/api/match/:id` | ジョブ状態・結果を取得 |
| `GET` | `/api/health` | ヘルスチェック |

---

## ディレクトリ構成

```
job-fit-checker/
├── backend/
│   ├── cmd/app/          # エントリポイント
│   ├── internal/
│   │   ├── api/          # HTTP ハンドラ
│   │   ├── config/       # 環境変数読み込み
│   │   ├── db/           # SQLite 接続・マイグレーション
│   │   ├── domain/       # ドメインモデル
│   │   ├── extractor/    # URL → テキスト抽出
│   │   ├── llm/          # Ollama クライアント
│   │   ├── middleware/    # CORS, ログ, レート制限等
│   │   ├── repository/   # DB アクセス
│   │   ├── service/      # ビジネスロジック
│   │   └── worker/       # 非同期ジョブワーカー
│   └── static/           # ビルド済みフロントエンド
├── frontend/
│   └── src/
│       ├── api/          # API クライアント
│       ├── components/   # Vue コンポーネント
│       ├── types/        # TypeScript 型定義
│       └── views/        # ページ
├── deploy/
│   ├── Dockerfile
│   └── docker-compose.yml
└── doc/
    └── job_fit_checker_codex_design_spec.md
```

---

## Ollama の準備（llm mac 側）

```bash
# Ollama をインストール後、モデルを取得
ollama pull llama3.1:8b

# LAN 内からアクセスできるように起動
OLLAMA_HOST=0.0.0.0 ollama serve
```

server mac 側の `OLLAMA_BASE_URL` を llm mac の IP に設定する:

```bash
export OLLAMA_BASE_URL=http://<llm-mac-ip>:11434
```

---

## Cloudflare Tunnel での公開

```bash
# cloudflared をインストール後
cloudflared tunnel --url http://localhost:8080
```

詳細は [Cloudflare Tunnel ドキュメント](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) を参照。
