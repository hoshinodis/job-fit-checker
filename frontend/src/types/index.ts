export interface ProfileInput {
  preferred_languages: string[]
  avoid_languages: string[]
  interests: string[]
  low_interests: string[]
  work_style: string[]
  desired_compensation: string
  notes: string
}

export interface JobInput {
  type: 'text' | 'url'
  value: string
}

export interface MatchRequest {
  profile: ProfileInput
  job_input: JobInput
}

export interface MatchResponse {
  id: string
  status: string
}

export interface MatchResultOutput {
  score: number
  summary: string
  pros: string[]
  cons: string[]
  questions_to_ask: string[]
  clipboard_text: string
  model: string
}

export interface MatchStatusResponse {
  id: string
  status: string
  result?: MatchResultOutput
  error?: string
}
