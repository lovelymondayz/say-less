import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import Landing from './components/Landing'
import Results from './components/Results'
import Share from './components/Share'
import { GenerateResult } from './lib/types'

type View = 'landing' | 'results' | 'share'

function App() {
  const [view, setView] = useState<View>('landing')
  const [result, setResult] = useState<GenerateResult | null>(null)
  const [shareID, setShareID] = useState<string>('')

  const handleGenerated = (r: GenerateResult) => {
    setResult(r)
    setView('results')
  }

  const handleReset = () => {
    setResult(null)
    setView('landing')
  }

  const handleShared = (id: string) => {
    setShareID(id)
    setView('share')
  }

  return (
    <div className="min-h-screen bg-dark-900">
      {view === 'landing' && <Landing onGenerated={handleGenerated} />}
      {view === 'results' && result && (
        <Results result={result} onReset={handleReset} onShared={handleShared} />
      )}
      {view === 'share' && shareID && (
        <Share shareID={shareID} onReset={handleReset} />
      )}
    </div>
  )
}

export default App
