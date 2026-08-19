import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { generatePlaylist } from '../lib/api'
import { GenerateResult } from '../lib/types'
import ProcessingOverlay from './ProcessingOverlay'

const EXAMPLES = [
  'I MISS YOU',
  'I WANT TO GO HOME',
  'TODAY IS A GOOD DAY',
  'PLEASE GIVE ME MONEY',
  'I NEED A VACATION',
  "I DON'T KNOW WHAT I'M DOING WITH MY LIFE",
]

interface Props {
  onGenerated: (result: GenerateResult) => void
}

export default function Landing({ onGenerated }: Props) {
  const [text, setText] = useState('')
  const [mode, setMode] = useState('smart')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [processing, setProcessing] = useState(false)
  const [placeholderIndex, setPlaceholderIndex] = useState(0)

  useEffect(() => {
    const interval = setInterval(() => {
      setPlaceholderIndex((i) => (i + 1) % EXAMPLES.length)
    }, 3000)
    return () => clearInterval(interval)
  }, [])

  const handleSubmit = async () => {
    if (!text.trim()) return
    setLoading(true)
    setProcessing(true)
    setError('')
    
    // Show processing overlay for at least 1.8s for UX
    const minDelay = new Promise(r => setTimeout(r, 1800))
    
    try {
      const [result] = await Promise.all([
        generatePlaylist(text, mode),
        minDelay,
      ])
      setProcessing(false)
      onGenerated(result)
    } catch (e: any) {
      setProcessing(false)
      setError(e.message || 'Something went wrong')
    } finally {
      setLoading(false)
    }
  }

  const handleExample = (ex: string) => setText(ex)

  return (
    <>
      <ProcessingOverlay visible={processing} />
      <div className="min-h-screen flex flex-col items-center justify-center px-4 relative overflow-hidden">
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-spotify/20 rounded-full blur-3xl animate-pulse-slow" />
          <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/20 rounded-full blur-3xl animate-pulse-slow delay-1000" />
        </div>

        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
          className="relative z-10 text-center max-w-3xl mx-auto"
        >
          <motion.h1
            className="text-6xl md:text-8xl font-black text-white mb-6 tracking-tight"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
          >
            Type it.<br />
            <span className="text-spotify">We'll playlist it.</span>
          </motion.h1>

          <motion.p
            className="text-xl md:text-2xl text-gray-300 mb-12 font-light"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4 }}
          >
            Turn your thoughts into a Spotify playlist,<br />
            one song title at a time.
          </motion.p>

          <motion.div
            className="mb-8"
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.6 }}
          >
            <input
              type="text"
              value={text}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
              placeholder={EXAMPLES[placeholderIndex]}
              className="w-full bg-dark-700/80 backdrop-blur-sm border border-white/20 rounded-2xl px-8 py-6 text-xl md:text-2xl text-white placeholder-gray-500 focus:outline-none input-glow transition-all"
            />
          </motion.div>

          <motion.div
            className="flex flex-wrap justify-center gap-3 mb-10"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.8 }}
          >
            {EXAMPLES.map((ex, i) => (
              <button
                key={i}
                onClick={() => handleExample(ex)}
                className="px-4 py-2 bg-dark-600/60 hover:bg-dark-500/80 border border-white/10 rounded-full text-sm text-gray-300 hover:text-white transition-all"
              >
                {ex}
              </button>
            ))}
          </motion.div>

          <motion.div
            className="flex justify-center gap-4 mb-10"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.9 }}
          >
            {[
              { id: 'exact', label: '🎯 Exact', desc: 'Close matches only' },
              { id: 'smart', label: '🧠 Smart', desc: 'Balanced (default)' },
              { id: 'chaos', label: '🌪️ Chaos', desc: 'Fun & unexpected' },
            ].map((m) => (
              <button
                key={m.id}
                onClick={() => setMode(m.id)}
                className={`px-5 py-3 rounded-xl border transition-all ${
                  mode === m.id
                    ? 'bg-spotify/20 border-spotify text-spotify'
                    : 'bg-dark-700/60 border-white/10 text-gray-400 hover:text-white'
                }`}
              >
                <div className="font-medium">{m.label}</div>
                <div className="text-xs opacity-70">{m.desc}</div>
              </button>
            ))}
          </motion.div>

          <motion.button
            onClick={handleSubmit}
            disabled={loading || !text.trim()}
            className="btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            {loading ? (
              <span className="flex items-center gap-2">
                <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
                Generating...
              </span>
            ) : (
              'Generate Playlist →'
            )}
          </motion.button>

          <AnimatePresence>
            {error && (
              <motion.p
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0 }}
                className="mt-6 text-red-400"
              >
                {error}
              </motion.p>
            )}
          </AnimatePresence>
        </motion.div>
      </div>
    </>
  )
}
