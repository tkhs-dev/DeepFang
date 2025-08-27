package service

import (
	"context"
	"fmt"
	"golib/env"
	"golib/neuralnet"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// ExperienceChunk represents a chunk of experiences to be sent
type ExperienceChunk struct {
	Experiences []env.Experience `json:"experiences"`
	ChunkID     int64            `json:"chunk_id"`
	Timestamp   time.Time        `json:"timestamp"`
}

// StreamSender interface for gRPC streaming (実際のgRPC実装に応じて調整)
type StreamSender interface {
	Send(*ExperienceChunk) error
}

// ChunkedExperienceBuffer manages experiences and sends them in chunks
type ChunkedExperienceBuffer struct {
	experiences []env.Experience
	mu          sync.RWMutex
	chunkSize   int
	sender      StreamSender
	chunkID     int64
	logger      *logrus.Logger

	// Statistics
	totalExperiences int64
	totalChunksSent  int64
}

func NewChunkedExperienceBuffer(chunkSize int, sender StreamSender, logger *logrus.Logger) *ChunkedExperienceBuffer {
	if logger == nil {
		logger = logrus.New()
	}

	return &ChunkedExperienceBuffer{
		experiences: make([]env.Experience, 0, chunkSize*2), // Buffer for 2 chunks
		chunkSize:   chunkSize,
		sender:      sender,
		logger:      logger,
	}
}

// Add adds an experience and sends chunk if buffer is full
func (ceb *ChunkedExperienceBuffer) Add(exp env.Experience) error {
	ceb.mu.Lock()
	defer ceb.mu.Unlock()

	ceb.experiences = append(ceb.experiences, exp)
	atomic.AddInt64(&ceb.totalExperiences, 1)

	// Send chunk if we have enough experiences
	if len(ceb.experiences) >= ceb.chunkSize {
		return ceb.sendChunk()
	}

	return nil
}

// sendChunk sends a chunk of experiences (must be called with lock held)
func (ceb *ChunkedExperienceBuffer) sendChunk() error {
	if len(ceb.experiences) == 0 {
		return nil
	}

	// Prepare chunk to send
	chunkSize := ceb.chunkSize
	if len(ceb.experiences) < chunkSize {
		chunkSize = len(ceb.experiences)
	}

	chunk := &ExperienceChunk{
		Experiences: make([]env.Experience, chunkSize),
		ChunkID:     atomic.AddInt64(&ceb.chunkID, 1),
		Timestamp:   time.Now(),
	}

	copy(chunk.Experiences, ceb.experiences[:chunkSize])

	// Remove sent experiences from buffer
	ceb.experiences = ceb.experiences[chunkSize:]

	// Send chunk (release lock temporarily for network operation)
	ceb.mu.Unlock()
	err := ceb.sender.Send(chunk)
	ceb.mu.Lock()

	if err != nil {
		ceb.logger.WithError(err).Error("Failed to send experience chunk")
		return err
	}

	atomic.AddInt64(&ceb.totalChunksSent, 1)
	ceb.logger.WithFields(logrus.Fields{
		"chunk_id":     chunk.ChunkID,
		"chunk_size":   len(chunk.Experiences),
		"total_chunks": atomic.LoadInt64(&ceb.totalChunksSent),
	}).Debug("Experience chunk sent successfully")

	return nil
}

// Flush sends remaining experiences in buffer
func (ceb *ChunkedExperienceBuffer) Flush() error {
	ceb.mu.Lock()
	defer ceb.mu.Unlock()

	return ceb.sendChunk()
}

// GetStats returns buffer statistics
func (ceb *ChunkedExperienceBuffer) GetStats() (totalExp int64, totalChunks int64, bufferSize int) {
	ceb.mu.RLock()
	defer ceb.mu.RUnlock()

	return atomic.LoadInt64(&ceb.totalExperiences),
		atomic.LoadInt64(&ceb.totalChunksSent),
		len(ceb.experiences)
}

// StreamingExperienceCollector manages parallel game execution with streaming
type StreamingExperienceCollector struct {
	buffer          *ChunkedExperienceBuffer
	experienceCount int64 // atomic counter
	logger          *logrus.Logger
}

func NewStreamingExperienceCollector(
	chunkSize int,
	sender StreamSender,
	logger *logrus.Logger,
) *StreamingExperienceCollector {

	if logger == nil {
		logger = logrus.New()
	}

	return &StreamingExperienceCollector{
		buffer: NewChunkedExperienceBuffer(chunkSize, sender, logger),
		logger: logger,
	}
}

// CollectExperiencesStreaming collects experiences and streams them in chunks
func (sec *StreamingExperienceCollector) CollectExperiencesStreaming(
	ctx context.Context,
	arg *env.EpisodeArg,
	modelRegistry map[string]neuralnet.Model,
	snakes []string,
	targetExperiences int, // 目安の経験数
	numWorkers int,
	reportInterval time.Duration,
) error {
	atomic.StoreInt64(&sec.experienceCount, 0)

	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()

	// Create worker goroutines
	var wg sync.WaitGroup

	// Error channel for worker errors
	errCh := make(chan error, numWorkers)

	// Timer
	startTime := time.Now()

	// Worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			workerLogger := sec.logger.WithField("worker_id", workerID)

			for {
				select {
				case <-workerCtx.Done():
					workerLogger.Debug("Worker stopping due to worker context cancellation")
					return
				case <-ctx.Done():
					workerLogger.Debug("Worker stopping due to gRPC context cancellation")
					return
				default:
					// Setup clients
					clients := make([]env.Client, len(snakes))
					for j, snakeName := range snakes {
						model, exists := modelRegistry[snakeName]
						if !exists {
							clients[j] = env.NewSimpleSafeClient()
						} else {
							clients[j] = env.NewDeepFang(model, env.GreedyDeepFangPolicy{}, func(exp *env.Experience) {
								atomic.AddInt64(&sec.experienceCount, 1)

								// Add to buffer (this may trigger chunk sending)
								if bufferErr := sec.buffer.Add(*exp); bufferErr != nil {
									workerLogger.WithError(bufferErr).Error("Failed to add experience to buffer")
								}
							})
						}
					}

					// Start a new episode
					err := env.StartEpisode(arg, clients)

					if err != nil {
						if workerCtx.Err() == nil && ctx.Err() == nil { // Only report error if not due to cancellation
							select {
							case errCh <- fmt.Errorf("worker %d error: %w", workerID, err):
							default:
							}
						}
						return
					}
				}
			}
		}(i)
	}

	// Progress monitoring goroutine
	go func() {
		ticker := time.NewTicker(reportInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				count := atomic.LoadInt64(&sec.experienceCount)
				totalExp, totalChunks, bufferSize := sec.buffer.GetStats()

				sec.logger.WithFields(logrus.Fields{
					"current_experiences": count,
					"target_experiences":  targetExperiences,
					"total_experiences":   totalExp,
					"chunks_sent":         totalChunks,
					"buffer_size":         bufferSize,
					"progress_percent":    float64(count) / float64(targetExperiences) * 100,
				}).Debug("Collection progress")

				// 目安に達したらワーカーを停止（ゲームは中断しない）
				if count >= int64(targetExperiences) {
					sec.logger.Debug("Target experience count reached, stopping collection")
					cancelWorkers()
					return
				}
			}
		}
	}()

	// Wait for completion or error
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		cancelWorkers()
		return err
	case <-done:
		// Flush remaining experiences
		if flushErr := sec.buffer.Flush(); flushErr != nil {
			sec.logger.WithError(flushErr).Error("Failed to flush remaining experiences")
		}

		totalExp, totalChunks, _ := sec.buffer.GetStats()
		sec.logger.WithFields(logrus.Fields{
			"total_experiences": totalExp,
			"total_chunks_sent": totalChunks,
			"total_time_msec":   time.Since(startTime).Milliseconds(),
		}).Info("Experience collection completed")

		return nil
	case <-ctx.Done():
		cancelWorkers()
		// Flush remaining experiences on cancellation
		if flushErr := sec.buffer.Flush(); flushErr != nil {
			sec.logger.WithError(flushErr).Error("Failed to flush remaining experiences on cancellation")
		}
		return ctx.Err()
	}
}
