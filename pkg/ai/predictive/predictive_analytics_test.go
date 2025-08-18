package predictive

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func TestNewPredictiveAnalytics(t *testing.T) {
	// Test with default config
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	if pa == nil {
		t.Fatal("Expected PredictiveAnalytics instance, got nil")
	}

	if pa.ID == "" {
		t.Error("Expected non-empty ID")
	}

	if pa.Config.MaxModels != 50 {
		t.Errorf("Expected MaxModels to be 50, got %d", pa.Config.MaxModels)
	}

	if pa.Config.MaxFeatures != 10000 {
		t.Errorf("Expected MaxFeatures to be 10000, got %d", pa.Config.MaxFeatures)
	}

	if pa.Config.TrainingInterval != 24*time.Hour {
		t.Errorf("Expected TrainingInterval to be 24h, got %v", pa.Config.TrainingInterval)
	}

	if pa.Config.EvaluationInterval != 6*time.Hour {
		t.Errorf("Expected EvaluationInterval to be 6h, got %v", pa.Config.EvaluationInterval)
	}

	if pa.Config.PredictionHorizon != 24*time.Hour {
		t.Errorf("Expected PredictionHorizon to be 24h, got %v", pa.Config.PredictionHorizon)
	}

	if pa.Config.MinDataPoints != 1000 {
		t.Errorf("Expected MinDataPoints to be 1000, got %d", pa.Config.MinDataPoints)
	}

	if pa.Config.ModelRetentionDays != 365 {
		t.Errorf("Expected ModelRetentionDays to be 365, got %d", pa.Config.ModelRetentionDays)
	}
}

func TestNewPredictiveAnalyticsCustomConfig(t *testing.T) {
	config := AnalyticsConfig{
		MaxModels:          100,
		MaxFeatures:        20000,
		TrainingInterval:   12 * time.Hour,
		EvaluationInterval: 3 * time.Hour,
		PredictionHorizon:  48 * time.Hour,
		MinDataPoints:      2000,
		EnableAutoTraining: true,
		ModelRetentionDays: 730,
	}

	pa := NewPredictiveAnalytics(config)

	if pa.Config.MaxModels != 100 {
		t.Errorf("Expected MaxModels to be 100, got %d", pa.Config.MaxModels)
	}

	if pa.Config.MaxFeatures != 20000 {
		t.Errorf("Expected MaxFeatures to be 20000, got %d", pa.Config.MaxFeatures)
	}

	if pa.Config.TrainingInterval != 12*time.Hour {
		t.Errorf("Expected TrainingInterval to be 12h, got %v", pa.Config.TrainingInterval)
	}

	if pa.Config.EvaluationInterval != 3*time.Hour {
		t.Errorf("Expected EvaluationInterval to be 3h, got %v", pa.Config.EvaluationInterval)
	}

	if pa.Config.PredictionHorizon != 48*time.Hour {
		t.Errorf("Expected PredictionHorizon to be 48h, got %v", pa.Config.PredictionHorizon)
	}

	if pa.Config.MinDataPoints != 2000 {
		t.Errorf("Expected MinDataPoints to be 2000, got %d", pa.Config.MinDataPoints)
	}

	if pa.Config.ModelRetentionDays != 730 {
		t.Errorf("Expected ModelRetentionDays to be 730, got %d", pa.Config.ModelRetentionDays)
	}
}

func TestStartStop(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test Start
	err := pa.Start()
	if err != nil {
		t.Errorf("Expected Start to succeed, got error: %v", err)
	}

	// Test Stop
	err = pa.Stop()
	if err != nil {
		t.Errorf("Expected Stop to succeed, got error: %v", err)
	}
}

func TestCreateModel(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test successful model creation
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price", "volume", "volatility"},
		map[string]interface{}{
			"n_estimators": 100,
			"max_depth":    10,
		},
	)

	if err != nil {
		t.Errorf("Expected CreateModel to succeed, got error: %v", err)
	}

	if model == nil {
		t.Fatal("Expected model instance, got nil")
	}

	if model.Name != "TestModel" {
		t.Errorf("Expected Name to be 'TestModel', got %s", model.Name)
	}

	if model.Type != ModelTypeRandomForest {
		t.Errorf("Expected Type to be ModelTypeRandomForest, got %v", model.Type)
	}

	if model.Asset != "BTC" {
		t.Errorf("Expected Asset to be 'BTC', got %s", model.Asset)
	}

	if model.PredictionType != PredictionTypePrice {
		t.Errorf("Expected PredictionType to be PredictionTypePrice, got %v", model.PredictionType)
	}

	if len(model.Features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(model.Features))
	}

	if model.Status != ModelStatusTraining {
		t.Errorf("Expected Status to be ModelStatusTraining, got %v", model.Status)
	}

	if pa.Metrics.TotalModels != 1 {
		t.Errorf("Expected TotalModels to be 1, got %d", pa.Metrics.TotalModels)
	}
}

func TestCreateModelValidation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test empty name
	_, err := pa.CreateModel("", ModelTypeRandomForest, "BTC", PredictionTypePrice, []string{"price"}, nil)
	if err == nil {
		t.Error("Expected error for empty name")
	}

	// Test empty features
	_, err = pa.CreateModel("TestModel", ModelTypeRandomForest, "BTC", PredictionTypePrice, []string{}, nil)
	if err == nil {
		t.Error("Expected error for empty features")
	}

	// Test nil features
	_, err = pa.CreateModel("TestModel", ModelTypeRandomForest, "BTC", PredictionTypePrice, nil, nil)
	if err == nil {
		t.Error("Expected error for nil features")
	}
}

func TestCreateModelMaxLimit(t *testing.T) {
	config := AnalyticsConfig{MaxModels: 2}
	pa := NewPredictiveAnalytics(config)

	// Create first model
	_, err := pa.CreateModel("Model1", ModelTypeRandomForest, "BTC", PredictionTypePrice, []string{"price"}, nil)
	if err != nil {
		t.Errorf("Expected first model creation to succeed, got error: %v", err)
	}

	// Create second model
	_, err = pa.CreateModel("Model2", ModelTypeRandomForest, "ETH", PredictionTypePrice, []string{"price"}, nil)
	if err != nil {
		t.Errorf("Expected second model creation to succeed, got error: %v", err)
	}

	// Try to create third model (should fail)
	_, err = pa.CreateModel("Model3", ModelTypeRandomForest, "ADA", PredictionTypePrice, []string{"price"}, nil)
	if err == nil {
		t.Error("Expected error when exceeding MaxModels limit")
	}
}

