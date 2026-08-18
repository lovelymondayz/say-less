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
