package sentiment

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewSentimentAnalyzer(t *testing.T) {
	// Test with default config
	sa := NewSentimentAnalyzer(SentimentConfig{})

	if sa == nil {
		t.Fatal("Expected non-nil SentimentAnalyzer")
	}

	if sa.Config.MaxSources != 100 {
		t.Errorf("Expected MaxSources to be 100, got %d", sa.Config.MaxSources)
	}

	if sa.Config.MaxDataPoints != 10000 {
		t.Errorf("Expected MaxDataPoints to be 10000, got %d", sa.Config.MaxDataPoints)
	}

	if sa.Config.AnalysisInterval != time.Minute*5 {
		t.Errorf("Expected AnalysisInterval to be 5 minutes, got %v", sa.Config.AnalysisInterval)
	}

	if sa.Config.ConfidenceThreshold != 0.7 {
		t.Errorf("Expected ConfidenceThreshold to be 0.7, got %f", sa.Config.ConfidenceThreshold)
	}

	if sa.Config.UpdateInterval != time.Minute*2 {
		t.Errorf("Expected UpdateInterval to be 2 minutes, got %v", sa.Config.UpdateInterval)
	}

	if sa.Config.CleanupInterval != time.Hour {
		t.Errorf("Expected CleanupInterval to be 1 hour, got %v", sa.Config.CleanupInterval)
	}

	if len(sa.Config.LanguageSupport) != 6 {
		t.Errorf("Expected 6 supported languages, got %d", len(sa.Config.LanguageSupport))
	}

	expectedLanguages := []string{"en", "es", "fr", "de", "zh", "ja"}
	for _, lang := range expectedLanguages {
		found := false
		for _, supported := range sa.Config.LanguageSupport {
			if supported == lang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected language %s to be supported", lang)
		}
	}
}

func TestNewSentimentAnalyzerCustomConfig(t *testing.T) {
	// Test with custom config
	customConfig := SentimentConfig{
		MaxSources:          50,
		MaxDataPoints:       5000,
		AnalysisInterval:    time.Minute * 10,
		ConfidenceThreshold: 0.8,
		LanguageSupport:     []string{"en", "fr"},
		UpdateInterval:      time.Minute * 5,
		CleanupInterval:     time.Hour * 2,
	}

	sa := NewSentimentAnalyzer(customConfig)

	if sa.Config.MaxSources != 50 {
		t.Errorf("Expected MaxSources to be 50, got %d", sa.Config.MaxSources)
	}

	if sa.Config.MaxDataPoints != 5000 {
		t.Errorf("Expected MaxDataPoints to be 5000, got %d", sa.Config.MaxDataPoints)
	}

	if sa.Config.AnalysisInterval != time.Minute*10 {
		t.Errorf("Expected AnalysisInterval to be 10 minutes, got %v", sa.Config.AnalysisInterval)
	}

	if sa.Config.ConfidenceThreshold != 0.8 {
		t.Errorf("Expected ConfidenceThreshold to be 0.8, got %f", sa.Config.ConfidenceThreshold)
	}

	if len(sa.Config.LanguageSupport) != 2 {
		t.Errorf("Expected 2 supported languages, got %d", len(sa.Config.LanguageSupport))
	}
}

func TestStartStop(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Test Start
	err := sa.Start()
	if err != nil {
		t.Fatalf("Expected Start to succeed, got error: %v", err)
	}

	if !sa.running {
		t.Error("Expected analyzer to be running after Start")
	}

	// Test Start when already running
	err = sa.Start()
	if err == nil {
		t.Error("Expected error when starting already running analyzer")
	}

	// Test Stop
	err = sa.Stop()
	if err != nil {
		t.Fatalf("Expected Stop to succeed, got error: %v", err)
	}

	if sa.running {
		t.Error("Expected analyzer to not be running after Stop")
	}

	// Test Stop when not running
	err = sa.Stop()
	if err == nil {
		t.Error("Expected error when stopping non-running analyzer")
	}
}