func TestTrainModel(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create model
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price", "volume"},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create training features (need at least 1000 data points)
	var features []MarketFeature
	baseTime := time.Now().Add(-time.Hour * 24) // Start 24 hours ago

	for i := 0; i < 1000; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		priceValue := 50000.0 + float64(i)*10.0      // Gradual price increase
		volumeValue := 1000000.0 + float64(i)*1000.0 // Gradual volume increase

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     priceValue,
			Feature:   "price",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volumeValue,
			Feature:   "volume",
			Source:    "test",
		})
	}

	// Train model
	err = pa.TrainModel(model.ID, features)
	if err != nil {
		t.Errorf("Expected TrainModel to succeed, got error: %v", err)
	}

	// Verify model status changed
	if model.Status != ModelStatusTrained {
		t.Errorf("Expected Status to be ModelStatusTrained, got %v", model.Status)
	}

	// Verify performance metrics were set
	if model.Performance.Accuracy == 0 {
		t.Error("Expected Accuracy to be set")
	}

	if model.Performance.TrainingTime == 0 {
		t.Error("Expected TrainingTime to be set")
	}

	if model.Performance.LastEvaluation.IsZero() {
		t.Error("Expected LastEvaluation to be set")
	}

	// Verify training data was set
	if model.TrainingData.DataPoints != 2000 { // 1000 price + 1000 volume
		t.Errorf("Expected DataPoints to be 2000, got %d", model.TrainingData.DataPoints)
	}

	if model.TrainingData.Features != 2 {
		t.Errorf("Expected Features to be 2, got %d", model.TrainingData.Features)
	}

	if pa.Metrics.TrainedModels != 1 {
		t.Errorf("Expected TrainedModels to be 1, got %d", pa.Metrics.TrainedModels)
	}
}

func TestTrainModelValidation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create model
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price"},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test training with insufficient data
	features := []MarketFeature{
		{Timestamp: time.Now(), Asset: "BTC", Value: 50000.0, Feature: "price", Source: "test"},
	}

	err = pa.TrainModel(model.ID, features)
	if err == nil {
		t.Error("Expected error for insufficient data points")
	}

	// Test training non-existent model
	err = pa.TrainModel("nonexistent", features)
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

func TestDeployModel(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create and train model
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price"},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Try to deploy untrained model (should fail)
	err = pa.DeployModel(model.ID)
	if err == nil {
		t.Error("Expected error when deploying untrained model")
	}

	// Train model (need at least 1000 data points)
	var features []MarketFeature
	baseTime := time.Now().Add(-time.Hour * 24) // Start 24 hours ago

	for i := 0; i < 1000; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		priceValue := 50000.0 + float64(i)*10.0
		volumeValue := 1000000.0 + float64(i)*1000.0
		volatilityValue := 0.25 + float64(i)*0.0001
		sentimentValue := 0.6 + float64(i)*0.0001

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     priceValue,
			Feature:   "price",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volumeValue,
			Feature:   "volume",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volatilityValue,
			Feature:   "volatility",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     sentimentValue,
			Feature:   "sentiment",
			Source:    "test",
		})
	}

	err = pa.TrainModel(model.ID, features)
	if err != nil {
		t.Fatalf("Failed to train model: %v", err)
	}

	// Deploy model
	err = pa.DeployModel(model.ID)
	if err != nil {
		t.Errorf("Expected DeployModel to succeed, got error: %v", err)
	}

	// Verify model status changed
	if model.Status != ModelStatusDeployed {
		t.Errorf("Expected Status to be ModelStatusDeployed, got %v", model.Status)
	}

	if pa.Metrics.DeployedModels != 1 {
		t.Errorf("Expected DeployedModels to be 1, got %d", pa.Metrics.DeployedModels)
	}
}

func TestDeployModelValidation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test deploying non-existent model
	err := pa.DeployModel("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}

	// Create model but don't train it
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price"},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Try to deploy untrained model
	err = pa.DeployModel(model.ID)
	if err == nil {
		t.Error("Expected error when deploying untrained model")
	}
}

func TestMakePrediction(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create, train, and deploy model
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price", "volume", "volatility"},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Train model (need at least 1000 data points)
	var features []MarketFeature
	baseTime := time.Now().Add(-time.Hour * 24) // Start 24 hours ago

	for i := 0; i < 1000; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		priceValue := 50000.0 + float64(i)*10.0
		volumeValue := 1000000.0 + float64(i)*1000.0
		volatilityValue := 0.25 + float64(i)*0.0001

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     priceValue,
			Feature:   "price",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volumeValue,
			Feature:   "volume",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volatilityValue,
			Feature:   "volatility",
			Source:    "test",
		})
	}

	err = pa.TrainModel(model.ID, features)
	if err != nil {
		t.Fatalf("Failed to train model: %v", err)
	}

	err = pa.DeployModel(model.ID)
	if err != nil {
		t.Fatalf("Failed to deploy model: %v", err)
	}

	// Make prediction
	predictionFeatures := map[string]float64{
		"price":      52000.0,
		"volume":     1200000.0,
		"volatility": 0.27,
	}

	prediction, err := pa.MakePrediction(model.ID, predictionFeatures, 24*time.Hour)
	if err != nil {
		t.Errorf("Expected MakePrediction to succeed, got error: %v", err)
	}

	if prediction == nil {
		t.Fatal("Expected prediction instance, got nil")
	}

	if prediction.ModelID != model.ID {
		t.Errorf("Expected ModelID to be %s, got %s", model.ID, prediction.ModelID)
	}

	if prediction.Asset != "BTC" {
		t.Errorf("Expected Asset to be 'BTC', got %s", prediction.Asset)
	}

	if prediction.Type != PredictionTypePrice {
		t.Errorf("Expected Type to be PredictionTypePrice, got %v", prediction.Type)
	}

	if prediction.Value <= 0 {
		t.Errorf("Expected positive Value, got %f", prediction.Value)
	}

	if prediction.Confidence <= 0 || prediction.Confidence > 1 {
		t.Errorf("Expected Confidence between 0 and 1, got %f", prediction.Confidence)
	}

	if prediction.Horizon != 24*time.Hour {
		t.Errorf("Expected Horizon to be 24h, got %v", prediction.Horizon)
	}

	if len(prediction.Features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(prediction.Features))
	}

	if pa.Metrics.TotalPredictions != 1 {
		t.Errorf("Expected TotalPredictions to be 1, got %d", pa.Metrics.TotalPredictions)
	}
}

