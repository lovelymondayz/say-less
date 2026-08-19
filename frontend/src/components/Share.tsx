import { useState, useEffect } from 'react'
import { motion } from 'framer-motion'
import { getShare } from '../lib/api'
import { GenerateResult } from '../lib/types'
import TrackCard from './TrackCard'

interface Props {
  shareID: string
  onReset: () => void
}

export default function Share({ shareID, onReset }: Props) {
  const [result, setResult] = useState<GenerateResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [copiedLink, setCopiedLink] = useState(false)

  useEffect(() => {
    const fetchShare = async () => {
      try {
        const data = await getShare(shareID)
        setResult(data)
      } catch (e) {
        setError('Share not found')
      } finally {
        setLoading(false)
      }
    }
    fetchShare()
  }, [shareID])

  const handleCopyLink = () => {
    const url = `${window.location.origin}/s/${shareID}`
    navigator.clipboard.writeText(url)
    setCopiedLink(true)
    setTimeout(() => setCopiedLink(false), 2000)
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    )
  }

  if (error || !result) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center">
        <div className="text-red-400 text-xl mb-4">{error || 'Not found'}</div>
        <button onClick={onReset} className="btn-secondary">← Back to Home</button>
      </div>
    )
  }

  return (
    <div className="min-h-screen py-12 px-4">
      <div className="max-w-3xl mx-auto">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="text-center mb-12">
          <h2 className="text-4xl md:text-5xl font-black text-white mb-4">
            Someone said it. Spotify played it.
          </h2>
        </motion.div>

        {/* Share card / collage */}
        <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.2 }} className="glass p-8 mb-8 relative overflow-hidden">
          {/* Background collage of album art */}
          <div className="absolute inset-0 opacity-20">
            <div className="grid grid-cols-4 gap-1 h-full">
              {result.tracks.slice(0, 8).map((track, i) => (
                <div key={i} className="bg-dark-500 relative overflow-hidden">
                  {track.image_url && (
                    <img src={track.image_url} alt="" className="w-full h-full object-cover blur-sm" />
                  )}
                </div>
              ))}
            </div>
          </div>
          
          <div className="relative z-10 text-center">
            <div className="text-sm text-gray-400 mb-2">I typed:</div>
            <div className="text-2xl md:text-3xl font-bold text-white mb-4">"{result.original_text}"</div>
            <div className="text-sm text-spotify mb-2">🎧 And Spotify said:</div>
            <div className="text-2xl md:text-3xl font-bold text-spotify mb-4">{result.reconstructed}</div>
            
            <div className="flex justify-center gap-6 text-sm text-gray-400">
              <span>{result.tracks.length} tracks</span>
              <span>•</span>
              <span>{Math.round(result.coverage)}% coverage</span>
            </div>
          </div>
        </motion.div>

        {/* Track list */}
        <div className="space-y-3 mb-10">
          {result.tracks.map((track, i) => <TrackCard key={track.id} track={track} index={i} onReplace={() => {}} />)}
        </div>

        {result.caption && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.6 }}
            className="text-center mb-10">
            <p className="text-gray-400 italic text-lg">"{result.caption}"</p>
          </motion.div>
        )}

        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.7 }}
          className="flex flex-col sm:flex-row justify-center gap-4">
          <button onClick={onReset} className="btn-primary">Try It Yourself →</button>
          <button onClick={handleCopyLink} className="btn-secondary">
            {copiedLink ? 'Link Copied!' : '🔗 Copy Link'}
          </button>
        </motion.div>
      </div>
    </div>
  )
}
