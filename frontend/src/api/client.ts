import type { MatchRequest, MatchResponse, MatchStatusResponse } from '@/types'

const BASE = '/api'

export async function postMatch(req: MatchRequest): Promise<MatchResponse> {
  const res = await fetch(`${BASE}/match`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(req),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'request failed' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function getMatchStatus(id: string): Promise<MatchStatusResponse> {
  const res = await fetch(`${BASE}/match/${id}`)
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'request failed' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}