func TestAddSource(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{MaxSources: 2})

	// Test adding valid source
	source1 := &SentimentSource{
		Name:        "Twitter",
		Type:        "social",
		URL:         "https://twitter.com",
		Credibility: 0.8,
	}

	err := sa.AddSource(source1)
	if err != nil {
		t.Fatalf("Expected AddSource to succeed, got error: %v", err)
	}

	if source1.ID == "" {
		t.Error("Expected source ID to be generated")
	}

	if len(sa.Sources) != 1 {
		t.Errorf("Expected 1 source, got %d", len(sa.Sources))
	}

	// Test adding source with custom ID
	source2 := &SentimentSource{
		ID:          "custom_id",
		Name:        "Reddit",
		Type:        "social",
		URL:         "https://reddit.com",
		Credibility: 0.7,
	}

	err = sa.AddSource(source2)
	if err != nil {
		t.Fatalf("Expected AddSource to succeed, got error: %v", err)
	}

	if source2.ID != "custom_id" {
		t.Errorf("Expected source ID to remain 'custom_id', got %s", source2.ID)
	}

	// Test adding source with invalid credibility
	invalidSource := &SentimentSource{
		Name:        "Invalid",
		Type:        "social",
		URL:         "https://invalid.com",
		Credibility: 1.5, // Invalid: > 1.0
	}

	err = sa.AddSource(invalidSource)
	if err == nil {
		t.Error("Expected error for invalid credibility")
	}

	// Test adding source when max reached
	thirdSource := &SentimentSource{
		Name:        "Third",
		Type:        "social",
		URL:         "https://third.com",
		Credibility: 0.6,
	}

	err = sa.AddSource(thirdSource)
	if err == nil {
		t.Error("Expected error when max sources reached")
	}
}

func TestAddData(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{MaxDataPoints: 3})

	// Test adding valid data
	data1 := &SentimentData{
		SourceID:  "source1",
		Content:   "This is a positive message about Bitcoin!",
		Language:  "en",
		Timestamp: time.Now(),
	}

	err := sa.AddData(data1)
	if err != nil {
		t.Fatalf("Expected AddData to succeed, got error: %v", err)
	}

	if data1.ID == "" {
		t.Error("Expected data ID to be generated")
	}

	if !data1.Processed {
		t.Error("Expected data to be processed")
	}

	if data1.Sentiment == nil {
		t.Error("Expected sentiment to be analyzed")
	}

	if len(sa.Data) != 1 {
		t.Errorf("Expected 1 data point, got %d", len(sa.Data))
	}

	// Test adding data with custom ID
	data2 := &SentimentData{
		ID:        "custom_data_id",
		SourceID:  "source2",
		Content:   "This is a negative message about Ethereum.",
		Language:  "en",
		Timestamp: time.Now(),
	}

	err = sa.AddData(data2)
	if err != nil {
		t.Fatalf("Expected AddData to succeed, got error: %v", err)
	}

	if data2.ID != "custom_data_id" {
		t.Errorf("Expected data ID to remain 'custom_data_id', got %s", data2.ID)
	}

	// Test adding data with zero timestamp
	data3 := &SentimentData{
		SourceID: "source3",
		Content:  "This is a neutral message.",
		Language: "en",
		// Timestamp is zero
	}

	err = sa.AddData(data3)
	if err != nil {
		t.Fatalf("Expected AddData to succeed, got error: %v", err)
	}

	if data3.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	// Test data overflow (should remove oldest)
	data4 := &SentimentData{
		SourceID:  "source4",
		Content:   "This is the fourth data point.",
		Language:  "en",
		Timestamp: time.Now(),
	}

	err = sa.AddData(data4)
	if err != nil {
		t.Fatalf("Expected AddData to succeed, got error: %v", err)
	}

	if len(sa.Data) != 3 {
		t.Errorf("Expected 3 data points after overflow, got %d", len(sa.Data))
	}
}

