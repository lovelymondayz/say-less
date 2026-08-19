import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Track } from '../lib/types'
import { searchTrack } from '../lib/api'

interface Props {
  track: Track
  index: number
  onReplace: (track: Track) => void
}

const matchColors: Record<string, string> = {
  exact: 'bg-green-500/20 text-green-400 border-green-500/30',
  phrase: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  partial: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
  semantic: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
  word: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
}

const matchLabels: Record<string, string> = {
  exact: 'Exact',
  phrase: 'Phrase',
  partial: 'Partial',
  semantic: 'Interpretation',
  word: 'Word',
}

export default function TrackCard({ track, index, onReplace }: Props) {
  const [showControls, setShowControls] = useState(false)
  const [replacing, setReplacing] = useState(false)
  const [alternatives, setAlternatives] = useState<Track[]>([])
  const [showAlt, setShowAlt] = useState(false)

  const handleReplace = async () => {
    setReplacing(true)
    setShowAlt(true)
    // Search for alternatives
    try {
      const res = await searchTrack(track.matched_phrase, 'smart')
      if (res.track && res.track.id !== track.id) {
        setAlternatives([res.track])
      } else {
        setAlternatives([])
      }
    } catch (e) {
      setAlternatives([])
    } finally {
      setReplacing(false)
    }
  }

  const handleSelectAlt = (alt: Track) => {
    onReplace(alt)
    setShowAlt(false)
    setAlternatives([])
  }

  return (
    <>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: index * 0.1 }}
        className="glass p-4 flex items-center gap-4 hover:bg-dark-600/60 transition-all group relative"
        onMouseEnter={() => setShowControls(true)}
        onMouseLeave={() => setShowControls(false)}
      >
        <div className="text-2xl font-black text-gray-600 w-8 text-center">
          {index + 1}
        </div>

        <div className="relative w-16 h-16 flex-shrink-0">
          {track.image_url ? (
            <img src={track.image_url} alt={track.album} className="w-full h-full object-cover rounded-lg" />
          ) : (
            <div className="w-full h-full bg-dark-500 rounded-lg flex items-center justify-center">
              <span className="text-2xl">🎵</span>
            </div>
          )}
        </div>

        <div className="flex-1 min-w-0">
          <h3 className="text-white font-semibold truncate">{track.title}</h3>
          <p className="text-gray-400 text-sm truncate">{track.artist}</p>
          <p className="text-gray-500 text-xs truncate">{track.album}</p>
        </div>

        <div className="flex-shrink-0 text-right">
          <span className={`inline-block px-3 py-1 rounded-full text-xs font-medium border ${
            matchColors[track.match_type] || matchColors.word
          }`}>
            {matchLabels[track.match_type] || 'Word'}
          </span>
          <div className="text-gray-500 text-xs mt-1">
            {Math.round(track.match_score)}% match
          </div>
        </div>

        {/* Per-track controls */}
        <AnimatePresence>
          {showControls && (
            <motion.div
              initial={{ opacity: 0, x: 10 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: 10 }}
              className="absolute right-4 top-full mt-2 flex gap-2 z-10"
            >
              <button
                onClick={handleReplace}
                className="px-3 py-1 bg-dark-600 hover:bg-dark-500 border border-white/10 rounded-full text-xs text-gray-300 hover:text-white transition-all"
              >
                {replacing ? '...' : 'Replace'}
              </button>
            </motion.div>
          )}
        </AnimatePresence>
      </motion.div>

      {/* Alternatives dropdown */}
      <AnimatePresence>
        {showAlt && alternatives.length > 0 && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="glass p-3 ml-12 mr-4"
          >
            <p className="text-gray-400 text-xs mb-2">Click to replace:</p>
            {alternatives.map((alt) => (
              <div
                key={alt.id}
                onClick={() => handleSelectAlt(alt)}
                className="flex items-center gap-3 p-2 hover:bg-dark-600/60 rounded-lg cursor-pointer transition-all"
              >
                {alt.image_url && (
                  <img src={alt.image_url} alt={alt.album} className="w-10 h-10 rounded object-cover" />
                )}
                <div className="flex-1 min-w-0">
                  <p className="text-white text-sm truncate">{alt.title}</p>
                  <p className="text-gray-400 text-xs truncate">{alt.artist}</p>
                </div>
                <span className="text-gray-500 text-xs">{Math.round(alt.match_score)}%</span>
              </div>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </>
  )
}
