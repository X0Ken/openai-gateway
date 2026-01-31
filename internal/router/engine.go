package router

import (
	"errors"
	"math/rand"
	"time"

	"github.com/X0Ken/openai-gateway/pkg/database"
)

// Engine handles intelligent routing with multi-factor scoring
type Engine struct {
	db *database.DB
}

// NewEngine creates a new routing engine
func NewEngine(db *database.DB) *Engine {
	return &Engine{db: db}
}

// RouteResult represents the result of a routing decision
type RouteResult struct {
	Channel          *database.Channel
	BackendModelName string
	SessionID        int64
	IsNew            bool
}

// channelMapping represents a channel with its model mapping details
type channelMapping struct {
	channel          *database.Channel
	backendModelName string
	weight           int
}

// Route selects the best channel for a request
func (e *Engine) Route(userID int64, model string) (*RouteResult, error) {
	// First, check for existing session (sticky routing)
	session, err := e.db.GetSessionByUser(userID)
	if err != nil {
		return nil, err
	}

	if session != nil {
		// Verify the channel still exists and supports the model
		channel, err := e.db.GetChannel(session.ChannelID)
		if err != nil {
			return nil, err
		}

		if channel != nil && channel.Enabled {
			// Get the model object by name to find its ID
			modelObj, err := e.db.GetModelByName(model)
			if err != nil {
				return nil, err
			}
			if modelObj != nil {
				// Check if this channel supports the requested model via model-channel mapping
				modelChannels, err := e.db.GetModelChannelsByChannel(channel.ID)
				if err != nil {
					return nil, err
				}
				for _, mc := range modelChannels {
					if mc.ModelID == modelObj.ID {
						// Update session last used time
						e.db.UpdateSessionLastUsed(session.ID)
						return &RouteResult{
							Channel:          channel,
							BackendModelName: mc.BackendModelName,
							SessionID:        session.ID,
							IsNew:            false,
						}, nil
					}
				}
			}
		}
	}

	// No valid session, find model by name
	modelObj, err := e.db.GetModelByName(model)
	if err != nil {
		return nil, err
	}
	if modelObj == nil {
		return nil, errors.New("model not found: " + model)
	}

	// Get all model-channel mappings for this model
	modelChannels, err := e.db.GetModelChannelsByModel(modelObj.ID)
	if err != nil {
		return nil, err
	}

	if len(modelChannels) == 0 {
		return nil, errors.New("no channels configured for model: " + model)
	}

	// Get channel objects for each mapping
	var mappings []channelMapping
	for _, mc := range modelChannels {
		channel, err := e.db.GetChannel(mc.ChannelID)
		if err != nil {
			return nil, err
		}
		if channel != nil && channel.Enabled {
			mappings = append(mappings, channelMapping{
				channel:          channel,
				backendModelName: mc.BackendModelName,
				weight:           mc.Weight,
			})
		}
	}

	if len(mappings) == 0 {
		return nil, errors.New("no suitable channel found for model: " + model)
	}

	// Score and select best channel using mapping weights
	bestMapping := e.selectBestMapping(mappings)

	// Create new session
	newSession := &database.Session{
		UserID:    userID,
		ChannelID: bestMapping.channel.ID,
	}
	if err := e.db.CreateSession(newSession); err != nil {
		return nil, err
	}

	return &RouteResult{
		Channel:          bestMapping.channel,
		BackendModelName: bestMapping.backendModelName,
		SessionID:        newSession.ID,
		IsNew:            true,
	}, nil
}

// selectBestChannel selects the best channel using weighted scoring
func (e *Engine) selectBestChannel(channels []*database.Channel) *database.Channel {
	if len(channels) == 1 {
		return channels[0]
	}

	// Calculate scores for each channel
	type scoredChannel struct {
		channel *database.Channel
		score   float64
	}

	var scored []scoredChannel
	for _, ch := range channels {
		score := e.calculateScore(ch)
		scored = append(scored, scoredChannel{channel: ch, score: score})
	}

	// Weighted random selection based on scores
	totalScore := 0.0
	for _, sc := range scored {
		totalScore += sc.score
	}

	// Generate random value
	r := rand.Float64() * totalScore

	// Select channel based on weighted probability
	cumulative := 0.0
	for _, sc := range scored {
		cumulative += sc.score
		if r <= cumulative {
			return sc.channel
		}
	}

	// Fallback to last channel
	return scored[len(scored)-1].channel
}

// selectBestMapping selects the best channel mapping using weighted scoring
func (e *Engine) selectBestMapping(mappings []channelMapping) channelMapping {
	if len(mappings) == 1 {
		return mappings[0]
	}

	// Calculate scores for each mapping
	type scoredMapping struct {
		mapping channelMapping
		score   float64
	}

	var scored []scoredMapping
	for _, m := range mappings {
		score := e.calculateScore(m.channel)
		// Factor in mapping weight
		score *= float64(m.weight)
		scored = append(scored, scoredMapping{mapping: m, score: score})
	}

	// Weighted random selection based on scores
	totalScore := 0.0
	for _, sc := range scored {
		totalScore += sc.score
	}

	// Generate random value
	r := rand.Float64() * totalScore

	// Select mapping based on weighted probability
	cumulative := 0.0
	for _, sc := range scored {
		cumulative += sc.score
		if r <= cumulative {
			return sc.mapping
		}
	}

	// Fallback to last mapping
	return scored[len(scored)-1].mapping
}

// calculateScore calculates a composite score for a channel
func (e *Engine) calculateScore(channel *database.Channel) float64 {
	// Base score from weight
	score := float64(channel.Weight)

	// Get metrics for this channel
	metrics, err := e.db.GetChannelMetrics(channel.ID)
	if err != nil || metrics == nil {
		// No metrics yet, use weight only
		return score
	}

	// Factor in latency (lower is better)
	if metrics.LatencyAvg > 0 {
		latencyFactor := 1.0 / (1.0 + metrics.LatencyAvg)
		score *= latencyFactor
	}

	// Factor in error rate (lower is better)
	errorFactor := 1.0 - metrics.ErrorRate
	if errorFactor < 0.1 {
		errorFactor = 0.1 // Minimum factor to avoid zero scores
	}
	score *= errorFactor

	return score
}

// contains checks if a string slice contains a value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// init initializes the random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}
