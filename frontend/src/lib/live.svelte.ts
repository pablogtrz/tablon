export type LiveState = 'idle' | 'connecting' | 'live' | 'offline' | 'unauthorized'

export const live = $state({
  state: 'idle' as LiveState,
})

export function setLiveState(state: LiveState) {
  live.state = state
}
