package oracle

import (
	"context"
	"fmt"
	"math/big"
	"time"
)

// Simple in-file test mock to satisfy existing tests
type TestOracleProvider struct {
	name   string
	prices map[string]*PriceData
	reqs   uint64
	succ   uint64
	fails  uint64
}

func NewTestOracleProvider(name, description, url string, reliability float64) *TestOracleProvider {
	return &TestOracleProvider{name: name, prices: make(map[string]*PriceData)}
}

// Backward compatible constructor expected by tests
func NewMockOracleProvider(name, description, url string, reliability float64) *TestOracleProvider {
	return NewTestOracleProvider(name, description, url, reliability)
}

// Implement OracleProvider interface
func (p *TestOracleProvider) GetPrice(ctx context.Context, asset string) (*PriceData, error) {
	p.reqs++
	var price *big.Int
	if pd, ok := p.prices[asset]; ok && pd != nil && pd.Price != nil {
		price = new(big.Int).Set(pd.Price)
	} else {
		price = big.NewInt(int64(len(asset) * 100))
	}
	pd := &PriceData{
		Asset:       asset,
		Price:       price,
		Timestamp:   time.Now(),
		BlockNumber: 0,
		Provider:    p.name,
		Confidence:  90,
		Source:      "test",
	}
	p.succ++
	return pd, nil
}

func (p *TestOracleProvider) ValidateProof(ctx context.Context, proof *OracleProof) error {
	if proof == nil {
		return fmt.Errorf("invalid proof: nil")
	}
	return nil
}

func (p *TestOracleProvider) UpdatePrice(ctx context.Context, asset string, price *big.Int, proof *OracleProof) error {
	return nil
}

func (p *TestOracleProvider) GetProviderInfo() *ProviderInfo {
	return &ProviderInfo{
		Name:        p.name,
		Description: "test oracle provider",
		URL:         "",
		PublicKey:   nil,
		Active:      true,
		LastUpdate:  time.Now(),
		Reliability: 1.0,
	}
}

// Test helpers expected by tests
func (p *TestOracleProvider) SetMockPrice(asset string, v *big.Int, confidence uint8) {
	if v == nil {
		v = big.NewInt(0)
	}
	p.prices[asset] = &PriceData{
		Asset:      asset,
		Price:      new(big.Int).Set(v),
		Timestamp:  time.Now(),
		Provider:   p.name,
		Confidence: confidence,
		Source:     "test",
	}
}

func (p *TestOracleProvider) GetMockPrice(asset string) *PriceData {
	if pd, ok := p.prices[asset]; ok {
		cp := *pd
		if pd.Price != nil {
			cp.Price = new(big.Int).Set(pd.Price)
		}
		return &cp
	}
	return &PriceData{Asset: asset, Price: big.NewInt(0), Provider: p.name, Timestamp: time.Now()}
}

func (p *TestOracleProvider) GetStats() (uint64, uint64, uint64) {
	return p.reqs, p.succ, p.fails
}
