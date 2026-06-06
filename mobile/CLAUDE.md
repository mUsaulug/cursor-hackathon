# Mobile — Expo

Entry: `mobile/index.ts` → `mobile/App.tsx`. Backend at `EXPO_PUBLIC_API_URL` (default `http://localhost:8080`).

## Run

```bash
cd mobile
npx expo start
# Press w for browser, scan QR for physical device
```

## Rules

- Physical device: use LAN IP, not `localhost`
- No new Expo modules without justification
- TypeScript strict mode; avoid `any`
- Do not eject to bare workflow during hackathon
- No secrets or `.env` files in commits

See root `AGENTS.md` for privacy rules and full project context.