func TestMakePredictionValidation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test prediction with non-existent model
	_, err := pa.MakePrediction("nonexistent", map[string]float64{"price": 50000.0}, time.Hour)
	if err == nil {
		t.Error("Expected error for non-existent model")
	}

	// Create model but don't train/deploy it
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price", "volume", "volatility"},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Try to make prediction with untrained model
	_, err = pa.MakePrediction(model.ID, map[string]float64{"price": 50000.0}, time.Hour)
	if err == nil {
		t.Error("Expected error for untrained model")
	}

	// Train model but don't deploy it (need at least 1000 data points)
	var features []MarketFeature
	baseTime := time.Now().Add(-time.Hour * 24) // Start 24 hours ago

	for i := 0; i < 1000; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		priceValue := 50000.0 + float64(i)*10.0
		volumeValue := 1000000.0 + float64(i)*1000.0
		volatilityValue := 0.25 + float64(i)*0.0001
		sentimentValue := 0.6 + float64(i)*0.0001

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     priceValue,
			Feature:   "price",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volumeValue,
			Feature:   "volume",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volatilityValue,
			Feature:   "volatility",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     sentimentValue,
			Feature:   "sentiment",
			Source:    "test",
		})
	}

	err = pa.TrainModel(model.ID, features)
	if err != nil {
		t.Fatalf("Failed to train model: %v", err)
	}

	// Try to make prediction with trained but not deployed model
	_, err = pa.MakePrediction(model.ID, map[string]float64{"price": 50000.0}, time.Hour)
	if err == nil {
		t.Error("Expected error for non-deployed model")
	}

	// Deploy model
	err = pa.DeployModel(model.ID)
	if err != nil {
		t.Fatalf("Failed to deploy model: %v", err)
	}

	// Test prediction with missing required features
	_, err = pa.MakePrediction(model.ID, map[string]float64{}, time.Hour)
	if err == nil {
		t.Error("Expected error for missing required features")
	}

	// Test prediction with incomplete features
	_, err = pa.MakePrediction(model.ID, map[string]float64{"price": 50000.0}, time.Hour)
	if err == nil {
		t.Error("Expected error for incomplete features")
	}
}

func TestAssessRisk(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test risk assessment with various feature combinations
	testCases := []struct {
		name     string
		features map[string]float64
		expected RiskLevel
	}{
		{
			name: "Low Risk",
			features: map[string]float64{
				"volatility":   0.1,
				"volume":       500000,
				"price_change": 0.02,
				"market_cap":   2000000000,
			},
			expected: RiskLevelLow,
		},
		{
			name: "Medium Risk",
			features: map[string]float64{
				"volatility":   0.35,
				"volume":       1000000,
				"price_change": 0.12,
				"market_cap":   1200000000,
			},
			expected: RiskLevelMedium,
		},
		{
			name: "High Risk",
			features: map[string]float64{
				"volatility":   0.6,
				"volume":       2000000,
				"price_change": 0.25,
				"market_cap":   400000000,
			},
			expected: RiskLevelMedium,
		},
		{
			name: "Critical Risk",
			features: map[string]float64{
				"volatility":   0.8,
				"volume":       3000000,
				"price_change": 0.35,
				"market_cap":   200000000,
			},
			expected: RiskLevelHigh,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assessment, err := pa.AssessRisk("BTC", tc.features)
			if err != nil {
				t.Errorf("Expected AssessRisk to succeed, got error: %v", err)
			}

			if assessment == nil {
				t.Fatal("Expected assessment instance, got nil")
			}

			if assessment.Asset != "BTC" {
				t.Errorf("Expected Asset to be 'BTC', got %s", assessment.Asset)
			}

			if assessment.RiskLevel != tc.expected {
				t.Errorf("Expected RiskLevel to be %v, got %v", tc.expected, assessment.RiskLevel)
			}

			if assessment.RiskScore < 0 || assessment.RiskScore > 1 {
				t.Errorf("Expected RiskScore between 0 and 1, got %f", assessment.RiskScore)
			}

			if assessment.VaR >= 0 {
				t.Errorf("Expected VaR to be negative, got %f", assessment.VaR)
			}

			if assessment.ExpectedShortfall >= 0 {
				t.Errorf("Expected ExpectedShortfall to be negative, got %f", assessment.ExpectedShortfall)
			}

			if assessment.Volatility <= 0 {
				t.Errorf("Expected positive Volatility, got %f", assessment.Volatility)
			}

			if len(assessment.Correlation) == 0 {
				t.Error("Expected non-empty correlation map")
			}

			if len(assessment.Factors) == 0 {
				t.Error("Expected non-empty risk factors")
			}
		})
	}
}

func TestGetModel(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create model
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price"},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Get model
	retrievedModel, err := pa.GetModel(model.ID)
	if err != nil {
		t.Errorf("Expected GetModel to succeed, got error: %v", err)
	}

	if retrievedModel == nil {
		t.Fatal("Expected model instance, got nil")
	}

	if retrievedModel.ID != model.ID {
		t.Errorf("Expected ID to be %s, got %s", model.ID, retrievedModel.ID)
	}

	if retrievedModel.Name != model.Name {
		t.Errorf("Expected Name to be %s, got %s", model.Name, retrievedModel.Name)
	}

	// Test getting non-existent model
	_, err = pa.GetModel("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent model")
	}
}

