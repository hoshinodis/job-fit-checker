package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/uhuko/job-fit-checker/backend/internal/domain"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func newID() string {
	return "match_" + ulid.Make().String()
}

func profileID() string {
	return "prof_" + ulid.Make().String()
}

func resultID() string {
	return "res_" + ulid.Make().String()
}

func extractedID() string {
	return "ext_" + ulid.Make().String()
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func parseJSON(s string) []string {
	var out []string
	_ = json.Unmarshal([]byte(s), &out)
	if out == nil {
		out = []string{}
	}
	return out
}

// Profile

func (r *Repo) CreateProfile(p domain.ProfileInput) (string, error) {
	id := profileID()
	_, err := r.db.Exec(
		`INSERT INTO profiles (id, preferred_languages_json, avoid_languages_json, interests_json, low_interests_json, work_style_json, desired_compensation, notes) VALUES (?,?,?,?,?,?,?,?)`,
		id,
		mustJSON(p.PreferredLanguages),
		mustJSON(p.AvoidLanguages),
		mustJSON(p.Interests),
		mustJSON(p.LowInterests),
		mustJSON(p.WorkStyle),
		p.DesiredCompensation,
		p.Notes,
	)
	return id, err
}

func (r *Repo) GetProfile(id string) (*domain.Profile, error) {
	row := r.db.QueryRow(`SELECT id, preferred_languages_json, avoid_languages_json, interests_json, low_interests_json, work_style_json, desired_compensation, notes, created_at FROM profiles WHERE id=?`, id)
	var p domain.Profile
	var pl, al, i, li, ws string
	if err := row.Scan(&p.ID, &pl, &al, &i, &li, &ws, &p.DesiredCompensation, &p.Notes, &p.CreatedAt); err != nil {
		return nil, err
	}
	p.PreferredLanguages = parseJSON(pl)
	p.AvoidLanguages = parseJSON(al)
	p.Interests = parseJSON(i)
	p.LowInterests = parseJSON(li)
	p.WorkStyle = parseJSON(ws)
	return &p, nil
}

// MatchJob

func (r *Repo) CreateMatchJob(profileID, inputType, inputValue string) (string, error) {
	id := newID()
	now := time.Now().UTC()
	_, err := r.db.Exec(
		`INSERT INTO match_jobs (id, profile_id, job_input_type, job_input_value, status, created_at, updated_at) VALUES (?,?,?,?,?,?,?)`,
		id, profileID, inputType, inputValue, domain.StatusQueued, now, now,
	)
	return id, err
}

func (r *Repo) GetMatchJob(id string) (*domain.MatchJob, error) {
	row := r.db.QueryRow(`SELECT id, profile_id, job_input_type, job_input_value, status, error_message, created_at, updated_at FROM match_jobs WHERE id=?`, id)
	var j domain.MatchJob
	if err := row.Scan(&j.ID, &j.ProfileID, &j.JobInputType, &j.JobInputValue, &j.Status, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt); err != nil {
		return nil, err
	}
	return &j, nil
}

func (r *Repo) UpdateJobStatus(id, status, errMsg string) error {
	_, err := r.db.Exec(`UPDATE match_jobs SET status=?, error_message=?, updated_at=? WHERE id=?`, status, errMsg, time.Now().UTC(), id)
	return err
}

func (r *Repo) PickQueuedJob() (*domain.MatchJob, error) {
	row := r.db.QueryRow(`SELECT id, profile_id, job_input_type, job_input_value, status, error_message, created_at, updated_at FROM match_jobs WHERE status=? ORDER BY created_at ASC LIMIT 1`, domain.StatusQueued)
	var j domain.MatchJob
	if err := row.Scan(&j.ID, &j.ProfileID, &j.JobInputType, &j.JobInputValue, &j.Status, &j.ErrorMessage, &j.CreatedAt, &j.UpdatedAt); err != nil {
		return nil, err
	}
	return &j, nil
}

// ExtractedJobText

func (r *Repo) CreateExtractedText(matchJobID, sourceURL, title, metaDesc, text string) error {
	id := extractedID()
	_, err := r.db.Exec(
		`INSERT INTO extracted_job_texts (id, match_job_id, source_url, page_title, meta_description, extracted_text) VALUES (?,?,?,?,?,?)`,
		id, matchJobID, sourceURL, title, metaDesc, text,
	)
	return err
}

// MatchResult

func (r *Repo) CreateMatchResult(matchJobID string, out domain.LLMOutput, model, raw string) error {
	id := resultID()
	_, err := r.db.Exec(
		`INSERT INTO match_results (id, match_job_id, score, summary, pros_json, cons_json, questions_to_ask_json, clipboard_text, model_name, raw_response) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		id, matchJobID, out.Score, out.Summary,
		mustJSON(out.Pros), mustJSON(out.Cons), mustJSON(out.QuestionsToAsk),
		out.ClipboardText, model, raw,
	)
	return err
}

func (r *Repo) GetMatchResult(matchJobID string) (*domain.MatchResult, error) {
	row := r.db.QueryRow(`SELECT id, match_job_id, score, summary, pros_json, cons_json, questions_to_ask_json, clipboard_text, model_name, raw_response, created_at FROM match_results WHERE match_job_id=?`, matchJobID)
	var m domain.MatchResult
	var pros, cons, q string
	if err := row.Scan(&m.ID, &m.MatchJobID, &m.Score, &m.Summary, &pros, &cons, &q, &m.ClipboardText, &m.ModelName, &m.RawResponse, &m.CreatedAt); err != nil {
		return nil, err
	}
	m.Pros = parseJSON(pros)
	m.Cons = parseJSON(cons)
	m.QuestionsToAsk = parseJSON(q)
	return &m, nil
}
