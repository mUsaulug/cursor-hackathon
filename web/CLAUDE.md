# Web — Next.js

App Router (`web/app/`). Backend at `NEXT_PUBLIC_API_URL` (default `http://localhost:8080`).

## Run

```bash
cd web
npm run dev
# → http://localhost:3000
```

## Rules

- No new npm packages without justification
- TypeScript strict mode; avoid `any`
- Tailwind for styles; no CSS-in-JS libraries
- Keep components in `web/app/`; extract to `web/components/` only when reused 2+
- No secrets or `.env` files in commits

See root `AGENTS.md` for privacy rules and full project context.