func TestGetModels(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create multiple models
	model1, err := pa.CreateModel("Model1", ModelTypeRandomForest, "BTC", PredictionTypePrice, []string{"price"}, nil)
	if err != nil {
		t.Fatalf("Failed to create model1: %v", err)
	}

	model2, err := pa.CreateModel("Model2", ModelTypeNeuralNetwork, "ETH", PredictionTypeVolatility, []string{"volatility"}, nil)
	if err != nil {
		t.Fatalf("Failed to create model2: %v", err)
	}

	// Get all models
	models := pa.GetModels()

	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	// Verify both models are present
	if _, exists := models[model1.ID]; !exists {
		t.Error("Expected model1 to be present in results")
	}

	if _, exists := models[model2.ID]; !exists {
		t.Error("Expected model2 to be present in results")
	}

	// Verify models are copies (not references)
	models[model1.ID].Name = "ModifiedName"
	if pa.Models[model1.ID].Name == "ModifiedName" {
		t.Error("Expected original model to not be modified when copy is changed")
	}
}

func TestAddFeature(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{MaxFeatures: 3})

	// Add features
	feature1 := MarketFeature{
		Timestamp: time.Now().Add(-time.Hour),
		Asset:     "BTC",
		Value:     50000.0,
		Feature:   "price",
		Source:    "test",
	}

	err := pa.AddFeature(feature1)
	if err != nil {
		t.Errorf("Expected AddFeature to succeed, got error: %v", err)
	}

	key := "BTC_price"
	if len(pa.Features[key]) != 1 {
		t.Errorf("Expected 1 feature, got %d", len(pa.Features[key]))
	}

	// Add more features to test overflow
	feature2 := MarketFeature{
		Timestamp: time.Now().Add(-30 * time.Minute),
		Asset:     "BTC",
		Value:     51000.0,
		Feature:   "price",
		Source:    "test",
	}

	err = pa.AddFeature(feature2)
	if err != nil {
		t.Errorf("Expected AddFeature to succeed, got error: %v", err)
	}

	feature3 := MarketFeature{
		Timestamp: time.Now(),
		Asset:     "BTC",
		Value:     52000.0,
		Feature:   "price",
		Source:    "test",
	}

	err = pa.AddFeature(feature3)
	if err != nil {
		t.Errorf("Expected AddFeature to succeed, got error: %v", err)
	}

	// Add fourth feature (should remove oldest)
	feature4 := MarketFeature{
		Timestamp: time.Now().Add(time.Minute),
		Asset:     "BTC",
		Value:     53000.0,
		Feature:   "price",
		Source:    "test",
	}

	err = pa.AddFeature(feature4)
	if err != nil {
		t.Errorf("Expected AddFeature to succeed, got error: %v", err)
	}

	// Verify oldest feature was removed
	if len(pa.Features[key]) != 3 {
		t.Errorf("Expected 3 features after overflow, got %d", len(pa.Features[key]))
	}

	// Verify first feature was removed (oldest)
	if pa.Features[key][0].Value != 51000.0 {
		t.Errorf("Expected oldest feature to be removed, got value %f", pa.Features[key][0].Value)
	}
}

func TestGetMetrics(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Get initial metrics
	metrics := pa.GetMetrics()

	if metrics.TotalModels != 0 {
		t.Errorf("Expected TotalModels to be 0, got %d", metrics.TotalModels)
	}

	if metrics.ActiveModels != 0 {
		t.Errorf("Expected ActiveModels to be 0, got %d", metrics.ActiveModels)
	}

	if metrics.TrainedModels != 0 {
		t.Errorf("Expected TrainedModels to be 0, got %d", metrics.TrainedModels)
	}

	if metrics.DeployedModels != 0 {
		t.Errorf("Expected DeployedModels to be 0, got %d", metrics.DeployedModels)
	}

	if metrics.TotalPredictions != 0 {
		t.Errorf("Expected TotalPredictions to be 0, got %d", metrics.TotalPredictions)
	}

	if metrics.AverageAccuracy != 0 {
		t.Errorf("Expected AverageAccuracy to be 0, got %f", metrics.AverageAccuracy)
	}

	// Create and train a model
	model, err := pa.CreateModel(
		"TestModel",
		ModelTypeRandomForest,
		"BTC",
		PredictionTypePrice,
		[]string{"price"},
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create training features (need at least 1000 data points)
	var features []MarketFeature
	baseTime := time.Now().Add(-time.Hour * 24) // Start 24 hours ago

	for i := 0; i < 1000; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)
		priceValue := 50000.0 + float64(i)*10.0
		volumeValue := 1000000.0 + float64(i)*1000.0
		volatilityValue := 0.25 + float64(i)*0.0001
		sentimentValue := 0.6 + float64(i)*0.0001

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     priceValue,
			Feature:   "price",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volumeValue,
			Feature:   "volume",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     volatilityValue,
			Feature:   "volatility",
			Source:    "test",
		})

		features = append(features, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     sentimentValue,
			Feature:   "sentiment",
			Source:    "test",
		})
	}

	err = pa.TrainModel(model.ID, features)
	if err != nil {
		t.Fatalf("Failed to train model: %v", err)
	}

	// Get updated metrics
	metrics = pa.GetMetrics()

	if metrics.TotalModels != 1 {
		t.Errorf("Expected TotalModels to be 1, got %d", metrics.TotalModels)
	}

	if metrics.ActiveModels != 1 {
		t.Errorf("Expected ActiveModels to be 1, got %d", metrics.ActiveModels)
	}

	if metrics.TrainedModels != 1 {
		t.Errorf("Expected TrainedModels to be 1, got %d", metrics.TrainedModels)
	}

	if metrics.AverageAccuracy <= 0 {
		t.Errorf("Expected positive AverageAccuracy, got %f", metrics.AverageAccuracy)
	}
}

