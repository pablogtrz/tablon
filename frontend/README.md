# Tablon frontend

Svelte 5 + TypeScript + Vite + Tailwind CSS v4 + shadcn-svelte. Requires Node 22.12+ and pnpm 11.

```bash
nvm use          # Node 22 via .nvmrc
pnpm install
pnpm dev
pnpm check
pnpm test
pnpm test:e2e   # Playwright smoke suite (Chromium, Firefox, WebKit)
pnpm build
```

`pnpm build` writes `dist/`, which is consumed by `go:embed all:frontend/dist`
in the Go server. Build it before producing a release or Docker binary, and
compile the server with `-tags embed`.

The Playwright suite starts a token-enabled Go server and Vite automatically.
Install the browsers once with:

```bash
pnpm exec playwright install chromium firefox webkit
```

The clipboard test is Chromium-only; the rest of the smoke suite runs on all
three browsers.

Theme tokens and base styles live in [`src/app.css`](./src/app.css). Tablon-specific primitives live in [`src/lib/components/tablon`](./src/lib/components/tablon). UI rules and review checks live in [`../docs/design-system.md`](../docs/design-system.md).
