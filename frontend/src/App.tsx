import { useState, useEffect, useCallback } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import Landing from './components/Landing'
import Results from './components/Results'
import Share from './components/Share'
import { GenerateResult } from './lib/types'

type View = 'landing' | 'results' | 'share'

// Parse URL for share link
function parseShareID(): string | null {
  const path = window.location.pathname
  const match = path.match(/^\/s\/([a-z0-9]{10})$/)
  return match ? match[1] : null
}

function App() {
  const [view, setView] = useState<View>(() => {
    return parseShareID() ? 'share' : 'landing'
  })
  const [result, setResult] = useState<GenerateResult | null>(null)
  const [shareID, setShareID] = useState<string>(() => parseShareID() || '')

  // Handle browser back/forward
  useEffect(() => {
    const handlePopState = () => {
      const id = parseShareID()
      if (id) {
        setShareID(id)
        setView('share')
      } else {
        setView('landing')
        setResult(null)
        setShareID('')
      }
    }
    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  const handleGenerated = (r: GenerateResult) => {
    setResult(r)
    setView('results')
  }

  const handleReset = () => {
    setResult(null)
    setShareID('')
    setView('landing')
    // Reset URL to root
    window.history.pushState({}, '', '/')
  }

  const handleShared = (id: string) => {
    setShareID(id)
    setView('share')
    // Update URL to share link
    window.history.pushState({}, '', `/s/${id}`)
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
