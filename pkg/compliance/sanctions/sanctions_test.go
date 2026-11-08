package sanctions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScreenWalletExact(t *testing.T) {
	sc := NewSanctionsChecker()
	_ = sc.LoadList(ListOFACSDN, []Entry{
		{ID: "1", List: ListOFACSDN, Name: "Bad Actor", WalletAddresses: []string{"0xbad"}},
		{ID: "2", List: ListUNSC, Name: "Other Bad", WalletAddresses: []string{"0xother"}},
	})

	matches, err := sc.ScreenWallet("0xbad")
	assert.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Equal(t, ListOFACSDN, matches[0].Entry.List)
	assert.Equal(t, "wallet_address", matches[0].Field)

	// Case-insensitive.
	matches, _ = sc.ScreenWallet("0xBAD")
	assert.Len(t, matches, 1)

	// No match.
	matches, _ = sc.ScreenWallet("0xgood")
	assert.Len(t, matches, 0)
}

func TestScreenName(t *testing.T) {
	sc := NewSanctionsChecker()
	_ = sc.LoadList(ListUNSC, []Entry{
		{ID: "1", List: ListUNSC, Name: "Terrorist Corp", Aliases: []string{"TC Corp"}},
	})

	matches, _ := sc.ScreenName("Terrorist Corp")
	assert.Len(t, matches, 1)

	matches, _ = sc.ScreenName("tc corp")
	assert.Len(t, matches, 1)

	// Substring: query contains sanctioned name.
	matches, _ = sc.ScreenName("doing business with Terrorist Corp today")
	assert.Len(t, matches, 1)

	// No match.
	matches, _ = sc.ScreenName("Legitimate Business")
	assert.Len(t, matches, 0)
}

func TestScreenAddress(t *testing.T) {
	sc := NewSanctionsChecker()
	_ = sc.LoadList(ListEU, []Entry{
		{ID: "1", List: ListEU, Name: "Sanctioned Entity", Addresses: []string{"123 Bad Street, Badtown"}},
	})

	matches, _ := sc.ScreenAddress("123 Bad Street, Badtown, Country")
	assert.Len(t, matches, 1)

	matches, _ = sc.ScreenAddress("123 Good Street, Goodtown")
	assert.Len(t, matches, 0)
}

func TestEmptyQueries(t *testing.T) {
	sc := NewSanctionsChecker()
	_, err := sc.ScreenWallet("")
	assert.Error(t, err)
	_, err = sc.ScreenName("")
	assert.Error(t, err)
	_, err = sc.ScreenAddress("")
	assert.Error(t, err)
}
