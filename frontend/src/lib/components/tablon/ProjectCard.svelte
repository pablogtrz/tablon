<script lang="ts">
  import IdChip from './IdChip.svelte'
  import ProjectLogo from './ProjectLogo.svelte'
  import ProjectStatusBadge from './ProjectStatusBadge.svelte'
  import { relativeTime, plural } from '$lib/format'
  import { projectPath, navigate } from '$lib/router.svelte'
  import type { ProjectListItem } from '$lib/types'

  let { project }: { project: ProjectListItem } = $props()

  const href = $derived(projectPath(project.identifier))
</script>

<a
  {href}
  class="glass group flex min-h-[208px] flex-col rounded-2xl p-5 transition-transform duration-[180ms] hover:-translate-y-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background motion-reduce:transition-none motion-reduce:hover:translate-y-0"
  onclick={(event) => {
    event.preventDefault()
    navigate(href)
  }}
>
  <div class="flex items-start gap-3">
    <ProjectLogo
      gradient={project.logo_gradient}
      from={project.logo_gradient_from}
      to={project.logo_gradient_to}
      angle={project.logo_gradient_angle}
      emoji={project.logo_emoji}
      name={project.name}
      size="md"
    />
    <div class="min-w-0 flex-1">
      <div class="flex min-w-0 items-center gap-2">
        <h3 class="truncate text-sm font-semibold tracking-[-.01em]">{project.name}</h3>
        <IdChip identifier={project.identifier} />
      </div>
      <p class="mt-0.5 truncate text-[11px] text-[var(--text-3)]">
        Updated {relativeTime(project.updated_at)}
      </p>
    </div>
    <ProjectStatusBadge status={project.status} />
  </div>

  <p class="mt-4 line-clamp-2 min-h-10 text-xs leading-relaxed text-[var(--text-2)]">
    {project.description || 'No description yet.'}
  </p>

  <div class="mt-auto pt-5">
    <div class="flex items-baseline justify-between gap-3">
      <p class="text-xs text-[var(--text-2)]">
        <span class="tnum text-sm font-semibold text-foreground">{project.done_count}/{project.issue_count}</span>
        <span class="ml-0.5">done</span>
      </p>
      <span class="tnum text-[11px] text-[var(--text-3)]">{plural(project.epic_count, 'epic')}</span>
    </div>
    <progress
      class="tablon-progress mt-3"
      value={project.done_count}
      max={Math.max(project.issue_count, 1)}
      style="--progress-fill: {project.status === 'done' ? 'var(--status-done)' : 'var(--brand)'}"
      aria-label={`${project.done_count} of ${project.issue_count} issues done in ${project.name}`}
    ></progress>
  </div>
</a>
