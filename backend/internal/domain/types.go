package domain

import "time"

type Profile struct {
	ID                  string    `json:"id"`
	PreferredLanguages  []string  `json:"preferred_languages"`
	AvoidLanguages      []string  `json:"avoid_languages"`
	Interests           []string  `json:"interests"`
	LowInterests        []string  `json:"low_interests"`
	WorkStyle           []string  `json:"work_style"`
	DesiredCompensation string    `json:"desired_compensation"`
	Notes               string    `json:"notes"`
	CreatedAt           time.Time `json:"created_at"`
}

type MatchJob struct {
	ID            string    `json:"id"`
	ProfileID     string    `json:"profile_id"`
	JobInputType  string    `json:"job_input_type"`
	JobInputValue string    `json:"job_input_value"`
	Status        string    `json:"status"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ExtractedJobText struct {
	ID              string    `json:"id"`
	MatchJobID      string    `json:"match_job_id"`
	SourceURL       string    `json:"source_url"`
	PageTitle       string    `json:"page_title"`
	MetaDescription string    `json:"meta_description"`
	ExtractedText   string    `json:"extracted_text"`
	CreatedAt       time.Time `json:"created_at"`
}

type MatchResult struct {
	ID             string    `json:"id"`
	MatchJobID     string    `json:"match_job_id"`
	Score          int       `json:"score"`
	Summary        string    `json:"summary"`
	Pros           []string  `json:"pros"`
	Cons           []string  `json:"cons"`
	QuestionsToAsk []string  `json:"questions_to_ask"`
	ClipboardText  string    `json:"clipboard_text"`
	ModelName      string    `json:"model_name"`
	RawResponse    string    `json:"raw_response"`
	CreatedAt      time.Time `json:"created_at"`
}

const (
	StatusQueued  = "queued"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"
)

type MatchRequest struct {
	Profile  ProfileInput `json:"profile"`
	JobInput JobInput     `json:"job_input"`
}

type ProfileInput struct {
	PreferredLanguages  []string `json:"preferred_languages"`
	AvoidLanguages      []string `json:"avoid_languages"`
	Interests           []string `json:"interests"`
	LowInterests        []string `json:"low_interests"`
	WorkStyle           []string `json:"work_style"`
	DesiredCompensation string   `json:"desired_compensation"`
	Notes               string   `json:"notes"`
}

type JobInput struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type MatchResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type MatchStatusResponse struct {
	ID     string             `json:"id"`
	Status string             `json:"status"`
	Result *MatchResultOutput `json:"result,omitempty"`
	Error  string             `json:"error,omitempty"`
}

type MatchResultOutput struct {
	Score          int      `json:"score"`
	Summary        string   `json:"summary"`
	Pros           []string `json:"pros"`
	Cons           []string `json:"cons"`
	QuestionsToAsk []string `json:"questions_to_ask"`
	ClipboardText  string   `json:"clipboard_text"`
	Model          string   `json:"model"`
}

type LLMOutput struct {
	Score          int      `json:"score"`
	Summary        string   `json:"summary"`
	Pros           []string `json:"pros"`
	Cons           []string `json:"cons"`
	QuestionsToAsk []string `json:"questions_to_ask"`
	ClipboardText  string   `json:"clipboard_text"`
}
