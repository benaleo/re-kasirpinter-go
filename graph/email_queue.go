package graph

import (
	"fmt"
	"log"
	"sync"

	"re-kasirpinter-go/graph/model"

	"gorm.io/gorm"
)

// EmailJob represents a single email job in the queue
type EmailJob struct {
	Email   string
	Code    string
	Action  string
	Retry   bool
	IP      *string
	Browser *string
	OS      *string
	DB      *gorm.DB
}

// EmailQueue manages the email sending queue
type EmailQueue struct {
	jobs chan EmailJob
	wg   sync.WaitGroup
}

var (
	emailQueue   *EmailQueue
	queueOnce    sync.Once
	maxQueueSize = 1000
	workerCount  = 5
)

// GetEmailQueue returns the singleton email queue instance
func GetEmailQueue() *EmailQueue {
	queueOnce.Do(func() {
		emailQueue = &EmailQueue{
			jobs: make(chan EmailJob, maxQueueSize),
		}
		emailQueue.startWorkers()
	})
	return emailQueue
}

// startWorkers starts the background workers for processing email jobs
func (eq *EmailQueue) startWorkers() {
	for i := 0; i < workerCount; i++ {
		eq.wg.Add(1)
		go eq.worker(i)
	}
}

// worker processes email jobs from the queue
func (eq *EmailQueue) worker(id int) {
	defer eq.wg.Done()
	log.Printf("Email worker %d started", id)

	for job := range eq.jobs {
		eq.processJob(job)
	}

	log.Printf("Email worker %d stopped", id)
}

// processJob handles a single email job
func (eq *EmailQueue) processJob(job EmailJob) {
	var status, message string
	var err error

	// Determine action type
	action := "create_otp"
	if job.Retry {
		action = "resend_otp"
	}

	// Send email
	err = SendPasswordResetEmail(job.Email, job.Code)
	if err != nil {
		status = "fail"
		message = fmt.Sprintf("error send otp: %v", err)
		log.Printf("Failed to send email to %s: %v", job.Email, err)
	} else {
		status = "success"
		message = "success send otp"
		log.Printf("Successfully sent email to %s", job.Email)
	}

	// Log to database
	logEntry := model.LogEmailDB{
		Email:   job.Email,
		Action:  action,
		Status:  status,
		Message: message,
		IP:      job.IP,
		Browser: job.Browser,
		OS:      job.OS,
	}

	if job.DB != nil {
		if err := job.DB.Create(&logEntry).Error; err != nil {
			log.Printf("Failed to log email to database: %v", err)
		}
	}
}

// Enqueue adds an email job to the queue
func (eq *EmailQueue) Enqueue(job EmailJob) error {
	select {
	case eq.jobs <- job:
		return nil
	default:
		return fmt.Errorf("email queue is full")
	}
}

// Stop gracefully stops the email queue workers
func (eq *EmailQueue) Stop() {
	close(eq.jobs)
	eq.wg.Wait()
}

// EnqueueEmailJob is a convenience function to enqueue an email job
func EnqueueEmailJob(db *gorm.DB, email, code string, retry bool, ip, browser, os *string) error {
	queue := GetEmailQueue()
	job := EmailJob{
		Email:   email,
		Code:    code,
		Retry:   retry,
		IP:      ip,
		Browser: browser,
		OS:      os,
		DB:      db,
	}
	return queue.Enqueue(job)
}