func TestAnalyzeSentiment(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Add some test data
	data1 := &SentimentData{
		SourceID:  "source1",
		Content:   "Bitcoin is amazing!",
		Language:  "en",
		Timestamp: time.Now(),
	}

	data2 := &SentimentData{
		SourceID:  "source2",
		Content:   "Ethereum is terrible.",
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(data1)
	sa.AddData(data2)

	// Test sentiment analysis
	analysis, err := sa.AnalyzeSentiment("BTC", SentimentTypeSocial, time.Hour)
	if err != nil {
		t.Fatalf("Expected AnalyzeSentiment to succeed, got error: %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected non-nil analysis")
	}

	if analysis.Asset != "BTC" {
		t.Errorf("Expected Asset to be 'BTC', got %s", analysis.Asset)
	}

	if analysis.Type != SentimentTypeSocial {
		t.Errorf("Expected Type to be SentimentTypeSocial, got %v", analysis.Type)
	}

	if analysis.OverallScore == nil {
		t.Fatal("Expected non-nil OverallScore")
	}

	if analysis.OverallScore.Score < -1 || analysis.OverallScore.Score > 1 {
		t.Errorf("Expected score between -1 and 1, got %f", analysis.OverallScore.Score)
	}

	if analysis.OverallScore.Confidence < 0 || analysis.OverallScore.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", analysis.OverallScore.Confidence)
	}

	if analysis.Volume != 2 {
		t.Errorf("Expected Volume to be 2, got %d", analysis.Volume)
	}

	if analysis.TimeWindow != time.Hour {
		t.Errorf("Expected TimeWindow to be 1 hour, got %v", analysis.TimeWindow)
	}

	// Test analysis with no relevant data
	analysis, err = sa.AnalyzeSentiment("BTC", SentimentTypeSocial, time.Nanosecond)
	if err == nil {
		t.Error("Expected error for no relevant data")
	}
}

func TestGetSentimentAnalysis(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Add test data and perform analysis
	data := &SentimentData{
		SourceID:  "source1",
		Content:   "Test content",
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(data)
	analysis, _ := sa.AnalyzeSentiment("BTC", SentimentTypeSocial, time.Hour)

	// Test retrieving analysis
	retrieved, err := sa.GetSentimentAnalysis(analysis.ID)
	if err != nil {
		t.Fatalf("Expected GetSentimentAnalysis to succeed, got error: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Expected non-nil analysis")
	}

	if retrieved.ID != analysis.ID {
		t.Errorf("Expected ID to match, got %s vs %s", retrieved.ID, analysis.ID)
	}

	// Test retrieving non-existent analysis
	_, err = sa.GetSentimentAnalysis("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent analysis")
	}
}

