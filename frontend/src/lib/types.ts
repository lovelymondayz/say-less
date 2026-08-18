export interface Track {
  id: string
  title: string
  artist: string
  album: string
  image_url: string
  preview_url?: string
  match_type: 'exact' | 'phrase' | 'partial' | 'semantic' | 'word'
  match_score: number
  matched_phrase: string
}

export interface GenerateResult {
  original_text: string
  normalized_text: string
  mode: string
  tracks: Track[]
  reconstructed: string
  coverage: number
  caption: string
}
