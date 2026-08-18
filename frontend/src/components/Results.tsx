import { useState } from 'react'
import { motion } from 'framer-motion'
import { GenerateResult, Track } from '../lib/types'
import TrackCard from './TrackCard'
import { saveShare, spotifyLogin, createPlaylist } from '../lib/api'

interface Props {
  result: GenerateResult
  onReset: () => void
  onShared: (id: string) => void
}

export default function Results({ result, onReset, onShared }: Props) {
  const [saving, setSaving] = useState(false)
  const [copied, setCopied] = useState(false)
  const [showSpotify, setShowSpotify] = useState(false)
  const [token, setToken] = useState<string>('')

  const handleShare = async () => {
    setSaving(true)
    try {
      const { share_id } = await saveShare(result)
      onShared(share_id)
    } catch (e) {
      alert('Failed to save share')
    } finally {
      setSaving(false)
    }
  }

  const handleCopy = () => {
    const text = `I typed "${result.original_text}" and Spotify said: ${result.reconstructed}`
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleSpotifyLogin = async () => {
    try {
      const url = await spotifyLogin()
      window.location.href = url
    } catch (e) {
      alert('Failed to initiate Spotify login')
    }
  }

  const handleCreatePlaylist = async () => {
    if (!token) {
      setShowSpotify(true)
      return
    }
    try {
      const trackIDs = result.tracks.map(t => t.id)
      const pl = await createPlaylist(token, 'Say Less — ' + result.original_text, trackIDs)
      window.open(pl.url, '_blank')
    } catch (e: any) {
      alert(e.message)
    }
  }

  return (
    <div className="min-h-screen py-12 px-4">
      <div className="max-w-3xl mx-auto">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="text-center mb-12">
          <h2 className="text-4xl md:text-5xl font-black text-white mb-4">
            Your message, but make it music.
          </h2>
          <p className="text-gray-400 text-lg">Here's what Spotify heard:</p>
        </motion.div>

        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.2 }}
          className="glass p-6 mb-8 text-center">
          <div className="text-sm text-gray-500 mb-2">You typed:</div>
          <div className="text-2xl md:text-3xl font-bold text-white">"{result.original_text}"</div>
        </motion.div>

        <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.3 }} className="glass p-6 mb-8 text-center border-spotify/30">
          <div className="text-sm text-spotify mb-2">Your playlist says:</div>
          <div className="text-2xl md:text-3xl font-bold text-spotify">{result.reconstructed}</div>
        </motion.div>

        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.4 }}
          className="flex justify-center gap-8 mb-10">
          <div className="text-center">
            <div className="text-3xl font-black text-white">{result.tracks.length}</div>
            <div className="text-gray-500 text-sm">Tracks</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-black text-white">{Math.round(result.coverage)}%</div>
            <div className="text-gray-500 text-sm">Coverage</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-black text-white capitalize">{result.mode}</div>
            <div className="text-gray-500 text-sm">Mode</div>
          </div>
        </motion.div>

        <div className="space-y-3 mb-10">
          {result.tracks.map((track, i) => <TrackCard key={track.id} track={track} index={i} />)}
        </div>

        {result.caption && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.6 }}
            className="text-center mb-10">
            <p className="text-gray-400 italic text-lg">"{result.caption}"</p>
          </motion.div>
        )}

        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.7 }}
          className="flex flex-col sm:flex-row justify-center gap-4">
          <button onClick={onReset} className="btn-secondary">← Type Another</button>
          <button onClick={handleCopy} className="btn-secondary">
            {copied ? 'Copied!' : 'Copy Result'}
          </button>
          <button onClick={handleShare} disabled={saving} className="btn-primary">
            {saving ? 'Saving...' : 'Share'}
          </button>
          <button onClick={handleCreatePlaylist} className="btn-primary bg-spotify text-black">
            Add to My Spotify
          </button>
        </motion.div>

        {showSpotify && !token && (
          <div className="mt-6 glass p-6 max-w-md mx-auto">
            <h3 className="text-white font-bold mb-4">Connect Spotify</h3>
            <button onClick={handleSpotifyLogin} className="btn-primary w-full">
              Login with Spotify
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
