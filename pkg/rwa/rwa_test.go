package rwa

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func id(b byte) Identity {
	var i Identity
	i[0] = b
	return i
}

func TestIdentityRegistry(t *testing.T) {
	reg := NewIdentityRegistry()
	issuer := id(1)
	reg.RegisterIssuer(issuer)
	assert.True(t, reg.IsRegisteredIssuer(issuer))

	owner := id(2)
	reg.AddClaim(owner, &IdentityClaim{
		Topic:     ClaimTopicKYC,
		Issuer:    issuer,
		Value:     []byte("kyc-1"),
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(time.Hour),
	})

	assert.True(t, reg.HasClaimTopic(owner, ClaimTopicKYC))
	assert.False(t, reg.HasClaimTopic(owner, ClaimTopicAccreditedInvestor))

	// Expired claim on a different topic (does not replace KYC).
	reg.AddClaim(owner, &IdentityClaim{
		Topic:     ClaimTopicAccreditedInvestor,
		Issuer:    issuer,
		Value:     []byte("accredited-expired"),
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(-time.Minute),
	})
	// The valid KYC claim still exists; expired accredited claim is filtered.
	claims := reg.GetClaims(owner)
	assert.GreaterOrEqual(t, len(claims), 1)
	hasKYC := false
	for _, c := range claims {
		if c.Topic == ClaimTopicKYC {
			hasKYC = true
		}
	}
	assert.True(t, hasKYC)

	// Unregistered issuer cannot add claims.
	unreg := id(9)
	err := reg.AddClaim(owner, &IdentityClaim{
		Topic:    ClaimTopicKYC,
		Issuer:   unreg,
		Value:    []byte("x"),
		IssuedAt: time.Now(),
	})
	assert.Error(t, err)
}

func TestTokenLifecycle(t *testing.T) {
	reg := NewIdentityRegistry()
	issuer := id(1)
	reg.RegisterIssuer(issuer)

	owner := id(2)
	token := NewToken("TKA", "Test RWA Asset", 18, owner, reg)
	assert.Equal(t, TokenStatusDraft, token.Status)
	assert.Equal(t, "TKA", token.Symbol)

	// Non-owner cannot issue.
	assert.Error(t, token.Issue(id(3)))
	// Owner issues.
	assert.NoError(t, token.Issue(owner))
	assert.Equal(t, TokenStatusIssued, token.Status)

	// Mint.
	user := id(5)
	reg.AddClaim(user, &IdentityClaim{
		Topic:    ClaimTopicKYC,
		Issuer:   issuer,
		Value:    []byte("kyc-user"),
		IssuedAt: time.Now(),
	})
	reg.AddClaim(user, &IdentityClaim{
		Topic:    ClaimTopicSanctionsClear,
		Issuer:   issuer,
		Value:    []byte("sanctions-clear"),
		IssuedAt: time.Now(),
	})

	assert.NoError(t, token.Mint(user, big.NewInt(1000), owner))
	assert.Equal(t, big.NewInt(1000), token.BalanceOf(user))
	assert.Equal(t, big.NewInt(1000), token.TotalSupplyOf())

	// Pause blocks transfers.
	assert.NoError(t, token.Pause(owner))
	assert.Equal(t, TokenStatusPaused, token.Status)
	err := token.Transfer(user, id(6), big.NewInt(100), user)
	assert.Error(t, err)
	assert.NoError(t, token.Unpause(owner))
	assert.Equal(t, TokenStatusIssued, token.Status)
}

func TestTransferCompliance(t *testing.T) {
	reg := NewIdentityRegistry()
	issuer := id(1)
	reg.RegisterIssuer(issuer)

	owner := id(2)
	token := NewToken("TKB", "Test RWA B", 18, owner, reg)
	assert.NoError(t, token.Issue(owner))

	from := id(3)
	to := id(4)
	// Both parties need KYC + sanctions clear.
	for _, iden := range []Identity{from, to} {
		reg.AddClaim(iden, &IdentityClaim{
			Topic:    ClaimTopicKYC,
			Issuer:   issuer,
			Value:    []byte("kyc"),
			IssuedAt: time.Now(),
		})
		reg.AddClaim(iden, &IdentityClaim{
			Topic:    ClaimTopicSanctionsClear,
			Issuer:   issuer,
			Value:    []byte("clear"),
			IssuedAt: time.Now(),
		})
	}

	assert.NoError(t, token.Mint(from, big.NewInt(1000), owner))

	// Compliant transfer.
	assert.NoError(t, token.Transfer(from, to, big.NewInt(100), from))
	assert.Equal(t, big.NewInt(900), token.BalanceOf(from))
	assert.Equal(t, big.NewInt(100), token.BalanceOf(to))

	// Receiver without KYC cannot receive.
	unverified := id(7)
	assert.Error(t, token.Transfer(from, unverified, big.NewInt(10), from))
}

func TestInsufficientBalance(t *testing.T) {
	reg := NewIdentityRegistry()
	issuer := id(1)
	reg.RegisterIssuer(issuer)

	owner := id(2)
	token := NewToken("TKC", "Test RWA C", 18, owner, reg)
	assert.NoError(t, token.Issue(owner))

	from := id(3)
	to := id(4)
	for _, iden := range []Identity{from, to} {
		reg.AddClaim(iden, &IdentityClaim{
			Topic:    ClaimTopicKYC,
			Issuer:   issuer,
			Value:    []byte("kyc"),
			IssuedAt: time.Now(),
		})
		reg.AddClaim(iden, &IdentityClaim{
			Topic:    ClaimTopicSanctionsClear,
			Issuer:   issuer,
			Value:    []byte("clear"),
			IssuedAt: time.Now(),
		})
	}

	assert.NoError(t, token.Mint(from, big.NewInt(10), owner))
	assert.Error(t, token.Transfer(from, to, big.NewInt(100), from))
}

func TestProofOfReserve(t *testing.T) {
	supply := big.NewInt(1000000)
	por := &ProofOfReserve{
		Auditor:         id(1),
		TokenSupply:     supply,
		ReserveValue:    supply,
		AttestedAt:      time.Now(),
		ReportReference: "ipfs://Qmreserve",
	}
	assert.True(t, por.IsValid())

	por2 := &ProofOfReserve{
		TokenSupply:  supply,
		ReserveValue: big.NewInt(999999),
	}
	assert.False(t, por2.IsValid())
	assert.False(t, (&ProofOfReserve{}).IsValid())
}

func TestAssetTypes(t *testing.T) {
	assert.Equal(t, "treasury", AssetTypeTreasury.String())
	assert.Equal(t, "real_estate", AssetTypeRealEstate.String())
	assert.Equal(t, "private_credit", AssetTypePrivateCredit.String())
	assert.Equal(t, "commodity", AssetTypeCommodity.String())
	assert.Equal(t, "equity", AssetTypeEquity.String())
	assert.Equal(t, "fund", AssetTypeFund.String())
	assert.Equal(t, "asset_99", AssetType(99).String())
}

func TestIdentityString(t *testing.T) {
	assert.Equal(t, "kyc", ClaimTopicKYC.String())
	assert.Equal(t, "accredited_investor", ClaimTopicAccreditedInvestor.String())
	assert.Equal(t, "sanctions_clear", ClaimTopicSanctionsClear.String())
	assert.Equal(t, "topic_99", ClaimTopic(99).String())
}
