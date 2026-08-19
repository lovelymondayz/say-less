import { motion, AnimatePresence } from 'framer-motion'

interface Props {
  visible: boolean
}

const steps = [
  { text: 'Finding songs that understand you...', icon: '🔍' },
  { text: 'Rearranging your thoughts into music...', icon: '🎵' },
  { text: 'Building your perfect playlist...', icon: '🎧' },
]

export default function ProcessingOverlay({ visible }: Props) {
  return (
    <AnimatePresence>
      {visible && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 bg-dark-900/90 backdrop-blur-sm z-50 flex items-center justify-center"
        >
          <div className="text-center">
            <motion.div
              initial={{ scale: 0.8, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ duration: 0.3 }}
              className="mb-8"
            >
              <div className="w-20 h-20 mx-auto bg-spotify/20 rounded-full flex items-center justify-center mb-6">
                <svg className="w-10 h-10 text-spotify animate-spin" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
              </div>
              <h2 className="text-3xl font-black text-white mb-2">Say Less</h2>
              <p className="text-gray-400 text-sm">Type it. We'll playlist it.</p>
            </motion.div>

            <div className="space-y-3 max-w-sm mx-auto">
              {steps.map((step, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.6 }}
                  className="flex items-center gap-3 text-gray-300"
                >
                  <span className="text-xl">{step.icon}</span>
                  <span className="text-sm">{step.text}</span>
                </motion.div>
              ))}
            </div>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