func TestUtilityFunctions(t *testing.T) {
	// Test copyStringSlice
	original := []string{"a", "b", "c"}
	copied := copyStringSlice(original)

	if len(copied) != len(original) {
		t.Errorf("Expected copied slice length to be %d, got %d", len(original), len(copied))
	}

	for i, v := range original {
		if copied[i] != v {
			t.Errorf("Expected copied[%d] to be %s, got %s", i, v, copied[i])
		}
	}

	// Verify it's a deep copy
	copied[0] = "x"
	if original[0] == "x" {
		t.Error("Expected original slice to not be modified when copy is changed")
	}

	// Test copyMap
	originalMap := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	copiedMap := copyMap(originalMap)

	if len(copiedMap) != len(originalMap) {
		t.Errorf("Expected copied map length to be %d, got %d", len(originalMap), len(copiedMap))
	}

	for k, v := range originalMap {
		if copiedMap[k] != v {
			t.Errorf("Expected copiedMap[%s] to be %v, got %v", k, v, copiedMap[k])
		}
	}

	// Verify it's a deep copy
	copiedMap["key1"] = "modified"
	if originalMap["key1"] == "modified" {
		t.Error("Expected original map to not be modified when copy is changed")
	}

	// Test copyFloatMap
	originalFloatMap := map[string]float64{
		"price":      50000.0,
		"volume":     1000000.0,
		"volatility": 0.25,
	}

	copiedFloatMap := copyFloatMap(originalFloatMap)

	if len(copiedFloatMap) != len(originalFloatMap) {
		t.Errorf("Expected copied float map length to be %d, got %d", len(originalFloatMap), len(copiedFloatMap))
	}

	for k, v := range originalFloatMap {
		if copiedFloatMap[k] != v {
			t.Errorf("Expected copiedFloatMap[%s] to be %f, got %f", k, v, copiedFloatMap[k])
		}
	}

	// Verify it's a deep copy
	copiedFloatMap["price"] = 60000.0
	if originalFloatMap["price"] == 60000.0 {
		t.Error("Expected original float map to not be modified when copy is changed")
	}
}

func TestIDGeneration(t *testing.T) {
	// Test analytics ID generation
	id1 := generateAnalyticsID()
	id2 := generateAnalyticsID()

	if id1 == id2 {
		t.Error("Expected unique analytics IDs")
	}

	if len(id1) == 0 {
		t.Error("Expected non-empty analytics ID")
	}

	// Test model ID generation
	modelID1 := generateModelID()
	modelID2 := generateModelID()

	if modelID1 == modelID2 {
		t.Error("Expected unique model IDs")
	}

	if len(modelID1) == 0 {
		t.Error("Expected non-empty model ID")
	}

	// Test prediction ID generation
	predID1 := generatePredictionID()
	predID2 := generatePredictionID()

	if predID1 == predID2 {
		t.Error("Expected unique prediction IDs")
	}

	if len(predID1) == 0 {
		t.Error("Expected non-empty prediction ID")
	}

	// Test risk assessment ID generation
	riskID1 := generateRiskAssessmentID()
	riskID2 := generateRiskAssessmentID()

	if riskID1 == riskID2 {
		t.Error("Expected unique risk assessment IDs")
	}

	if len(riskID1) == 0 {
		t.Error("Expected non-empty risk assessment ID")
	}
}

func TestModelPerformanceGeneration(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test performance generation for different model types
	modelTypes := []ModelType{
		ModelTypeLinearRegression,
		ModelTypeRandomForest,
		ModelTypeNeuralNetwork,
		ModelTypeLSTM,
		ModelTypeXGBoost,
	}

	for _, modelType := range modelTypes {
		t.Run(fmt.Sprintf("ModelType_%d", modelType), func(t *testing.T) {
			performance := pa.generateSimulatedPerformance(modelType)

			// Verify performance metrics are within expected ranges
			if performance.Accuracy < 0.6 || performance.Accuracy > 0.98 {
				t.Errorf("Expected Accuracy between 0.6 and 0.98, got %f", performance.Accuracy)
			}

			if performance.Precision < 0.5 || performance.Precision > 1.1 {
				t.Errorf("Expected Precision between 0.5 and 1.1, got %f", performance.Precision)
			}

			if performance.Recall < 0.5 || performance.Recall > 1.1 {
				t.Errorf("Expected Recall between 0.5 and 1.1, got %f", performance.Recall)
			}

			if performance.F1Score < 0.5 || performance.F1Score > 1.1 {
				t.Errorf("Expected F1Score between 0.5 and 1.1, got %f", performance.F1Score)
			}

			if performance.RMSE < 0 || performance.RMSE > 1.1 {
				t.Errorf("Expected RMSE between 0 and 1.1, got %f", performance.RMSE)
			}

			if performance.MAE < 0 || performance.MAE > 1.1 {
				t.Errorf("Expected MAE between 0 and 1.1, got %f", performance.MAE)
			}

			if performance.R2Score < 0 || performance.R2Score > 1.1 {
				t.Errorf("Expected R2Score between 0 and 1.1, got %f", performance.R2Score)
			}

			if performance.LastEvaluation.IsZero() {
				t.Error("Expected LastEvaluation to be set")
			}
		})
	}
}

func TestPredictionGeneration(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create a test model
	model := &Model{
		ID:       "test_model",
		Type:     ModelTypeRandomForest,
		Asset:    "BTC",
		Features: []string{"price", "volume"},
		Performance: ModelPerformance{
			Accuracy: 0.85,
		},
	}

	// Test prediction generation with various features
	testCases := []struct {
		name              string
		features          map[string]float64
		expectedBaseValue float64
	}{
		{
			name: "With Price Feature",
			features: map[string]float64{
				"price":  50000.0,
				"volume": 1000000.0,
			},
			expectedBaseValue: 50000.0,
		},
		{
			name: "With Close Feature",
			features: map[string]float64{
				"close":  52000.0,
				"volume": 1200000.0,
			},
			expectedBaseValue: 52000.0,
		},
		{
			name: "Without Price/Close",
			features: map[string]float64{
				"volume": 1000000.0,
			},
			expectedBaseValue: 100.0, // Default base value
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prediction := pa.generatePredictionValue(model, tc.features)

			if prediction <= 0 {
				t.Errorf("Expected positive prediction, got %f", prediction)
			}

			// Prediction should be close to base value (within reasonable noise range)
			noiseRange := tc.expectedBaseValue * 0.4 // Allow for noise
			if prediction < tc.expectedBaseValue-noiseRange || prediction > tc.expectedBaseValue+noiseRange {
				t.Errorf("Expected prediction close to %f, got %f", tc.expectedBaseValue, prediction)
			}
		})
	}
}