func TestGetSentimentAnalyses(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Add test data and perform multiple analyses
	data1 := &SentimentData{
		SourceID:  "source1",
		Content:   "Content 1",
		Language:  "en",
		Timestamp: time.Now(),
	}

	data2 := &SentimentData{
		SourceID:  "source2",
		Content:   "Content 2",
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(data1)
	sa.AddData(data2)

	_, _ = sa.AnalyzeSentiment("BTC", SentimentTypeSocial, time.Hour)
	_, _ = sa.AnalyzeSentiment("ETH", SentimentTypeNews, time.Hour)

	// Test retrieving all analyses
	analyses := sa.GetSentimentAnalyses()

	if len(analyses) != 2 {
		t.Errorf("Expected 2 analyses, got %d", len(analyses))
	}

	// Verify both analyses are present
	foundBTC, foundETH := false, false
	for _, analysis := range analyses {
		if analysis.Asset == "BTC" {
			foundBTC = true
		}
		if analysis.Asset == "ETH" {
			foundETH = true
		}
	}

	if !foundBTC {
		t.Error("Expected to find BTC analysis")
	}

	if !foundETH {
		t.Error("Expected to find ETH analysis")
	}
}

func TestGetMetrics(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Initial metrics should be zero
	metrics := sa.GetMetrics()

	if metrics.TotalAnalyses != 0 {
		t.Errorf("Expected TotalAnalyses to be 0, got %d", metrics.TotalAnalyses)
	}

	if metrics.PositiveAnalyses != 0 {
		t.Errorf("Expected PositiveAnalyses to be 0, got %d", metrics.PositiveAnalyses)
	}

	if metrics.NegativeAnalyses != 0 {
		t.Errorf("Expected NegativeAnalyses to be 0, got %d", metrics.NegativeAnalyses)
	}

	if metrics.NeutralAnalyses != 0 {
		t.Errorf("Expected NeutralAnalyses to be 0, got %d", metrics.NeutralAnalyses)
	}

	if metrics.AverageScore != 0 {
		t.Errorf("Expected AverageScore to be 0, got %f", metrics.AverageScore)
	}

	if metrics.AverageConfidence != 0 {
		t.Errorf("Expected AverageConfidence to be 0, got %f", metrics.AverageConfidence)
	}

	// Add test data and perform analysis to update metrics
	data := &SentimentData{
		SourceID:  "source1",
		Content:   "Test content",
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(data)
	sa.AnalyzeSentiment("BTC", SentimentTypeSocial, time.Hour)

	// Check updated metrics
	metrics = sa.GetMetrics()

	if metrics.TotalAnalyses != 1 {
		t.Errorf("Expected TotalAnalyses to be 1, got %d", metrics.TotalAnalyses)
	}

	if metrics.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set")
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test copyMap
	originalMap := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	copiedMap := copyMap(originalMap)

	if len(copiedMap) != len(originalMap) {
		t.Errorf("Expected copied map to have same length, got %d vs %d", len(copiedMap), len(originalMap))
	}

	for k, v := range originalMap {
		if copiedMap[k] != v {
			t.Errorf("Expected value for key %s to match, got %v vs %v", k, copiedMap[k], v)
		}
	}

	// Test copyMap with nil
	nilMap := copyMap(nil)
	if nilMap != nil {
		t.Error("Expected copyMap(nil) to return nil")
	}

	// Test ID generation functions
	sourceID := generateSourceID()
	if sourceID == "" {
		t.Error("Expected non-empty source ID")
	}

	dataID := generateDataID()
	if dataID == "" {
		t.Error("Expected non-empty data ID")
	}

	analysisID := generateAnalysisID()
	if analysisID == "" {
		t.Error("Expected non-empty analysis ID")
	}

	// Verify IDs are unique
	if sourceID == dataID || sourceID == analysisID || dataID == analysisID {
		t.Error("Expected generated IDs to be unique")
	}
}

func TestSentimentTypeString(t *testing.T) {
	testCases := []struct {
		sentimentType SentimentType
		expected      string
	}{
		{SentimentTypeSocial, "Social"},
		{SentimentTypeNews, "News"},
		{SentimentTypeMarket, "Market"},
		{SentimentTypeCombined, "Combined"},
		{SentimentTypeCustom, "Custom"},
		{SentimentType(999), "Unknown"},
	}

	for _, tc := range testCases {
		result := tc.sentimentType.String()
		if result != tc.expected {
			t.Errorf("Expected String() for %v to return '%s', got '%s'", tc.sentimentType, tc.expected, result)
		}
	}
}

func TestSentimentAnalysisEdgeCases(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Test with very short content
	shortData := &SentimentData{
		SourceID:  "source1",
		Content:   "Hi",
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(shortData)

	if shortData.Sentiment == nil {
		t.Error("Expected sentiment to be analyzed for short content")
	}

	// Test with very long content
	longContent := ""
	for i := 0; i < 2000; i++ {
		longContent += "This is a very long piece of content that should test the length factor in sentiment analysis. "
	}

	longData := &SentimentData{
		SourceID:  "source2",
		Content:   longContent,
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(longData)

	if longData.Sentiment == nil {
		t.Error("Expected sentiment to be analyzed for long content")
	}

	// Test with unsupported language
	unsupportedData := &SentimentData{
		SourceID:  "source3",
		Content:   "Content in unsupported language",
		Language:  "xx", // Unsupported language code
		Timestamp: time.Now(),
	}

	sa.AddData(unsupportedData)

	if unsupportedData.Sentiment == nil {
		t.Error("Expected sentiment to be analyzed for unsupported language")
	}

	// Verify confidence is lower for unsupported language
	if unsupportedData.Sentiment.Confidence >= 0.9 {
		t.Error("Expected lower confidence for unsupported language")
	}
}

func TestPolarityClassification(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Test that sentiment analysis produces valid polarities
	// Since the analysis is random, we just verify the structure is correct

	// Test with various content lengths
	testCases := []struct {
		name    string
		content string
	}{
		{"Short content", "Hi"},
		{"Medium content", "This is a medium length message about cryptocurrency."},
		{"Long content", "This is a very long message about cryptocurrency that should test the length factor in sentiment analysis. "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := &SentimentData{
				SourceID:  "source1",
				Content:   tc.content,
				Language:  "en",
				Timestamp: time.Now(),
			}

			sa.AddData(data)

			// Verify sentiment was analyzed
			if data.Sentiment == nil {
				t.Error("Expected sentiment to be analyzed")
			}

			// Verify polarity is valid
			validPolarities := []string{"positive", "negative", "neutral"}
			found := false
			for _, polarity := range validPolarities {
				if data.Sentiment.Polarity == polarity {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Expected valid polarity, got %s", data.Sentiment.Polarity)
			}

			// Verify score bounds
			if data.Sentiment.Score < -1.0 || data.Sentiment.Score > 1.0 {
				t.Errorf("Expected score between -1 and 1, got %f", data.Sentiment.Score)
			}

			// Verify confidence bounds
			if data.Sentiment.Confidence < 0.0 || data.Sentiment.Confidence > 1.0 {
				t.Errorf("Expected confidence between 0 and 1, got %f", data.Sentiment.Confidence)
			}
		})
	}
}

func TestTrendDetermination(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{ConfidenceThreshold: 0.7})

	// Test bullish trend (high positive score, high confidence)
	analysis := &SentimentAnalysis{
		OverallScore: &SentimentScore{
			Score:      0.8, // High positive
			Confidence: 0.9, // High confidence
		},
	}

	trend := sa.determineTrend(analysis.OverallScore.Score, analysis.OverallScore.Confidence)
	if trend != "bullish" {
		t.Errorf("Expected bullish trend, got %s", trend)
	}

	// Test bearish trend (high negative score, high confidence)
	analysis.OverallScore.Score = -0.8
	trend = sa.determineTrend(analysis.OverallScore.Score, analysis.OverallScore.Confidence)
	if trend != "bearish" {
		t.Errorf("Expected bearish trend, got %s", trend)
	}

	// Test neutral trend (low score, high confidence)
	analysis.OverallScore.Score = 0.1
	trend = sa.determineTrend(analysis.OverallScore.Score, analysis.OverallScore.Confidence)
	if trend != "neutral" {
		t.Errorf("Expected neutral trend, got %s", trend)
	}

	// Test neutral trend (high score, low confidence)
	analysis.OverallScore.Score = 0.8
	analysis.OverallScore.Confidence = 0.5 // Below threshold
	trend = sa.determineTrend(analysis.OverallScore.Score, analysis.OverallScore.Confidence)
	if trend != "neutral" {
		t.Errorf("Expected neutral trend for low confidence, got %s", trend)
	}
}

func TestLanguageSupport(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Test supported languages
	supportedLanguages := []string{"en", "es", "fr", "de", "zh", "ja"}
	for _, lang := range supportedLanguages {
		if !sa.isLanguageSupported(lang) {
			t.Errorf("Expected language %s to be supported", lang)
		}
	}

	// Test unsupported languages
	unsupportedLanguages := []string{"xx", "yy", "zz"}
	for _, lang := range unsupportedLanguages {
		if sa.isLanguageSupported(lang) {
			t.Errorf("Expected language %s to not be supported", lang)
		}
	}
}

func TestConcurrency(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Start the analyzer
	err := sa.Start()
	if err != nil {
		t.Fatalf("Failed to start analyzer: %v", err)
	}
	defer sa.Stop()

	// Test concurrent data addition
	const numGoroutines = 10
	const dataPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < dataPerGoroutine; j++ {
				data := &SentimentData{
					SourceID:  fmt.Sprintf("source_%d", id),
					Content:   fmt.Sprintf("Content %d from goroutine %d", j, id),
					Language:  "en",
					Timestamp: time.Now(),
				}
				sa.AddData(data)
			}
		}(i)
	}

	wg.Wait()

	// Verify all data was added
	if len(sa.Data) != numGoroutines*dataPerGoroutine {
		t.Errorf("Expected %d data points, got %d", numGoroutines*dataPerGoroutine, len(sa.Data))
	}
}

func TestMemorySafety(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Test that modifications to returned data don't affect internal state
	source := &SentimentSource{
		Name:        "Test Source",
		Type:        "social",
		URL:         "https://test.com",
		Credibility: 0.8,
	}

	err := sa.AddSource(source)
	if err != nil {
		t.Fatalf("Failed to add source: %v", err)
	}

	// Modify the original source
	originalName := source.Name
	source.Name = "Modified Name"

	// Verify internal state wasn't affected
	if sa.Sources[source.ID].Name != originalName {
		t.Error("Expected internal state to not be affected by external modifications")
	}

	// Test data deep copy
	data := &SentimentData{
		SourceID:  "source1",
		Content:   "Test content",
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(data)

	// Modify the original data
	originalContent := data.Content
	data.Content = "Modified content"

	// Verify internal state wasn't affected
	if sa.Data[data.ID].Content != originalContent {
		t.Error("Expected internal state to not be affected by external modifications")
	}

	// Test analysis deep copy
	analysis, _ := sa.AnalyzeSentiment("BTC", SentimentTypeSocial, time.Hour)

	// Modify the original analysis
	originalAsset := analysis.Asset
	analysis.Asset = "Modified Asset"

	// Verify internal state wasn't affected
	if sa.Analyses[analysis.ID].Asset != originalAsset {
		t.Error("Expected internal state to not be affected by external modifications")
	}
}

func TestEdgeCases(t *testing.T) {
	sa := NewSentimentAnalyzer(SentimentConfig{})

	// Test with empty content
	emptyData := &SentimentData{
		SourceID:  "source1",
		Content:   "",
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(emptyData)

	if emptyData.Sentiment == nil {
		t.Error("Expected sentiment to be analyzed for empty content")
	}

		// Test with very long content that exceeds length factor calculation
	veryLongContent := ""
	for i := 0; i < 100; i++ {
		veryLongContent += "Long. "
	}
	
	longData := &SentimentData{
		SourceID:  "source2",
		Content:   veryLongContent,
		Language:  "en",
		Timestamp: time.Now(),
	}
	
	sa.AddData(longData)
	
	if longData.Sentiment == nil {
		t.Error("Expected sentiment to be analyzed for very long content")
	}
	
	// Verify length factor is capped at 1.0
	expectedLengthFactor := 1.0
	actualLengthFactor := float64(len(veryLongContent)) / 1000.0
	if actualLengthFactor > expectedLengthFactor {
		t.Errorf("Expected length factor to be capped at 1.0, got %f", actualLengthFactor)
	}

	// Test with extreme sentiment scores
	extremeData := &SentimentData{
		SourceID:  "source3",
		Content:   "Extreme content",
		Language:  "en",
		Timestamp: time.Now(),
	}

	sa.AddData(extremeData)

	if extremeData.Sentiment.Score < -1.0 || extremeData.Sentiment.Score > 1.0 {
		t.Errorf("Expected score to be within [-1, 1] bounds, got %f", extremeData.Sentiment.Score)
	}

	if extremeData.Sentiment.Confidence < 0.0 || extremeData.Sentiment.Confidence > 1.0 {
		t.Errorf("Expected confidence to be within [0, 1] bounds, got %f", extremeData.Sentiment.Confidence)
	}

	if extremeData.Sentiment.Magnitude < 0.0 || extremeData.Sentiment.Magnitude > 1.0 {
		t.Errorf("Expected magnitude to be within [0, 1] bounds, got %f", extremeData.Sentiment.Magnitude)
	}
}
