# Tablon frontend

Svelte 5 + TypeScript + Vite + Tailwind CSS v4 + shadcn-svelte. Requires Node 22.12+ and pnpm 11.

```bash
nvm use          # Node 22 via .nvmrc
pnpm install
pnpm dev
pnpm check
pnpm test
pnpm test:e2e   # Playwright smoke suite
pnpm build
```

The Playwright suite starts a token-enabled Go server and Vite automatically. Install the browser once with `pnpm exec playwright install chromium`.

Theme tokens and base styles live in [`src/app.css`](./src/app.css). Tablon-specific primitives live in [`src/lib/components/tablon`](./src/lib/components/tablon). UI rules and review checks live in [`../docs/design-system.md`](../docs/design-system.md).
