import { motion } from 'framer-motion'
import { Track } from '../lib/types'

interface Props {
  track: Track
  index: number
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
  semantic: 'Semantic',
  word: 'Word',
}

export default function TrackCard({ track, index }: Props) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.1 }}
      className="glass p-4 flex items-center gap-4 hover:bg-dark-600/60 transition-all group"
    >
      {/* Number */}
      <div className="text-2xl font-black text-gray-600 w-8 text-center">
        {index + 1}
      </div>

      {/* Album art */}
      <div className="relative w-16 h-16 flex-shrink-0">
        {track.image_url ? (
          <img
            src={track.image_url}
            alt={track.album}
            className="w-full h-full object-cover rounded-lg"
          />
        ) : (
          <div className="w-full h-full bg-dark-500 rounded-lg flex items-center justify-center">
            <span className="text-2xl">🎵</span>
          </div>
        )}
      </div>

      {/* Track info */}
      <div className="flex-1 min-w-0">
        <h3 className="text-white font-semibold truncate">{track.title}</h3>
        <p className="text-gray-400 text-sm truncate">{track.artist}</p>
        <p className="text-gray-500 text-xs truncate">{track.album}</p>
      </div>

      {/* Match badge */}
      <div className="flex-shrink-0 text-right">
        <span
          className={`inline-block px-3 py-1 rounded-full text-xs font-medium border ${
            matchColors[track.match_type] || matchColors.word
          }`}
        >
          {matchLabels[track.match_type] || 'Word'}
        </span>
        <div className="text-gray-500 text-xs mt-1">
          {Math.round(track.match_score)}% match
        </div>
      </div>

      {/* Preview */}
      {track.preview_url && (
        <audio
          src={track.preview_url}
          className="hidden group-hover:block w-0"
          controls
        />
      )}
    </motion.div>
  )
}
