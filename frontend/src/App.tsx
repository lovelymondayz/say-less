import { useState } from 'react'
import Landing from './components/Landing'
import Results from './components/Results'
import { GenerateResult } from './lib/types'

type View = 'landing' | 'results'

function App() {
  const [view, setView] = useState<View>('landing')
  const [result, setResult] = useState<GenerateResult | null>(null)

  const handleGenerated = (r: GenerateResult) => {
    setResult(r)
    setView('results')
  }

  const handleReset = () => {
    setResult(null)
    setView('landing')
  }

  return (
    <div className="min-h-screen bg-dark-900">
      {view === 'landing' && <Landing onGenerated={handleGenerated} />}
      {view === 'results' && result && (
        <Results result={result} onReset={handleReset} />
      )}
    </div>
  )
}

export default App
