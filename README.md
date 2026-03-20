# vantro-backend

Primary production Go/Fiber API for VANTRO.

## Entrypoints

- API server: `cmd/api/main.go`
- Database migrations: `cmd/migrate/main.go`

## Required Environment

- `DATABASE_URL`
- `JWT_SECRET`

## Common Optional Environment

- `PORT`
- `ENV`
- `API_KEY`
- `ADMIN_KEY`
- `ADMIN_API_KEY`
- `PUBLIC_BASE_URL`
- `RAZORPAY_KEY_ID`
- `RAZORPAY_KEY_SECRET`
- `RAZORPAY_WEBHOOK_SECRET`
- `TWILIO_ACCOUNT_SID`
- `TWILIO_AUTH_TOKEN`
- `TWILIO_WHATSAPP_FROM`

## Commands

- `go run ./cmd/migrate`
- `go run ./cmd/api`
- `go test ./...`
- `go build ./cmd/api`

## Notes

- The clean production frontend lives in `../vantro-ui`
- This folder may contain alternate frontend files, but they are not required to ship the API
- Do not rely on committed `.env` files for production secrets