func TestConfidenceCalculation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create test models with different accuracy levels
	testCases := []struct {
		name                  string
		accuracy              float64
		features              map[string]float64
		expectedMinConfidence float64
	}{
		{
			name:                  "High Accuracy, Complete Features",
			accuracy:              0.95,
			features:              map[string]float64{"price": 50000.0, "volume": 1000000.0},
			expectedMinConfidence: 0.85,
		},
		{
			name:                  "Medium Accuracy, Complete Features",
			accuracy:              0.80,
			features:              map[string]float64{"price": 50000.0, "volume": 1000000.0},
			expectedMinConfidence: 0.70,
		},
		{
			name:                  "High Accuracy, Partial Features",
			accuracy:              0.95,
			features:              map[string]float64{"price": 50000.0},
			expectedMinConfidence: 0.40,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model := &Model{
				ID:       "test_model",
				Features: []string{"price", "volume"},
				Performance: ModelPerformance{
					Accuracy: tc.accuracy,
				},
			}

			confidence := pa.calculateConfidence(model, tc.features)

			if confidence < 0.1 || confidence > 0.99 {
				t.Errorf("Expected confidence between 0.1 and 0.99, got %f", confidence)
			}

			if confidence < tc.expectedMinConfidence {
				t.Errorf("Expected confidence >= %f, got %f", tc.expectedMinConfidence, confidence)
			}
		})
	}
}

func TestRiskScoreCalculation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test risk score calculation with various feature combinations
	testCases := []struct {
		name        string
		features    map[string]float64
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "Low Risk Features",
			features: map[string]float64{
				"volatility":   0.1,
				"volume":       500000,
				"price_change": 0.02,
				"market_cap":   2000000000,
			},
			expectedMin: 0.0,
			expectedMax: 0.3,
		},
		{
			name: "Medium Risk Features",
			features: map[string]float64{
				"volatility":   0.35,
				"volume":       1000000,
				"price_change": 0.12,
				"market_cap":   1200000000,
			},
			expectedMin: 0.25,
			expectedMax: 0.65,
		},
		{
			name: "High Risk Features",
			features: map[string]float64{
				"volatility":   0.6,
				"volume":       2000000,
				"price_change": 0.25,
				"market_cap":   400000000,
			},
			expectedMin: 0.55,
			expectedMax: 0.65,
		},
		{
			name: "Critical Risk Features",
			features: map[string]float64{
				"volatility":   0.8,
				"volume":       3000000,
				"price_change": 0.35,
				"market_cap":   200000000,
			},
			expectedMin: 0.65,
			expectedMax: 0.75,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			riskScore := pa.calculateRiskScore(tc.features)

			if riskScore < tc.expectedMin || riskScore > tc.expectedMax {
				t.Errorf("Expected risk score between %f and %f, got %f", tc.expectedMin, tc.expectedMax, riskScore)
			}
		})
	}
}

func TestRiskLevelClassification(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test risk level classification
	testCases := []struct {
		riskScore     float64
		expectedLevel RiskLevel
	}{
		{0.1, RiskLevelLow},
		{0.2, RiskLevelLow},
		{0.3, RiskLevelMedium},
		{0.4, RiskLevelMedium},
		{0.5, RiskLevelMedium},
		{0.6, RiskLevelHigh},
		{0.7, RiskLevelHigh},
		{0.8, RiskLevelCritical},
		{0.9, RiskLevelCritical},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("RiskScore_%.1f", tc.riskScore), func(t *testing.T) {
			level := pa.classifyRiskLevel(tc.riskScore)

			if level != tc.expectedLevel {
				t.Errorf("Expected risk level %v for score %f, got %v", tc.expectedLevel, tc.riskScore, level)
			}
		})
	}
}

func TestVaRCalculation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test VaR calculation
	testCases := []struct {
		volatility  float64
		expectedVaR float64
	}{
		{0.1, -0.1645}, // 10% volatility
		{0.2, -0.329},  // 20% volatility
		{0.3, -0.4935}, // 30% volatility
		{0.5, -0.8225}, // 50% volatility
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Volatility_%.1f", tc.volatility), func(t *testing.T) {
			features := map[string]float64{"volatility": tc.volatility}
			varValue := pa.calculateVaR(features)

			// Allow for small floating point differences
			tolerance := 0.001
			if math.Abs(varValue-tc.expectedVaR) > tolerance {
				t.Errorf("Expected VaR %f for volatility %f, got %f", tc.expectedVaR, tc.volatility, varValue)
			}

			// VaR should always be negative
			if varValue >= 0 {
				t.Errorf("Expected negative VaR, got %f", varValue)
			}
		})
	}

	// Test VaR calculation without volatility feature (should use default)
	features := map[string]float64{}
	varValue := pa.calculateVaR(features)

	if varValue >= 0 {
		t.Errorf("Expected negative VaR for default volatility, got %f", varValue)
	}
}

func TestExpectedShortfallCalculation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test expected shortfall calculation
	testCases := []struct {
		varValue    float64
		expectedMin float64
		expectedMax float64
	}{
		{-0.1, -0.15, -0.2}, // VaR -0.1
		{-0.2, -0.3, -0.4},  // VaR -0.2
		{-0.5, -0.75, -1.0}, // VaR -0.5
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("VaR_%.1f", tc.varValue), func(t *testing.T) {
			features := map[string]float64{}
			expectedShortfall := pa.calculateExpectedShortfall(features, tc.varValue)

			if expectedShortfall < tc.expectedMax || expectedShortfall > tc.expectedMin {
				t.Errorf("Expected expected shortfall between %f and %f for VaR %f, got %f",
					tc.expectedMax, tc.expectedMin, tc.varValue, expectedShortfall)
			}

			// Expected shortfall should always be negative
			if expectedShortfall >= 0 {
				t.Errorf("Expected negative expected shortfall, got %f", expectedShortfall)
			}
		})
	}
}

func TestVolatilityCalculation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test volatility calculation with feature
	features := map[string]float64{"volatility": 0.25}
	volatility := pa.calculateVolatility(features)

	if volatility != 0.25 {
		t.Errorf("Expected volatility 0.25, got %f", volatility)
	}

	// Test volatility calculation without feature (should use default)
	features = map[string]float64{}
	volatility = pa.calculateVolatility(features)

	if volatility < 0.2 || volatility > 0.5 {
		t.Errorf("Expected default volatility between 0.2 and 0.5, got %f", volatility)
	}
}

