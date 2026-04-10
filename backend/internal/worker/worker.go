package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/uhuko/job-fit-checker/backend/internal/config"
	"github.com/uhuko/job-fit-checker/backend/internal/domain"
	"github.com/uhuko/job-fit-checker/backend/internal/extractor"
	"github.com/uhuko/job-fit-checker/backend/internal/llm"
	"github.com/uhuko/job-fit-checker/backend/internal/repository"
)

type Worker struct {
	repo *repository.Repo
	llm  *llm.Client
	cfg  *config.Config
}

func New(repo *repository.Repo, llmClient *llm.Client, cfg *config.Config) *Worker {
	return &Worker{repo: repo, llm: llmClient, cfg: cfg}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("[worker] started")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[worker] stopped")
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *Worker) poll(ctx context.Context) {
	job, err := w.repo.PickQueuedJob()
	if err != nil {
		if err == sql.ErrNoRows {
			return
		}
		log.Printf("[worker] pick error: %v", err)
		return
	}

	log.Printf("[worker] processing job %s (type=%s)", job.ID, job.JobInputType)
	if err := w.repo.UpdateJobStatus(job.ID, domain.StatusRunning, ""); err != nil {
		log.Printf("[worker] failed to update status to running: %v", err)
		return
	}

	if err := w.process(ctx, job); err != nil {
		log.Printf("[worker] job %s failed: %v", job.ID, err)
		_ = w.repo.UpdateJobStatus(job.ID, domain.StatusFailed, err.Error())
		return
	}

	if err := w.repo.UpdateJobStatus(job.ID, domain.StatusDone, ""); err != nil {
		log.Printf("[worker] failed to update status to done: %v", err)
	}
	log.Printf("[worker] job %s done", job.ID)
}

func (w *Worker) process(ctx context.Context, job *domain.MatchJob) error {
	profile, err := w.repo.GetProfile(job.ProfileID)
	if err != nil {
		return err
	}

	profileInput := domain.ProfileInput{
		PreferredLanguages:  profile.PreferredLanguages,
		AvoidLanguages:      profile.AvoidLanguages,
		Interests:           profile.Interests,
		LowInterests:        profile.LowInterests,
		WorkStyle:           profile.WorkStyle,
		DesiredCompensation: profile.DesiredCompensation,
		Notes:               profile.Notes,
	}

	var jobText string

	switch job.JobInputType {
	case "text":
		jobText = extractor.TruncateText(job.JobInputValue, w.cfg.MaxJobTextLength)
	case "url":
		ext, err := extractor.FetchAndExtract(job.JobInputValue, w.cfg.MaxHTMLBytes, w.cfg.RequestTimeoutSec)
		if err != nil {
			return err
		}
		if err := w.repo.CreateExtractedText(job.ID, job.JobInputValue, ext.Title, ext.MetaDescription, ext.Text); err != nil {
			log.Printf("[worker] failed to save extracted text: %v", err)
		}
		jobText = extractor.TruncateText(ext.Text, w.cfg.MaxJobTextLength)
		if ext.Title != "" {
			jobText = "タイトル: " + ext.Title + "\n\n" + jobText
		}
		if ext.MetaDescription != "" {
			jobText = "概要: " + ext.MetaDescription + "\n\n" + jobText
		}
	}

	if jobText == "" {
		return fmt.Errorf("empty job text")
	}

	out, raw, err := w.llm.Judge(ctx, profileInput, jobText)
	if err != nil {
		return err
	}

	return w.repo.CreateMatchResult(job.ID, *out, w.llm.Model, raw)
}
