import { GenerateResult } from './types'

const API_BASE = import.meta.env.VITE_API_BASE || ''

export async function generatePlaylist(text: string, mode: string): Promise<GenerateResult> {
  const res = await fetch(`${API_BASE}/api/generate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text, mode }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || 'Failed to generate playlist')
  }
  return res.json()
}

export async function saveShare(data: any): Promise<{ share_id: string; url: string }> {
  const res = await fetch(`${API_BASE}/api/share`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  if (!res.ok) throw new Error('Failed to save share')
  return res.json()
}

export async function getShare(id: string): Promise<GenerateResult> {
  const res = await fetch(`${API_BASE}/api/share/${id}`)
  if (!res.ok) throw new Error('Share not found')
  return res.json()
}

export async function spotifyLogin(): Promise<string> {
  const res = await fetch(`${API_BASE}/api/spotify/login`)
  if (!res.ok) throw new Error('Failed to initiate login')
  const data = await res.json()
  return data.url
}

export async function createPlaylist(token: string, name: string, trackIDs: string[]): Promise<any> {
  const res = await fetch(`${API_BASE}/api/spotify/playlist`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token, name, track_ids: trackIDs }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || 'Failed to create playlist')
  }
  return res.json()
}

export async function searchTrack(phrase: string, mode: string): Promise<any> {
  const res = await fetch(`${API_BASE}/api/search-track`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ phrase, mode }),
  })
  if (!res.ok) throw new Error('Failed to search track')
  return res.json()
}

export async function regeneratePlaylist(text: string, mode: string, strategy: string): Promise<GenerateResult> {
  const res = await fetch(`${API_BASE}/api/regenerate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text, mode, strategy }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || 'Failed to regenerate')
  }
  return res.json()
}
