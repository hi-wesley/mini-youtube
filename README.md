# mini-youtube

WIP -- NOT A WORKING REPOSITORY

## Quick start (local)

```bash
docker compose up db  # optional local Postgres
export GCS_BUCKET=dev-mini-yt
export GOOGLE_APPLICATION_CREDENTIALS=svc.json
go run ./backend/cmd/server
pnpm --prefix frontend dev
```

## Deploy

gcloud builds submit …
gcloud run deploy …