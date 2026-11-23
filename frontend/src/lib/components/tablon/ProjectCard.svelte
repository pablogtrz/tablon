<script lang="ts">
  import IdChip from './IdChip.svelte'
  import ProjectLogo from './ProjectLogo.svelte'
  import ProjectStatusBadge from './ProjectStatusBadge.svelte'
  import { relativeTime, plural } from '$lib/format'
  import { projectPath, navigate } from '$lib/router.svelte'
  import type { ProjectListItem } from '$lib/types'

  let { project }: { project: ProjectListItem } = $props()

  const progress = $derived(
    project.issue_count === 0 ? 0 : Math.round((project.done_count / project.issue_count) * 100),
  )
  const href = $derived(projectPath(project.identifier))
</script>

<a
  {href}
  class="glass group flex min-h-[188px] flex-col rounded-2xl p-4 transition-transform duration-[180ms] hover:-translate-y-0.5 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background motion-reduce:transition-none motion-reduce:hover:translate-y-0"
  onclick={(event) => {
    event.preventDefault()
    navigate(href)
  }}
>
  <div class="flex items-start gap-3">
    <ProjectLogo gradient={project.logo_gradient} emoji={project.logo_emoji} name={project.name} size="md" />
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

  <p class="mt-3 line-clamp-2 min-h-10 text-xs leading-relaxed text-[var(--text-2)]">
    {project.description || 'No description yet.'}
  </p>

  <div class="mt-auto pt-4">
    <div class="flex items-center justify-between gap-3 text-[11px] text-[var(--text-3)]">
      <span class="tnum">{project.done_count}/{project.issue_count} done</span>
      <span class="tnum">{plural(project.epic_count, 'epic')}</span>
    </div>
    <div class="mt-2 h-1 overflow-hidden rounded-full bg-white/[.06]" aria-hidden="true">
      <div class="h-full rounded-full bg-primary transition-[width] duration-[320ms]" style="width: {progress}%"></div>
    </div>
  </div>
</a>
