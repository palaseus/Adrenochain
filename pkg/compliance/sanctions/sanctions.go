// Package sanctions implements sanctions screening for Adrenochain, aligned with
// FATF Recommendation 15 (risk-based approach) and major sanctions regimes as of
// November 2026:
//
//   - OFAC (US) SDN, SSI, and non-SDN lists
//   - UN Security Council Consolidated List
//   - EU Consolidated List of persons, groups and entities
//   - HM Treasury (UK) Consolidated List
//   - VARA (UAE) designated persons
//   - MAS (Singapore) designated entities
//
// Screening is deterministic and offline: load a sanctions list into a
// SanctionsChecker and test names, addresses, and wallet addresses against it.
// See pkg/compliance/sanctions/README.md for loading guidance.

package sanctions

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// ListType identifies a sanctions regime.
type ListType int

const (
	// ListOFACSDN is the US OFAC Specially Designated Nationals list.
	ListOFACSDN ListType = iota
	// ListUNSC is the UN Security Council Consolidated List.
	ListUNSC
	// ListEU is the EU Consolidated List.
	ListEU
	// ListUK is the HM Treasury (UK) Consolidated List.
	ListUK
	// ListVARA is the VARA (UAE) designated persons list.
	ListVARA
	// ListMAS is the MAS (Singapore) designated entities list.
	ListMAS
)

// String returns the regime name.
func (l ListType) String() string {
	names := map[ListType]string{
		ListOFACSDN: "OFAC SDN",
		ListUNSC:    "UN Security Council",
		ListEU:      "EU Consolidated",
		ListUK:      "UK HM Treasury",
		ListVARA:    "VARA (UAE)",
		ListMAS:     "MAS (Singapore)",
	}
	if n, ok := names[l]; ok {
		return n
	}
	return "Unknown"
}

// Entry is a single sanctioned-party record.
type Entry struct {
	// ID is the unique entry identifier within its list.
	ID string
	// List is the sanctions regime this entry belongs to.
	List ListType
	// Name is the sanctioned party name.
	Name string
	// Aliases are known alternate names.
	Aliases []string
	// Addresses are known addresses associated with the party.
	Addresses []string
	// WalletAddresses are known crypto wallet addresses.
	WalletAddresses []string
	// Program is the sanctions program (e.g. "UKRAINE-RUSSIA", "SDGT").
	Program string
}

// Match records why an entry matched a screening query.
type Match struct {
	Entry Entry
	Field string  // which field matched: name, alias, address, wallet
	Query string  // the value that was screened
	Score float64 // fuzzy match score (0-1)
}

// SanctionsChecker screens queries against loaded sanctions lists.
type SanctionsChecker struct {
	mu       sync.RWMutex
	lists    map[ListType][]Entry
	byWallet map[string][]Entry // wallet address -> entries
	byName   map[string][]Entry // lowercase name -> entries
}

// NewSanctionsChecker creates an empty sanctions checker.
func NewSanctionsChecker() *SanctionsChecker {
	return &SanctionsChecker{
		lists:    make(map[ListType][]Entry),
		byWallet: make(map[string][]Entry),
		byName:   make(map[string][]Entry),
	}
}

// LoadList loads or replaces a sanctions list.
func (sc *SanctionsChecker) LoadList(list ListType, entries []Entry) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.lists[list] = entries
	sc.rebuildIndexesUnlocked()
	return nil
}

// rebuildIndexesUnlocked rebuilds lookup indexes (caller must hold lock).
func (sc *SanctionsChecker) rebuildIndexesUnlocked() {
	sc.byWallet = make(map[string][]Entry)
	sc.byName = make(map[string][]Entry)
	for _, entries := range sc.lists {
		for _, e := range entries {
			for _, w := range e.WalletAddresses {
				sc.byWallet[strings.ToLower(w)] = append(sc.byWallet[strings.ToLower(w)], e)
			}
			nameKey := strings.ToLower(e.Name)
			sc.byName[nameKey] = append(sc.byName[nameKey], e)
			for _, a := range e.Aliases {
				sc.byName[strings.ToLower(a)] = append(sc.byName[strings.ToLower(a)], e)
			}
		}
	}
}

// ScreenWallet checks a wallet address against all loaded sanctions lists.
func (sc *SanctionsChecker) ScreenWallet(walletAddress string) ([]Match, error) {
	if walletAddress == "" {
		return nil, errors.New("empty wallet address")
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	key := strings.ToLower(walletAddress)
	entries, ok := sc.byWallet[key]
	if !ok {
		return nil, nil
	}

	matches := make([]Match, 0, len(entries))
	for _, e := range entries {
		matches = append(matches, Match{
			Entry: e,
			Field: "wallet_address",
			Query: walletAddress,
			Score: 1.0,
		})
	}
	return matches, nil
}

// ScreenName checks a name against all loaded sanctions lists using exact and
// substring matching.
func (sc *SanctionsChecker) ScreenName(name string) ([]Match, error) {
	if name == "" {
		return nil, errors.New("empty name")
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	lower := strings.ToLower(name)
	matched := make(map[string]bool) // entry ID -> matched

	var matches []Match
	for key, entries := range sc.byName {
		// Exact match.
		if key == lower {
			for _, e := range entries {
				if !matched[e.ID] {
					matched[e.ID] = true
					matches = append(matches, Match{
						Entry: e,
						Field: "name",
						Query: name,
						Score: 1.0,
					})
				}
			}
			continue
		}
		// Substring match (name contains sanctioned name or vice versa).
		if strings.Contains(lower, key) || strings.Contains(key, lower) {
			for _, e := range entries {
				if !matched[e.ID] {
					matched[e.ID] = true
					matches = append(matches, Match{
						Entry: e,
						Field: "name",
						Query: name,
						Score: 0.8,
					})
				}
			}
		}
	}
	return matches, nil
}

// ScreenAddress checks a physical address against sanctions lists.
func (sc *SanctionsChecker) ScreenAddress(address string) ([]Match, error) {
	if address == "" {
		return nil, errors.New("empty address")
	}
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	lower := strings.ToLower(address)
	var matches []Match
	for _, entries := range sc.lists {
		for _, e := range entries {
			for _, a := range e.Addresses {
				if strings.Contains(lower, strings.ToLower(a)) {
					matches = append(matches, Match{
						Entry: e,
						Field: "address",
						Query: address,
						Score: 0.7,
					})
				}
			}
		}
	}
	return matches, nil
}

// ListCount returns the number of entries loaded for a list.
func (sc *SanctionsChecker) ListCount(list ListType) int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return len(sc.lists[list])
}

// Ensure fmt is used.
var _ = fmt.Sprintf
