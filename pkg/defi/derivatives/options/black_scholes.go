package options

import (
	"math"
	"math/big"
	"errors"
)

// BlackScholesModel represents the Black-Scholes options pricing model
type BlackScholesModel struct {
	// Risk-free interest rate (annualized)
	RiskFreeRate *big.Float
	// Volatility (annualized)
	Volatility *big.Float
	// Time to expiration (in years)
	TimeToExpiry *big.Float
}

// OptionType represents the type of option
type OptionType int

const (
	Call OptionType = iota
	Put
)

// Option represents a financial option contract
type Option struct {
	Type          OptionType
	StrikePrice   *big.Float
	CurrentPrice  *big.Float
	TimeToExpiry  *big.Float
	RiskFreeRate  *big.Float
	Volatility    *big.Float
}

// NewBlackScholesModel creates a new Black-Scholes model instance
func NewBlackScholesModel(riskFreeRate, volatility, timeToExpiry *big.Float) (*BlackScholesModel, error) {
	if riskFreeRate == nil || volatility == nil || timeToExpiry == nil {
		return nil, errors.New("all parameters must be non-nil")
	}
	
	// Validate parameters
	if riskFreeRate.Sign() < 0 {
		return nil, errors.New("risk-free rate must be non-negative")
	}
	if volatility.Sign() <= 0 {
		return nil, errors.New("volatility must be positive")
	}
	if timeToExpiry.Sign() <= 0 {
		return nil, errors.New("time to expiry must be positive")
	}
	
	return &BlackScholesModel{
		RiskFreeRate:  new(big.Float).Copy(riskFreeRate),
		Volatility:    new(big.Float).Copy(volatility),
		TimeToExpiry:  new(big.Float).Copy(timeToExpiry),
	}, nil
}

// NewOption creates a new option contract
func NewOption(optionType OptionType, strikePrice, currentPrice, timeToExpiry, riskFreeRate, volatility *big.Float) (*Option, error) {
	if strikePrice == nil || currentPrice == nil || timeToExpiry == nil || riskFreeRate == nil || volatility == nil {
		return nil, errors.New("all parameters must be non-nil")
	}
	
	// Validate parameters
	if strikePrice.Sign() <= 0 {
		return nil, errors.New("strike price must be positive")
	}
	if currentPrice.Sign() <= 0 {
		return nil, errors.New("current price must be positive")
	}
	if timeToExpiry.Sign() <= 0 {
		return nil, errors.New("time to expiry must be positive")
	}
	if riskFreeRate.Sign() < 0 {
		return nil, errors.New("risk-free rate must be non-negative")
	}
	if volatility.Sign() <= 0 {
		return nil, errors.New("volatility must be positive")
	}
	
	return &Option{
		Type:         optionType,
		StrikePrice:  new(big.Float).Copy(strikePrice),
		CurrentPrice: new(big.Float).Copy(currentPrice),
		TimeToExpiry: new(big.Float).Copy(timeToExpiry),
		RiskFreeRate: new(big.Float).Copy(riskFreeRate),
		Volatility:   new(big.Float).Copy(volatility),
	}, nil
}

// Price calculates the option price using the Black-Scholes model
func (bs *BlackScholesModel) Price(option *Option) (*big.Float, error) {
	if option == nil {
		return nil, errors.New("option cannot be nil")
	}
	
	// Convert big.Float to float64 for calculations
	S := option.CurrentPrice
	K := option.StrikePrice
	T := option.TimeToExpiry
	r := option.RiskFreeRate
	sigma := option.Volatility
	
	// Calculate d1 and d2
	d1, d2, err := bs.calculateD1D2(S, K, T, r, sigma)
	if err != nil {
		return nil, err
	}
	
	// Calculate option price based on type
	var price *big.Float
	if option.Type == Call {
		price = bs.callPrice(S, K, T, r, d1, d2)
	} else {
		price = bs.putPrice(S, K, T, r, d1, d2)
	}
	
	return price, nil
}

