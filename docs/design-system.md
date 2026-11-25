# Tablon design system

Dark-only v0.1. Svelte 5 + TypeScript + Vite + Tailwind CSS v4 + shadcn-svelte.

Tokens and base styles live in `frontend/src/app.css`. This file explains how to use them. Change both together.

## Where things live

| Path | Purpose |
|---|---|
| `frontend/src/app.css` | Tokens, shadcn bridge, base styles, glass/focus utilities. |
| `frontend/src/lib/components/tablon/` | StatusCell, PriorityBadge, Select, MarkdownRenderer, MarkdownEditor, IssueId, IdChip, AgentChip, IssueTable, BoardCard, IssueBoard, EpicStatusBar, EpicList, CommentsSection, CommentItem, CodeBlockDialog, LiveIndicator. |
| `frontend/src/lib/components/ui/` | shadcn-svelte components, added on demand. |
| `frontend/src/lib/utils.ts` | `cn()` helper. |
| `frontend/src/lib/types.ts` | Status/priority unions, labels and tones. |
| `frontend/src/lib/meta.ts` | Status/priority icons, tones and select options. |
| `frontend/src/lib/issue-grouping.ts` | List grouping by none, epic, status or agent. |
| `frontend/src/lib/preferences.ts` | Typed, best-effort localStorage preferences. |
| `frontend/src/lib/confetti.ts` | One-shot canvas completion burst; reduced motion is a no-op. |
| `frontend/src/lib/markdown.ts` | Markdown render, sanitization, code block affordances and task toggling. |
| `frontend/src/lib/editor-theme.ts` | CodeMirror theme and custom Markdown decorations. |
| `frontend/components.json` | shadcn-svelte config. |

Do not create a parallel style folder.

## Rules

1. Accent is `#FFA443`. Use it only for brand, primary actions, active indices/icons, focus and progress. Never for workflow status.
2. No left color accent bars. Active nav uses a neutral surface + amber index/icon.
3. Glass only on large surfaces: app shell, sheets, project cards, dialogs, toasts. Chips and badges are flat.
4. Issue status is a tinted status cell with icon + label. Priority is a soft tone-tinted icon + label badge. No colored borders, glows or colored-only signals.
5. Canonical issue statuses: Backlog, Todo, In Progress, In Review, Done, Canceled. API slugs: `backlog`, `todo`, `in-progress`, `in-review`, `done`, `canceled`.
6. Epic statuses: Backlog, In Progress, Done. API slugs: `backlog`, `in-progress`, `done`.
7. Project statuses: Planning, Active, Archived, Done. Tones: Planning `--status-backlog`, Active `--status-progress`, Done `--status-done`, Archived `--text-3`.
8. Done is green. Amber is brand/action. Cyan is agent. Red is only Urgent, destructive actions and errors.
9. Fewer elements per screen. No stat strips or activity feeds in the main shell.
10. Dark only. No light theme in v0.1.
11. No React, Radix, dnd-kit, TanStack, date-fns or similar in the Svelte app.
12. Use `Select.svelte` for status/priority controls. It relies on native `appearance: base-select` with a plain native fallback.
13. Issue descriptions render Markdown through `MarkdownRenderer`. External images are https-only; GFM tables and task lists are supported; raw HTML is stripped.
14. Icons use Lucide via `@lucide/svelte`, imported per icon. Primary actions carry icon + label; compact secondary actions are icon-only with an accessible name.
15. Interactive controls use `cursor: pointer`; disabled controls use `cursor: not-allowed`.

## Color

- Canvas: `--bg`, elevated `--bg-2`, surfaces `--surface-1/2/3`, borders `--border` / `--border-strong`, text `--text` / `--text-2` / `--text-3`.
- Brand: `--brand`, `--brand-hover`, `--brand-active`, `--brand-contrast`, `--brand-soft`, `--brand-line`, `--ring`.
- Semantic: `--status-*`, `--status-*-soft`, `--priority-*`, `--priority-*-soft`, `--epic-done-bg`, `--epic-done-border`, `--agent`, `--danger`. Values live in `app.css`.
- `--accent` in shadcn is a hover surface, not the brand. Use `--brand` / `bg-primary`.
- Body text ≥ 4.5:1 on its surface. Amber fills use dark ink.

## Type

