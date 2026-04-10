package service

import (
	"database/sql"
	"fmt"

	"github.com/uhuko/job-fit-checker/backend/internal/domain"
	"github.com/uhuko/job-fit-checker/backend/internal/repository"
)

type Service struct {
	repo *repository.Repo
}

func New(repo *repository.Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateMatchJob(req domain.MatchRequest) (string, error) {
	if req.JobInput.Type != "text" && req.JobInput.Type != "url" {
		return "", fmt.Errorf("job_input.type must be 'text' or 'url'")
	}
	if req.JobInput.Value == "" {
		return "", fmt.Errorf("job_input.value is required")
	}

	profileID, err := s.repo.CreateProfile(req.Profile)
	if err != nil {
		return "", fmt.Errorf("failed to save profile: %w", err)
	}

	jobID, err := s.repo.CreateMatchJob(profileID, req.JobInput.Type, req.JobInput.Value)
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	return jobID, nil
}

func (s *Service) GetJobStatus(id string) (*domain.MatchStatusResponse, error) {
	job, err := s.repo.GetMatchJob(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, err
	}

	resp := &domain.MatchStatusResponse{
		ID:     job.ID,
		Status: job.Status,
	}

	if job.Status == domain.StatusFailed {
		resp.Error = job.ErrorMessage
	}

	if job.Status == domain.StatusDone {
		result, err := s.repo.GetMatchResult(job.ID)
		if err == nil {
			resp.Result = &domain.MatchResultOutput{
				Score:          result.Score,
				Summary:        result.Summary,
				Pros:           result.Pros,
				Cons:           result.Cons,
				QuestionsToAsk: result.QuestionsToAsk,
				ClipboardText:  result.ClipboardText,
				Model:          result.ModelName,
			}
		}
	}

	return resp, nil
}