func TestCorrelationCalculation(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	features := map[string]float64{}
	correlations := pa.calculateCorrelations(features)

	// Should have correlations with major assets
	expectedAssets := []string{"BTC", "ETH", "SPY", "QQQ", "GLD"}

	if len(correlations) != len(expectedAssets) {
		t.Errorf("Expected %d correlations, got %d", len(expectedAssets), len(correlations))
	}

	for _, asset := range expectedAssets {
		if _, exists := correlations[asset]; !exists {
			t.Errorf("Expected correlation with asset %s", asset)
		}

		// Correlation should be between -1 and 1
		corr := correlations[asset]
		if corr < -1 || corr > 1 {
			t.Errorf("Expected correlation for %s between -1 and 1, got %f", asset, corr)
		}
	}
}

func TestRiskFactorIdentification(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test risk factor identification with various features
	testCases := []struct {
		name            string
		features        map[string]float64
		expectedFactors int
	}{
		{
			name: "High Volatility",
			features: map[string]float64{
				"volatility": 0.4, // Above 0.3 threshold
			},
			expectedFactors: 2, // High Volatility + Market Sentiment
		},
		{
			name: "High Volume",
			features: map[string]float64{
				"volume": 600000, // Above 500k threshold
			},
			expectedFactors: 2, // High Volume + Market Sentiment
		},
		{
			name: "Large Price Movement",
			features: map[string]float64{
				"price_change": 0.12, // Above 0.1 threshold
			},
			expectedFactors: 2, // Large Price Movement + Market Sentiment
		},
		{
			name: "Multiple Risk Factors",
			features: map[string]float64{
				"volatility":   0.4,
				"volume":       600000,
				"price_change": 0.12,
			},
			expectedFactors: 4, // All three + Market Sentiment
		},
		{
			name: "Low Risk",
			features: map[string]float64{
				"volatility":   0.2,
				"volume":       400000,
				"price_change": 0.05,
			},
			expectedFactors: 1, // Only Market Sentiment
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			factors := pa.identifyRiskFactors(tc.features)

			if len(factors) != tc.expectedFactors {
				t.Errorf("Expected %d risk factors, got %d", tc.expectedFactors, len(factors))
			}

			// Verify all factors have required fields
			for i, factor := range factors {
				if factor.Factor == "" {
					t.Errorf("Factor %d has empty Factor field", i)
				}

				if factor.Impact <= 0 {
					t.Errorf("Factor %d has non-positive Impact: %f", i, factor.Impact)
				}

				if factor.Weight <= 0 || factor.Weight > 1 {
					t.Errorf("Factor %d has invalid Weight: %f", i, factor.Weight)
				}

				if factor.Description == "" {
					t.Errorf("Factor %d has empty Description field", i)
				}
			}

			// Verify Market Sentiment factor is always present
			hasMarketSentiment := false
			for _, factor := range factors {
				if factor.Factor == "Market Sentiment" {
					hasMarketSentiment = true
					break
				}
			}

			if !hasMarketSentiment {
				t.Error("Expected Market Sentiment factor to always be present")
			}
		})
	}
}

func TestCopyModel(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create a model with all fields populated
	originalModel := &Model{
		ID:              "test_model",
		Name:            "Test Model",
		Type:            ModelTypeRandomForest,
		Status:          ModelStatusTrained,
		Asset:           "BTC",
		PredictionType:  PredictionTypePrice,
		Features:        []string{"price", "volume"},
		Hyperparameters: map[string]interface{}{"n_estimators": 100},
		Performance:     ModelPerformance{Accuracy: 0.85},
		TrainingData:    TrainingData{DataPoints: 1000},
		LastUpdate:      time.Now(),
	}

	// Copy the model
	copiedModel := pa.copyModel(originalModel)

	// Verify all fields are copied correctly
	if copiedModel.ID != originalModel.ID {
		t.Errorf("Expected ID %s, got %s", originalModel.ID, copiedModel.ID)
	}

	if copiedModel.Name != originalModel.Name {
		t.Errorf("Expected Name %s, got %s", originalModel.Name, copiedModel.Name)
	}

	if copiedModel.Type != originalModel.Type {
		t.Errorf("Expected Type %v, got %v", originalModel.Type, copiedModel.Type)
	}

	if copiedModel.Status != originalModel.Status {
		t.Errorf("Expected Status %v, got %v", originalModel.Status, copiedModel.Status)
	}

	if copiedModel.Asset != originalModel.Asset {
		t.Errorf("Expected Asset %s, got %s", originalModel.Asset, copiedModel.Asset)
	}

	if copiedModel.PredictionType != originalModel.PredictionType {
		t.Errorf("Expected PredictionType %v, got %v", originalModel.PredictionType, copiedModel.PredictionType)
	}

	if len(copiedModel.Features) != len(originalModel.Features) {
		t.Errorf("Expected %d features, got %d", len(originalModel.Features), len(copiedModel.Features))
	}

	for i, feature := range originalModel.Features {
		if copiedModel.Features[i] != feature {
			t.Errorf("Expected feature[%d] %s, got %s", i, feature, copiedModel.Features[i])
		}
	}

	if len(copiedModel.Hyperparameters) != len(originalModel.Hyperparameters) {
		t.Errorf("Expected %d hyperparameters, got %d", len(originalModel.Hyperparameters), len(copiedModel.Hyperparameters))
	}

	for k, v := range originalModel.Hyperparameters {
		if copiedModel.Hyperparameters[k] != v {
			t.Errorf("Expected hyperparameter[%s] %v, got %v", k, v, copiedModel.Hyperparameters[k])
		}
	}

	if copiedModel.Performance.Accuracy != originalModel.Performance.Accuracy {
		t.Errorf("Expected Performance.Accuracy %f, got %f", originalModel.Performance.Accuracy, copiedModel.Performance.Accuracy)
	}

	if copiedModel.TrainingData.DataPoints != originalModel.TrainingData.DataPoints {
		t.Errorf("Expected TrainingData.DataPoints %d, got %d", originalModel.TrainingData.DataPoints, copiedModel.TrainingData.DataPoints)
	}

	// Verify it's a deep copy
	copiedModel.Name = "Modified Name"
	if originalModel.Name == "Modified Name" {
		t.Error("Expected original model to not be modified when copy is changed")
	}

	copiedModel.Features[0] = "modified_feature"
	if originalModel.Features[0] == "modified_feature" {
		t.Error("Expected original model features to not be modified when copy is changed")
	}

	copiedModel.Hyperparameters["n_estimators"] = 200
	if originalModel.Hyperparameters["n_estimators"] == 200 {
		t.Error("Expected original model hyperparameters to not be modified when copy is changed")
	}
}