// calculateD1D2 calculates the d1 and d2 parameters for Black-Scholes
func (bs *BlackScholesModel) calculateD1D2(S, K, T, r, sigma *big.Float) (*big.Float, *big.Float, error) {
	// d1 = (ln(S/K) + (r + σ²/2)T) / (σ√T)
	// d2 = d1 - σ√T
	
	// Convert to float64 for mathematical operations
	s, _ := S.Float64()
	k, _ := K.Float64()
	t, _ := T.Float64()
	rate, _ := r.Float64()
	vol, _ := sigma.Float64()
	
	if s <= 0 || k <= 0 || t <= 0 || vol <= 0 {
		return nil, nil, errors.New("invalid parameters for d1/d2 calculation")
	}
	
	// Calculate ln(S/K)
	logSK := math.Log(s / k)
	
	// Calculate (r + σ²/2)T
	sigmaSquared := vol * vol
	halfSigmaSquared := sigmaSquared / 2.0
	ratePlusHalfSigma := rate + halfSigmaSquared
	rateTerm := ratePlusHalfSigma * t
	
	// Calculate σ√T
	sigmaSqrtT := vol * math.Sqrt(t)
	
	// Calculate d1
	d1 := (logSK + rateTerm) / sigmaSqrtT
	
	// Calculate d2
	d2 := d1 - sigmaSqrtT
	
	// Convert back to big.Float
	d1Big := new(big.Float).SetFloat64(d1)
	d2Big := new(big.Float).SetFloat64(d2)
	
	return d1Big, d2Big, nil
}

// callPrice calculates the call option price
func (bs *BlackScholesModel) callPrice(S, K, T, r *big.Float, d1, d2 *big.Float) *big.Float {
	// C = S*N(d1) - K*e^(-rT)*N(d2)
	
	// Convert to float64 for calculations
	s, _ := S.Float64()
	k, _ := K.Float64()
	t, _ := T.Float64()
	rate, _ := r.Float64()
	d1Val, _ := d1.Float64()
	d2Val, _ := d2.Float64()
	
	// Calculate N(d1) and N(d2) using standard normal CDF
	N1 := normalCDF(d1Val)
	N2 := normalCDF(d2Val)
	
	// Calculate K*e^(-rT)
	discountFactor := math.Exp(-rate * t)
	discountedStrike := k * discountFactor
	
	// Calculate call price
	callPrice := s*N1 - discountedStrike*N2
	
	return new(big.Float).SetFloat64(callPrice)
}

// putPrice calculates the put option price
func (bs *BlackScholesModel) putPrice(S, K, T, r *big.Float, d1, d2 *big.Float) *big.Float {
	// P = K*e^(-rT)*N(-d2) - S*N(-d1)
	
	// Convert to float64 for calculations
	s, _ := S.Float64()
	k, _ := K.Float64()
	t, _ := T.Float64()
	rate, _ := r.Float64()
	d1Val, _ := d1.Float64()
	d2Val, _ := d2.Float64()
	
	// Calculate N(-d1) and N(-d2)
	N1Neg := normalCDF(-d1Val)
	N2Neg := normalCDF(-d2Val)
	
	// Calculate K*e^(-rT)
	discountFactor := math.Exp(-rate * t)
	discountedStrike := k * discountFactor
	
	// Calculate put price
	putPrice := discountedStrike*N2Neg - s*N1Neg
	
	return new(big.Float).SetFloat64(putPrice)
}

// normalCDF calculates the cumulative distribution function of the standard normal distribution
func normalCDF(x float64) float64 {
	// Use approximation for standard normal CDF
	// This is a high-precision approximation
	a1 := 0.254829592
	a2 := -0.284496736
	a3 := 1.421413741
	a4 := -1.453152027
	a5 := 1.061405429
	p := 0.3275911
	
	// Save the sign of x
	sign := 1.0
	if x < 0 {
		sign = -1.0
	}
	x = math.Abs(x) / math.Sqrt(2.0)
	
	// A&S formula 7.1.26
	t := 1.0 / (1.0 + p*x)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)
	
	return 0.5 * (1.0 + sign*y)
}

// Delta calculates the option delta (first derivative of price with respect to underlying price)
func (bs *BlackScholesModel) Delta(option *Option) (*big.Float, error) {
	if option == nil {
		return nil, errors.New("option cannot be nil")
	}
	
	// Calculate d1
	d1, _, err := bs.calculateD1D2(option.CurrentPrice, option.StrikePrice, option.TimeToExpiry, option.RiskFreeRate, option.Volatility)
	if err != nil {
		return nil, err
	}
	
	d1Val, _ := d1.Float64()
	
	var delta float64
	if option.Type == Call {
		delta = normalCDF(d1Val)
	} else {
		delta = normalCDF(d1Val) - 1.0
	}
	
	return new(big.Float).SetFloat64(delta), nil
}