- Inter Variable, 14px base, line-height 1.55.
- Three sizes per view maximum. Weight before size.
- Sentence case for titles, labels and buttons. No all-caps labels.
- Mono only for IDs, branches, URLs, CLI and tabular data.
- Lucide outline icons, 1.6px stroke, when a component needs icons.

## Space, shape, glass

- 4px spacing base. Controls 36px; small 30px; large 42px; mobile target 40px.
- Radii: 6 / 8 / 10 / 12 / 16 / 20 / 28 / full. Controls 8-10, cards 12-20, dialogs 16.
- Glass: white 6.5% → 2.2%, 1px border, `blur(20px) saturate(140%)`, inset top highlight.
- Blur only on large surfaces. Cap at two blurred layers per screen; never blur scrolling lists.
- Ambient: near-black base + amber glow top-right + cool slate glow left + grain. Fixed decoration.
- Project logo: curated gradient tile by default; emoji overrides it. The eight pairs live in `app.css` in this order: amber, blue, violet, green, cyan, yellow, slate, indigo. Do not invent per-project pairs.

## Layout

- Desktop sidebar 248px; tablet rail 64px; mobile off-canvas drawer. Content max 1360px.
- Sidebar: brand, search, workspace views, projects, settings. Keep agent/idle detail out of the default sidebar.
- Top bar: breadcrumb, one live/SSE state, one primary action (`New issue`). The shell shows one quiet `LiveIndicator` (connecting, live, offline, token required) while a project stream is active.
- Project header: logo, title, status icon + label, one quiet meta line. Issue lists lead with a split ID (muted project prefix, highlighted number), tinted status cell, priority badge, then title.
- List: flat rows on the page surface, no glass panel. `Group by` supports None, Epic, Status and Agent; group headers are quiet full-width bands with icon, label and count. Persist the choice in localStorage.
- Board: four exact columns `Todo`, `In Progress`, `In Review`, `Done`; counts in headers. Backlog can be shown as an optional fifth column via a labelled checkbox persisted in localStorage; Canceled stays reachable from the card move menu. Cards show split ID, title, priority and one owner/category at most. Done previews the first 10 cards with `See all` / `Show less`. No glass on cards or scrolling columns.
- Epics: 1-column mobile, 2-column desktop. Card shows title, updated time, large tinted status, plain 2-line description excerpt, done/total progress with native `<progress>`, and a stacked status-count bar with a text legend. Done epics use a muted greenish surface. The card has a stretched title link plus Edit/Delete actions; it opens an epic detail view with the full Markdown description and task table.
- Mobile: board snaps horizontally and exposes the move menu; tables stack; drawer nav.

## Components

- Buttons: primary amber fill + dark ink; secondary flat + border; ghost text; danger red tint. One primary action per view. Loading keeps its label. Icon + label for primary actions; icon-only secondary actions need `aria-label`.
- Inputs: 36px, 10px radius, `surface-1`, visible labels, amber focus ring.
- Status/priority: use `StatusCell`, `PriorityBadge`, `ProjectStatusBadge`, `EpicStatusBadge`. Use `Select` for status/priority controls. Do not recreate their CSS inline. Priority badges are 26px with a soft tone background, tone icon and tone text; keep the label.
- Cards: glass only for large items. Project card: logo, name, ID, status icon + label, progress bar, counts, updated time. Epic card: as described under Layout, with a single stretched title link and independent actions above it.
- Progress: use native `<progress class="epic-progress">` with an accessible label. Amber fill is the default; Done epic fills use `--status-done`.
- Markdown: `MarkdownRenderer` for read mode and preview; `MarkdownEditor` for editing. Task checkboxes toggle the source; https images, tables and task lists are allowed; malformed HTML is sanitized.
- Editor: CodeMirror 6 core with custom Markdown decorations. No textarea editing surface. Headings, emphasis, inline code, links, lists and fenced blocks are styled in source while the Markdown markers stay visible. Keep the editor in the lazy IssueView chunk.
- Code blocks: insert through `CodeBlockDialog` with an optional language. Rendered blocks get a language header and Copy action. No language-specific syntax highlighting in v0.1.
- Comments: full-width section on issue detail, chronological, agent chip or `Local`, edited marker, inline Markdown edit, confirm delete, compact composer.
- Board interactions: native HTML5 drag on desktop; every card also has a 40px move menu with all six statuses; `Alt+ArrowLeft/Right` moves across visible columns and restores focus.
- Native dialog: `<dialog>` with `showModal()`, accessible name, `closedby="any"`, `<form method="dialog">`. No manual focus trap or Esc handler. Visible close button. Entry motion uses `@starting-style`, `transition-behavior: allow-discrete` and `overlay`.
- Tables: region + `tabindex="0"` + caption; flat surface with row hover, no card shell; stacked divided rows below 640px.
- Toasts: bottom-right, polite live region, keyboard-focusable close, pause on hover/focus.
- Skeletons: match final layout; shimmer only when reduced motion is not requested.

