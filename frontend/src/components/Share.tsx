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

        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: 0.2 }}
          className="glass p-6 mb-8 text-center">
          <div className="text-sm text-gray-500 mb-2">They typed:</div>
          <div className="text-2xl md:text-3xl font-bold text-white">"{result.original_text}"</div>
        </motion.div>

        <motion.div initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.3 }} className="glass p-6 mb-8 text-center border-spotify/30">
          <div className="text-sm text-spotify mb-2">Spotify said:</div>
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
          <button onClick={onReset} className="btn-primary">Try It Yourself →</button>
        </motion.div>
      </div>
    </div>
  )
}