// Gamma calculates the option gamma (second derivative of price with respect to underlying price)
func (bs *BlackScholesModel) Gamma(option *Option) (*big.Float, error) {
	if option == nil {
		return nil, errors.New("option cannot be nil")
	}
	
	// Convert to float64 for calculations
	s, _ := option.CurrentPrice.Float64()
	sigma, _ := option.Volatility.Float64()
	t, _ := option.TimeToExpiry.Float64()
	
	// Calculate d1
	d1, _, err := bs.calculateD1D2(option.CurrentPrice, option.StrikePrice, option.TimeToExpiry, option.RiskFreeRate, option.Volatility)
	if err != nil {
		return nil, err
	}
	
	d1Val, _ := d1.Float64()
	
	// Gamma = N'(d1) / (S * σ * √T)
	// N'(d1) is the standard normal PDF
	normalPDF := math.Exp(-0.5*d1Val*d1Val) / math.Sqrt(2.0*math.Pi)
	gamma := normalPDF / (s * sigma * math.Sqrt(t))
	
	return new(big.Float).SetFloat64(gamma), nil
}

// Theta calculates the option theta (first derivative of price with respect to time)
func (bs *BlackScholesModel) Theta(option *Option) (*big.Float, error) {
	if option == nil {
		return nil, errors.New("option cannot be nil")
	}
	
	// Convert to float64 for calculations
	s, _ := option.CurrentPrice.Float64()
	k, _ := option.StrikePrice.Float64()
	t, _ := option.TimeToExpiry.Float64()
	r, _ := option.RiskFreeRate.Float64()
	sigma, _ := option.Volatility.Float64()
	
	// Calculate d1 and d2
	d1, d2, err := bs.calculateD1D2(option.CurrentPrice, option.StrikePrice, option.TimeToExpiry, option.RiskFreeRate, option.Volatility)
	if err != nil {
		return nil, err
	}
	
	d1Val, _ := d1.Float64()
	d2Val, _ := d2.Float64()
	
	// Theta calculation
	var theta float64
	if option.Type == Call {
		theta = -(s*sigma*math.Exp(-0.5*d1Val*d1Val))/(2*math.Sqrt(2*math.Pi*t)) - r*k*math.Exp(-r*t)*normalCDF(d2Val)
	} else {
		theta = -(s*sigma*math.Exp(-0.5*d1Val*d1Val))/(2*math.Sqrt(2*math.Pi*t)) + r*k*math.Exp(-r*t)*normalCDF(-d2Val)
	}
	
	return new(big.Float).SetFloat64(theta), nil
}

// Vega calculates the option vega (first derivative of price with respect to volatility)
func (bs *BlackScholesModel) Vega(option *Option) (*big.Float, error) {
	if option == nil {
		return nil, errors.New("option cannot be nil")
	}
	
	// Convert to float64 for calculations
	s, _ := option.CurrentPrice.Float64()
	t, _ := option.TimeToExpiry.Float64()
	
	// Calculate d1
	d1, _, err := bs.calculateD1D2(option.CurrentPrice, option.StrikePrice, option.TimeToExpiry, option.RiskFreeRate, option.Volatility)
	if err != nil {
		return nil, err
	}
	
	d1Val, _ := d1.Float64()
	
	// Vega = S * √T * N'(d1)
	normalPDF := math.Exp(-0.5*d1Val*d1Val) / math.Sqrt(2.0*math.Pi)
	vega := s * math.Sqrt(t) * normalPDF
	
	return new(big.Float).SetFloat64(vega), nil
}

// Rho calculates the option rho (first derivative of price with respect to risk-free rate)
func (bs *BlackScholesModel) Rho(option *Option) (*big.Float, error) {
	if option == nil {
		return nil, errors.New("option cannot be nil")
	}
	
	// Convert to float64 for calculations
	k, _ := option.StrikePrice.Float64()
	t, _ := option.TimeToExpiry.Float64()
	r, _ := option.RiskFreeRate.Float64()
	
	// Calculate d2
	_, d2, err := bs.calculateD1D2(option.CurrentPrice, option.StrikePrice, option.TimeToExpiry, option.RiskFreeRate, option.Volatility)
	if err != nil {
		return nil, err
	}
	
	d2Val, _ := d2.Float64()
	
	var rho float64
	if option.Type == Call {
		rho = k * t * math.Exp(-r*t) * normalCDF(d2Val)
	} else {
		rho = -k * t * math.Exp(-r*t) * normalCDF(-d2Val)
	}
	
	return new(big.Float).SetFloat64(rho), nil
}
