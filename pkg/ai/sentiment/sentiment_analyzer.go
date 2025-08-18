package sentiment

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// SentimentType represents the type of sentiment analysis
type SentimentType int

const (
	SentimentTypeSocial SentimentType = iota
	SentimentTypeNews
	SentimentTypeMarket
	SentimentTypeCombined
	SentimentTypeCustom
)

func (st SentimentType) String() string {
	switch st {
	case SentimentTypeSocial:
		return "Social"
	case SentimentTypeNews:
		return "News"
	case SentimentTypeMarket:
		return "Market"
	case SentimentTypeCombined:
		return "Combined"
	case SentimentTypeCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// SentimentScore represents the sentiment score and confidence
type SentimentScore struct {
	Score      float64 `json:"score"`      // Range: -1.0 (very negative) to 1.0 (very positive)
	Confidence float64 `json:"confidence"` // Range: 0.0 to 1.0
	Magnitude  float64 `json:"magnitude"`  // Range: 0.0 to 1.0 (strength of sentiment)
	Polarity   string  `json:"polarity"`   // "positive", "negative", "neutral"
}

// SentimentSource represents the source of sentiment data
type SentimentSource struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // "twitter", "reddit", "news", "market"
	URL         string    `json:"url"`
	Credibility float64   `json:"credibility"` // Range: 0.0 to 1.0
	LastUpdate  time.Time `json:"last_update"`
}

// SentimentData represents raw sentiment data from a source
type SentimentData struct {
	ID        string                 `json:"id"`
	SourceID  string                 `json:"source_id"`
	Content   string                 `json:"content"`
	Language  string                 `json:"language"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
	RawScore  float64                `json:"raw_score"`
	Processed bool                   `json:"processed"`
	Sentiment *SentimentScore        `json:"sentiment,omitempty"`
}

// SentimentAnalysis represents a complete sentiment analysis result
type SentimentAnalysis struct {
	ID           string                     `json:"id"`
	Asset        string                     `json:"asset"`
	Type         SentimentType              `json:"type"`
	OverallScore *SentimentScore            `json:"overall_score"`
	SourceScores map[string]*SentimentScore `json:"source_scores"`
	Trend        string                     `json:"trend"` // "bullish", "bearish", "neutral"
	Confidence   float64                    `json:"confidence"`
	Volume       int                        `json:"volume"` // Number of data points analyzed
	Timestamp    time.Time                  `json:"timestamp"`
	TimeWindow   time.Duration              `json:"time_window"`
	Metadata     map[string]interface{}     `json:"metadata"`
}

// SentimentMetrics represents aggregated sentiment metrics
type SentimentMetrics struct {
	TotalAnalyses     uint64    `json:"total_analyses"`
	PositiveAnalyses  uint64    `json:"positive_analyses"`
	NegativeAnalyses  uint64    `json:"negative_analyses"`
	NeutralAnalyses   uint64    `json:"neutral_analyses"`
	AverageScore      float64   `json:"average_score"`
	AverageConfidence float64   `json:"average_confidence"`
	LastUpdate        time.Time `json:"last_update"`
}

// SentimentConfig represents configuration for the sentiment analyzer
type SentimentConfig struct {
	MaxSources          uint          `json:"max_sources"`
	MaxDataPoints       uint          `json:"max_data_points"`
	AnalysisInterval    time.Duration `json:"analysis_interval"`
	ConfidenceThreshold float64       `json:"confidence_threshold"`
	LanguageSupport     []string      `json:"language_support"`
	UpdateInterval      time.Duration `json:"update_interval"`
	CleanupInterval     time.Duration `json:"cleanup_interval"`
}

// SentimentAnalyzer represents the main sentiment analysis system
type SentimentAnalyzer struct {
	mu       sync.RWMutex
	Sources  map[string]*SentimentSource   `json:"sources"`
	Data     map[string]*SentimentData     `json:"data"`
	Analyses map[string]*SentimentAnalysis `json:"analyses"`
	Metrics  SentimentMetrics              `json:"metrics"`
	Config   SentimentConfig               `json:"config"`
	running  bool
	stopChan chan struct{}
}

// NewSentimentAnalyzer creates a new sentiment analyzer instance
func NewSentimentAnalyzer(config SentimentConfig) *SentimentAnalyzer {
	if config.MaxSources == 0 {
		config.MaxSources = 100
	}
	if config.MaxDataPoints == 0 {
		config.MaxDataPoints = 10000
	}
	if config.AnalysisInterval == 0 {
		config.AnalysisInterval = time.Minute * 5
	}
	if config.ConfidenceThreshold == 0 {
		config.ConfidenceThreshold = 0.7
	}
	if config.UpdateInterval == 0 {
		config.UpdateInterval = time.Minute * 2
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = time.Hour
	}
	if len(config.LanguageSupport) == 0 {
		config.LanguageSupport = []string{"en", "es", "fr", "de", "zh", "ja"}
	}

	return &SentimentAnalyzer{
		Sources:  make(map[string]*SentimentSource),
		Data:     make(map[string]*SentimentData),
		Analyses: make(map[string]*SentimentAnalysis),
		Config:   config,
		stopChan: make(chan struct{}),
	}
}

// Start starts the sentiment analyzer background processes
func (sa *SentimentAnalyzer) Start() error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if sa.running {
		return fmt.Errorf("sentiment analyzer already running")
	}

	sa.running = true

	// Start background goroutines
	go sa.dataCollectionLoop()
	go sa.analysisLoop()
	go sa.cleanupLoop()

	return nil
}

// Stop stops the sentiment analyzer background processes
func (sa *SentimentAnalyzer) Stop() error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if !sa.running {
		return fmt.Errorf("sentiment analyzer not running")
	}

	sa.running = false
	close(sa.stopChan)

	return nil
}

// AddSource adds a new sentiment data source
func (sa *SentimentAnalyzer) AddSource(source *SentimentSource) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if len(sa.Sources) >= int(sa.Config.MaxSources) {
		return fmt.Errorf("maximum number of sources reached")
	}

	if source.ID == "" {
		source.ID = generateSourceID()
	}

	if source.Credibility < 0 || source.Credibility > 1 {
		return fmt.Errorf("credibility must be between 0 and 1")
	}

	source.LastUpdate = time.Now()
	sa.Sources[source.ID] = sa.copySource(source)

	return nil
}

// AddData adds new sentiment data for analysis
func (sa *SentimentAnalyzer) AddData(data *SentimentData) error {
	sa.mu.Lock()
	defer sa.mu.Unlock()

	if len(sa.Data) >= int(sa.Config.MaxDataPoints) {
		// Remove oldest data point
		var oldestID string
		var oldestTime time.Time
		for id, d := range sa.Data {
			if oldestTime.IsZero() || d.Timestamp.Before(oldestTime) {
				oldestTime = d.Timestamp
				oldestID = id
			}
		}
		if oldestID != "" {
			delete(sa.Data, oldestID)
		}
	}

	if data.ID == "" {
		data.ID = generateDataID()
	}

	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now()
	}

	// Process the sentiment data
	data.Sentiment = sa.analyzeSentiment(data.Content, data.Language)
	data.Processed = true

	sa.Data[data.ID] = sa.copyData(data)

	return nil
}

// AnalyzeSentiment performs sentiment analysis for a specific asset
func (sa *SentimentAnalyzer) AnalyzeSentiment(asset string, sentimentType SentimentType, timeWindow time.Duration) (*SentimentAnalysis, error) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	// Filter data by asset and time window
	var relevantData []*SentimentData
	cutoffTime := time.Now().Add(-timeWindow)

	for _, data := range sa.Data {
		if data.Processed && data.Timestamp.After(cutoffTime) {
			// For now, we'll assume all data is relevant to the asset
			// In a real implementation, we'd filter by asset-specific keywords
			relevantData = append(relevantData, data)
		}
	}

	if len(relevantData) == 0 {
		return nil, fmt.Errorf("no relevant data found for analysis")
	}

	// Perform sentiment analysis
	analysis := sa.performSentimentAnalysis(asset, sentimentType, relevantData, timeWindow)

	// Store the analysis
	sa.Analyses[analysis.ID] = sa.copyAnalysis(analysis)

	// Update metrics
	sa.updateMetrics(analysis)

	return sa.copyAnalysis(analysis), nil
}

// GetSentimentAnalysis retrieves a specific sentiment analysis
func (sa *SentimentAnalyzer) GetSentimentAnalysis(analysisID string) (*SentimentAnalysis, error) {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	analysis, exists := sa.Analyses[analysisID]
	if !exists {
		return nil, fmt.Errorf("sentiment analysis %s not found", analysisID)
	}

	return sa.copyAnalysis(analysis), nil
}

// GetSentimentAnalyses retrieves all sentiment analyses
func (sa *SentimentAnalyzer) GetSentimentAnalyses() map[string]*SentimentAnalysis {
	sa.mu.RLock()
	defer sa.mu.RUnlock()

	analyses := make(map[string]*SentimentAnalysis)
	for id, analysis := range sa.Analyses {
		analyses[id] = sa.copyAnalysis(analysis)
	}
	return analyses
}

// GetMetrics returns the current sentiment metrics
func (sa *SentimentAnalyzer) GetMetrics() SentimentMetrics {
	sa.mu.RLock()
	defer sa.mu.RUnlock()
	return sa.Metrics
}

// performSentimentAnalysis performs the actual sentiment analysis
func (sa *SentimentAnalyzer) performSentimentAnalysis(asset string, sentimentType SentimentType, data []*SentimentData, timeWindow time.Duration) *SentimentAnalysis {
	// Calculate overall sentiment score
	var totalScore, totalConfidence, totalMagnitude float64
	sourceScores := make(map[string]*SentimentScore)

	for _, d := range data {
		if d.Sentiment != nil {
			totalScore += d.Sentiment.Score
			totalConfidence += d.Sentiment.Confidence
			totalMagnitude += d.Sentiment.Magnitude

			// Aggregate by source
			if sourceScore, exists := sourceScores[d.SourceID]; exists {
				sourceScore.Score = (sourceScore.Score + d.Sentiment.Score) / 2
				sourceScore.Confidence = (sourceScore.Confidence + d.Sentiment.Confidence) / 2
				sourceScore.Magnitude = (sourceScore.Magnitude + d.Sentiment.Magnitude) / 2
			} else {
				sourceScores[d.SourceID] = &SentimentScore{
					Score:      d.Sentiment.Score,
					Confidence: d.Sentiment.Confidence,
					Magnitude:  d.Sentiment.Magnitude,
					Polarity:   d.Sentiment.Polarity,
				}
			}
		}
	}

	count := float64(len(data))
	overallScore := &SentimentScore{
		Score:      totalScore / count,
		Confidence: totalConfidence / count,
		Magnitude:  totalMagnitude / count,
		Polarity:   sa.classifyPolarity(totalScore / count),
	}

	// Determine trend
	trend := sa.determineTrend(overallScore.Score, overallScore.Confidence)

	analysis := &SentimentAnalysis{
		ID:           generateAnalysisID(),
		Asset:        asset,
		Type:         sentimentType,
		OverallScore: overallScore,
		SourceScores: sourceScores,
		Trend:        trend,
		Confidence:   overallScore.Confidence,
		Volume:       len(data),
		Timestamp:    time.Now(),
		TimeWindow:   timeWindow,
		Metadata:     make(map[string]interface{}),
	}

	return analysis
}

// analyzeSentiment analyzes the sentiment of a single piece of content
func (sa *SentimentAnalyzer) analyzeSentiment(content, language string) *SentimentScore {
	// Simulate sentiment analysis
	// In a real implementation, this would use NLP libraries or ML models

	// Generate a simulated sentiment score based on content length and random factors
	baseScore := (rand.Float64() - 0.5) * 2 // Range: -1 to 1

	// Adjust based on content length (longer content tends to be more neutral)
	lengthFactor := math.Min(float64(len(content))/1000.0, 1.0)
	adjustedScore := baseScore * (1.0 - lengthFactor*0.3)

	// Ensure score is within bounds
	score := math.Max(-1.0, math.Min(1.0, adjustedScore))

	// Calculate confidence based on content length and language support
	confidence := 0.5 + 0.3*lengthFactor
	if sa.isLanguageSupported(language) {
		confidence += 0.2
	}
	confidence = math.Min(confidence, 1.0)

	// Calculate magnitude (strength of sentiment)
	magnitude := math.Abs(score)

	return &SentimentScore{
		Score:      score,
		Confidence: confidence,
		Magnitude:  magnitude,
		Polarity:   sa.classifyPolarity(score),
	}
}

// classifyPolarity classifies the sentiment polarity
func (sa *SentimentAnalyzer) classifyPolarity(score float64) string {
	if score > 0.1 {
		return "positive"
	} else if score < -0.1 {
		return "negative"
	}
	return "neutral"
}

// determineTrend determines the market trend based on sentiment
func (sa *SentimentAnalyzer) determineTrend(score, confidence float64) string {
	if confidence < sa.Config.ConfidenceThreshold {
		return "neutral"
	}

	if score > 0.3 {
		return "bullish"
	} else if score < -0.3 {
		return "bearish"
	}
	return "neutral"
}

// isLanguageSupported checks if a language is supported
func (sa *SentimentAnalyzer) isLanguageSupported(language string) bool {
	for _, supported := range sa.Config.LanguageSupport {
		if supported == language {
			return true
		}
	}
	return false
}

// updateMetrics updates the sentiment metrics
func (sa *SentimentAnalyzer) updateMetrics(analysis *SentimentAnalysis) {
	sa.Metrics.TotalAnalyses++

	if analysis.OverallScore.Score > 0.1 {
		sa.Metrics.PositiveAnalyses++
	} else if analysis.OverallScore.Score < -0.1 {
		sa.Metrics.NegativeAnalyses++
	} else {
		sa.Metrics.NeutralAnalyses++
	}

	// Update average scores
	total := float64(sa.Metrics.TotalAnalyses)
	sa.Metrics.AverageScore = (sa.Metrics.AverageScore*(total-1) + analysis.OverallScore.Score) / total
	sa.Metrics.AverageConfidence = (sa.Metrics.AverageConfidence*(total-1) + analysis.OverallScore.Confidence) / total

	sa.Metrics.LastUpdate = time.Now()
}

// Background loops
func (sa *SentimentAnalyzer) dataCollectionLoop() {
	ticker := time.NewTicker(sa.Config.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sa.collectData()
		case <-sa.stopChan:
			return
		}
	}
}

func (sa *SentimentAnalyzer) analysisLoop() {
	ticker := time.NewTicker(sa.Config.AnalysisInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sa.performPeriodicAnalysis()
		case <-sa.stopChan:
			return
		}
	}
}

func (sa *SentimentAnalyzer) cleanupLoop() {
	ticker := time.NewTicker(sa.Config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sa.cleanupOldData()
		case <-sa.stopChan:
			return
		}
	}
}

// Placeholder functions for background operations
func (sa *SentimentAnalyzer) collectData() {
	// In a real implementation, this would fetch data from external sources
}

func (sa *SentimentAnalyzer) performPeriodicAnalysis() {
	// In a real implementation, this would perform periodic sentiment analysis
}

func (sa *SentimentAnalyzer) cleanupOldData() {
	// In a real implementation, this would clean up old data based on retention policy
}

// Utility functions for deep copying
func (sa *SentimentAnalyzer) copySource(source *SentimentSource) *SentimentSource {
	if source == nil {
		return nil
	}

	copied := *source
	return &copied
}

func (sa *SentimentAnalyzer) copyData(data *SentimentData) *SentimentData {
	if data == nil {
		return nil
	}

	copied := *data
	copied.Metadata = copyMap(data.Metadata)
	if data.Sentiment != nil {
		copied.Sentiment = &SentimentScore{
			Score:      data.Sentiment.Score,
			Confidence: data.Sentiment.Confidence,
			Magnitude:  data.Sentiment.Magnitude,
			Polarity:   data.Sentiment.Polarity,
		}
	}
	return &copied
}

func (sa *SentimentAnalyzer) copyAnalysis(analysis *SentimentAnalysis) *SentimentAnalysis {
	if analysis == nil {
		return nil
	}

	copied := *analysis
	copied.Metadata = copyMap(analysis.Metadata)

	// Deep copy source scores
	copied.SourceScores = make(map[string]*SentimentScore)
	for k, v := range analysis.SourceScores {
		if v != nil {
			copied.SourceScores[k] = &SentimentScore{
				Score:      v.Score,
				Confidence: v.Confidence,
				Magnitude:  v.Magnitude,
				Polarity:   v.Polarity,
			}
		}
	}

	return &copied
}

// Utility functions
func copyMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}

	copied := make(map[string]interface{})
	for k, v := range m {
		copied[k] = v
	}
	return copied
}

func generateSourceID() string {
	return fmt.Sprintf("source_%d", time.Now().UnixNano())
}

func generateDataID() string {
	return fmt.Sprintf("data_%d", time.Now().UnixNano())
}

func generateAnalysisID() string {
	return fmt.Sprintf("analysis_%d", time.Now().UnixNano())
}