## Motion

- Durations: 120 / 180 / 240 / 320ms. Standard `cubic-bezier(.2,.8,.2,1)`, enter `cubic-bezier(.16,1,.3,1)`.
- Animate opacity and transform only. No decorative loops or scroll-jacking.
- One exception for completion feedback: moving an issue to Done on the board may fire a one-shot canvas confetti burst (~1.2s, pointer-events none, aria-hidden). It is skipped entirely when reduced motion is requested.
- Route content enters with a 6px/180ms fade-slide; tabs use a 120ms fade; dialogs, drawers and select pickers use top-layer entry motion.
- After route changes, scroll to the top and focus the new page heading/region.
- Reduced motion: remove spatial transforms, shorten fades, keep color transitions. Never apply a global `animation-duration: .01ms` reset.

## Accessibility

- Focus-visible: amber ring, 2px offset, always visible.
- Status and priority: icon or glyph plus label; never color alone. Status cells and priority badges add soft tinted backgrounds.
- 40px mobile targets (24px minimum for secondary dismiss controls); 8px between actions.
- Native dialog handles focus, Esc and return-to-invoker.
- Tables and Markdown tables expose a region, caption/label and keyboard scrolling.
- Custom selects use the native select keyboard contract; unsupported browsers get the native picker.
- SSE updates use a polite live region; do not block the UI when the stream is unavailable. Announce issue/comment changes only while the document is visible.
- CodeMirror content is labelled (`aria-label`) and keyboard operable. Copy feedback uses a polite status, not a toast.

## Svelte 5

- Use `$props()`, `$state()`, `$derived()`, `$effect()`. No stores for local component state.
- `$lib` is aliased in `vite.config.ts` and `tsconfig.app.json`.
- Keep classes static for Tailwind scanning; use `cn()` for variants.
- Custom primitives live in `$lib/components/tablon`; generic UI in `$lib/components/ui`.

## Review checklist

- [ ] Accent is `#FFA443`; no ad-hoc UI colors, gradients or shadows (project logo gradients use the approved palette).
- [ ] No left color accent bar.
- [ ] Glass only on large surfaces; chips and badges flat.
- [ ] Issue status uses a tinted status cell with icon + label; priority uses a soft tinted icon + label badge.
- [ ] Status/priority selects use `Select` and keep visible labels.
- [ ] Issue IDs render as plain split text: muted project prefix, highlighted number.
- [ ] List grouping, the optional Backlog column and the Done preview keep localStorage preferences and remain keyboard accessible.
- [ ] Epic cards show progress and a status-count bar; the detail view renders Markdown and a task table.
- [ ] Markdown descriptions sanitize HTML, allow https images, and support tables and task lists.
- [ ] Route entry/top-layer motion respects reduced motion and focus moves to the new heading.
- [ ] One primary action per view; fewer elements than the last pass.
- [ ] 360, 390 and 768px checked; no horizontal overflow.
- [ ] Tables stack below 640px; dialogs use native `<dialog>`.
- [ ] Board has four default columns, optional Backlog, native drag, a keyboard/menu alternative and no horizontal page overflow. Done previews 10 cards.
- [ ] Comments are chronological, sanitized, editable and deletable with focus restoration.
- [ ] Editing uses CodeMirror, not a textarea; headings and code fences are visually distinct without hiding Markdown markers.
- [ ] Code blocks expose a language header and keyboard-accessible Copy action.
- [ ] Focus, labels, keyboard path and reduced motion checked.
- [ ] UI text uses canonical labels and sentence case.
- [ ] Interactive controls show pointer cursor; disabled controls show not-allowed.

## Changing the system

Change `frontend/src/app.css` and this file together. A new value without a token and a documented rule is a regression.
