import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { GenerateResult, Track } from '../lib/types'
import TrackCard from './TrackCard'
import { saveShare, spotifyLogin, createPlaylist, regeneratePlaylist } from '../lib/api'

interface Props {
  result: GenerateResult
  onReset: () => void
  onShared: (id: string) => void
}

const REGEN_OPTIONS = [
  { id: 'all', label: '🔄 Regenerate All', desc: 'New interpretation' },
  { id: 'improve', label: '✨ Improve Matches', desc: 'Better title matches' },
  { id: 'popular', label: '🔥 More Popular', desc: 'Well-known songs' },
  { id: 'accurate', label: '🎯 More Accurate', desc: 'Exact matches' },
  { id: 'chaos', label: '🌪️ Make Chaotic', desc: 'Funny & unexpected' },
]

export default function Results({ result, onReset, onShared }: Props) {
  const [saving, setSaving] = useState(false)
  const [copied, setCopied] = useState(false)
  const [copiedCaption, setCopiedCaption] = useState(false)
  const [showSpotify, setShowSpotify] = useState(false)
  const [token, setToken] = useState<string>('')
  const [showRegen, setShowRegen] = useState(false)
  const [regening, setRegening] = useState(false)
  const [currentResult, setCurrentResult] = useState(result)

  const handleShare = async () => {
    setSaving(true)
    try {
      const { share_id } = await saveShare(currentResult)
      onShared(share_id)
    } catch (e) {
      alert('Failed to save share')
    } finally {
      setSaving(false)
    }
  }

  const handleCopy = () => {
    const text = `I typed "${currentResult.original_text}" and Spotify said: ${currentResult.reconstructed}`
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleCopyCaption = () => {
    navigator.clipboard.writeText(currentResult.caption)
    setCopiedCaption(true)
    setTimeout(() => setCopiedCaption(false), 2000)
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
      const trackIDs = currentResult.tracks.map(t => t.id)
      const pl = await createPlaylist(token, 'Say Less — ' + currentResult.original_text, trackIDs)
      window.open(pl.url, '_blank')
    } catch (e: any) {
      alert(e.message)
    }
  }

  const handleRegenerate = async (strategy: string) => {
    setRegening(true)
    setShowRegen(false)
    try {
      const newResult = await regeneratePlaylist(currentResult.original_text, currentResult.mode, strategy)
      setCurrentResult(newResult)
    } catch (e: any) {
      alert(e.message)
    } finally {
      setRegening(false)
    }
  }

  const handleReplaceTrack = (index: number, newTrack: Track) => {
    const newTracks = [...currentResult.tracks]
    newTracks[index] = newTrack
    setCurrentResult({ ...currentResult, tracks: newTracks })
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

        {/* Original text */}
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.2 }}
          className="glass p-6 mb-8 text-center">
          <div className="text-sm text-gray-500 mb-2">You typed:</div>
          <div className="text-2xl md:text-3xl font-bold text-white">"{currentResult.original_text}"</div>
        </motion.div>

        {/* Reconstructed */}
        <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.3 }} className="glass p-6 mb-8 text-center border-spotify/30">
          <div className="text-sm text-spotify mb-2">Your playlist says:</div>
          <div className="text-2xl md:text-3xl font-bold text-spotify">{currentResult.reconstructed}</div>
        </motion.div>

        {/* Stats */}
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.4 }}
          className="flex justify-center gap-8 mb-10">
          <div className="text-center">
            <div className="text-3xl font-black text-white">{currentResult.tracks.length}</div>
            <div className="text-gray-500 text-sm">Tracks</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-black text-white">{Math.round(currentResult.coverage)}%</div>
            <div className="text-gray-500 text-sm">Coverage</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-black text-white capitalize">{currentResult.mode}</div>
            <div className="text-gray-500 text-sm">Mode</div>
          </div>
        </motion.div>

        {/* Track list */}
        <div className="space-y-3 mb-10">
          {currentResult.tracks.map((track, i) => (
            <TrackCard key={`${track.id}-${i}`} track={track} index={i} onReplace={(t) => handleReplaceTrack(i, t)} />
          ))}
        </div>

        {/* Caption */}
        {currentResult.caption && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.6 }}
            className="text-center mb-10">
            <p className="text-gray-400 italic text-lg">"{currentResult.caption}"</p>
            <button onClick={handleCopyCaption} className="mt-2 text-sm text-spotify hover:text-green-400 transition-all">
              {copiedCaption ? 'Copied!' : 'Copy Caption'}
            </button>
          </motion.div>
        )}

        {/* Actions */}
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.7 }}
          className="flex flex-col sm:flex-row justify-center gap-4 mb-6">
          <button onClick={onReset} className="btn-secondary">← Type Another</button>
          <button onClick={handleCopy} className="btn-secondary">
            {copied ? 'Copied!' : 'Copy Result'}
          </button>
          <button onClick={() => setShowRegen(!showRegen)} className="btn-secondary">
            🔄 Regenerate
          </button>
          <button onClick={handleShare} disabled={saving} className="btn-primary">
            {saving ? 'Saving...' : 'Share'}
          </button>
          <button onClick={handleCreatePlaylist} className="btn-primary bg-spotify text-black">
            Add to My Spotify
          </button>
        </motion.div>

        {/* Regenerate options */}
        <AnimatePresence>
          {showRegen && (
            <motion.div
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: 10 }}
              className="glass p-6 mb-6 max-w-md mx-auto"
            >
              <h3 className="text-white font-bold mb-4 text-center">Regenerate Options</h3>
              <div className="space-y-2">
                {REGEN_OPTIONS.map((opt) => (
                  <button
                    key={opt.id}
                    onClick={() => handleRegenerate(opt.id)}
                    disabled={regening}
                    className="w-full px-4 py-3 bg-dark-600/60 hover:bg-dark-500/80 border border-white/10 rounded-xl text-left transition-all disabled:opacity-50"
                  >
                    <div className="font-medium text-white">{opt.label}</div>
                    <div className="text-xs text-gray-400">{opt.desc}</div>
                  </button>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Spotify section */}
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