func TestUpdateMetrics(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Create and train multiple models
	model1, err := pa.CreateModel("Model1", ModelTypeRandomForest, "BTC", PredictionTypePrice, []string{"price"}, nil)
	if err != nil {
		t.Fatalf("Failed to create model1: %v", err)
	}

	model2, err := pa.CreateModel("Model2", ModelTypeNeuralNetwork, "ETH", PredictionTypeVolatility, []string{"volatility"}, nil)
	if err != nil {
		t.Fatalf("Failed to create model2: %v", err)
	}

	// Train both models (need at least 1000 data points each)
	var features1 []MarketFeature
	var features2 []MarketFeature
	baseTime := time.Now().Add(-time.Hour * 24) // Start 24 hours ago

	for i := 0; i < 1000; i++ {
		timestamp := baseTime.Add(time.Duration(i) * time.Minute)

		// BTC features for model1 (requires price)
		priceValue := 50000.0 + float64(i)*10.0
		features1 = append(features1, MarketFeature{
			Timestamp: timestamp,
			Asset:     "BTC",
			Value:     priceValue,
			Feature:   "price",
			Source:    "test",
		})

		// ETH features for model2 (requires volatility)
		ethVolatilityValue := 0.2 + float64(i)*0.0001

		features2 = append(features2, MarketFeature{
			Timestamp: timestamp,
			Asset:     "ETH",
			Value:     ethVolatilityValue,
			Feature:   "volatility",
			Source:    "test",
		})

	}

	err = pa.TrainModel(model1.ID, features1)
	if err != nil {
		t.Fatalf("Failed to train model1: %v", err)
	}

	err = pa.TrainModel(model2.ID, features2)
	if err != nil {
		t.Fatalf("Failed to train model2: %v", err)
	}

	

	// Deploy one model
	err = pa.DeployModel(model1.ID)
	if err != nil {
		t.Fatalf("Failed to deploy model1: %v", err)
	}

	// Metrics are automatically updated by TrainModel and DeployModel

	// Verify metrics are correct
	if pa.Metrics.TotalModels != 2 {
		t.Errorf("Expected TotalModels to be 2, got %d", pa.Metrics.TotalModels)
	}

	if pa.Metrics.ActiveModels != 2 {
		t.Errorf("Expected ActiveModels to be 2, got %d", pa.Metrics.ActiveModels)
	}

	if pa.Metrics.TrainedModels != 1 {
		t.Errorf("Expected TrainedModels to be 1 (after deploying one model), got %d", pa.Metrics.TrainedModels)
	}

	if pa.Metrics.DeployedModels != 1 {
		t.Errorf("Expected DeployedModels to be 1, got %d", pa.Metrics.DeployedModels)
	}

	if pa.Metrics.AverageAccuracy <= 0 {
		t.Errorf("Expected positive AverageAccuracy, got %f", pa.Metrics.AverageAccuracy)
	}

	if pa.Metrics.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set")
	}
}

func TestEdgeCases(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test with empty features map
	assessment, err := pa.AssessRisk("BTC", map[string]float64{})
	if err != nil {
		t.Errorf("Expected AssessRisk to succeed with empty features, got error: %v", err)
	}

	if assessment == nil {
		t.Fatal("Expected assessment instance, got nil")
	}

	// Test with nil features map
	assessment, err = pa.AssessRisk("BTC", nil)
	if err != nil {
		t.Errorf("Expected AssessRisk to succeed with nil features, got error: %v", err)
	}

	if assessment == nil {
		t.Fatal("Expected assessment instance, got nil")
	}

	// Test with extreme feature values
	extremeFeatures := map[string]float64{
		"volatility":   1.0,           // Maximum volatility
		"volume":       10000000,      // Very high volume
		"price_change": 1.0,           // Maximum price change
		"market_cap":   1000000000000, // Very high market cap
	}

	assessment, err = pa.AssessRisk("BTC", extremeFeatures)
	if err != nil {
		t.Errorf("Expected AssessRisk to succeed with extreme features, got error: %v", err)
	}

	if assessment.RiskScore < 0 || assessment.RiskScore > 1 {
		t.Errorf("Expected RiskScore between 0 and 1, got %f", assessment.RiskScore)
	}
}

func TestConcurrency(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test concurrent model creation
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			model, err := pa.CreateModel(
				fmt.Sprintf("ConcurrentModel%d", id),
				ModelTypeRandomForest,
				"BTC",
				PredictionTypePrice,
				[]string{"price"},
				nil,
			)

			if err != nil {
				t.Errorf("Goroutine %d: Expected CreateModel to succeed, got error: %v", id, err)
				return
			}

			if model == nil {
				t.Errorf("Goroutine %d: Expected model instance, got nil", id)
				return
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all models were created
	if pa.Metrics.TotalModels != uint64(numGoroutines) {
		t.Errorf("Expected %d total models, got %d", numGoroutines, pa.Metrics.TotalModels)
	}
}

func TestMemorySafety(t *testing.T) {
	pa := NewPredictiveAnalytics(AnalyticsConfig{})

	// Test that large feature sets don't cause memory issues
	largeFeatures := make(map[string]float64, 1000)
	for i := 0; i < 1000; i++ {
		largeFeatures[fmt.Sprintf("feature_%d", i)] = float64(i)
	}

	// This should not panic or cause memory issues
	assessment, err := pa.AssessRisk("BTC", largeFeatures)
	if err != nil {
		t.Errorf("Expected AssessRisk to succeed with large features, got error: %v", err)
	}

	if assessment == nil {
		t.Fatal("Expected assessment instance, got nil")
	}

	// Test with very long feature names
	longFeatureName := strings.Repeat("very_long_feature_name_", 100)
	featuresWithLongNames := map[string]float64{
		longFeatureName: 1.0,
	}

	assessment, err = pa.AssessRisk("BTC", featuresWithLongNames)
	if err != nil {
		t.Errorf("Expected AssessRisk to succeed with long feature names, got error: %v", err)
	}

	if assessment == nil {
		t.Fatal("Expected assessment instance, got nil")
	}
}
