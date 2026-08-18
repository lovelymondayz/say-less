import { motion } from 'framer-motion'
import { GenerateResult } from '../lib/types'
import TrackCard from './TrackCard'

interface Props {
  result: GenerateResult
  onReset: () => void
}

export default function Results({ result, onReset }: Props) {
  return (
    <div className="min-h-screen py-12 px-4">
      <div className="max-w-3xl mx-auto">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center mb-12"
        >
          <h2 className="text-4xl md:text-5xl font-black text-white mb-4">
            Your message, but make it music.
          </h2>
          <p className="text-gray-400 text-lg">
            Here's what Spotify heard:
          </p>
        </motion.div>

        {/* Original text */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.2 }}
          className="glass p-6 mb-8 text-center"
        >
          <div className="text-sm text-gray-500 mb-2">You typed:</div>
          <div className="text-2xl md:text-3xl font-bold text-white">
            "{result.original_text}"
          </div>
        </motion.div>

        {/* Reconstructed */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ delay: 0.3 }}
          className="glass p-6 mb-8 text-center border-spotify/30"
        >
          <div className="text-sm text-spotify mb-2">Your playlist says:</div>
          <div className="text-2xl md:text-3xl font-bold text-spotify">
            {result.reconstructed}
          </div>
        </motion.div>

        {/* Stats */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.4 }}
          className="flex justify-center gap-8 mb-10"
        >
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

        {/* Track list */}
        <div className="space-y-3 mb-10">
          {result.tracks.map((track, i) => (
            <TrackCard key={track.id} track={track} index={i} />
          ))}
        </div>

        {/* Caption */}
        {result.caption && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.6 }}
            className="text-center mb-10"
          >
            <p className="text-gray-400 italic text-lg">"{result.caption}"</p>
          </motion.div>
        )}

        {/* Actions */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.7 }}
          className="flex flex-col sm:flex-row justify-center gap-4"
        >
          <button
            onClick={onReset}
            className="btn-secondary"
          >
            ← Type Another
          </button>
          <button
            className="btn-primary"
            onClick={() => {
              const text = `I typed "${result.original_text}" and Spotify said: ${result.reconstructed}`
              navigator.clipboard.writeText(text)
            }}
          >
            Copy Result
          </button>
        </motion.div>
      </div>
    </div>
  )
}
