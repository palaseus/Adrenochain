package hd

import "fmt"

// Purpose is the BIP-44 purpose level. 44' = BIP-44 accounts.
const PurposeBIP44 = 0x80000000 + 44

// CoinType is a SLIP-44 registered coin type.
type CoinType uint32

// SLIP-44 coin types (https://github.com/satoshilabs/slips/blob/master/slip-0044.md).
const (
	CoinBitcoin   CoinType = 0x80000000 + 0    // Bitcoin
	CoinLitecoin  CoinType = 0x80000000 + 2    // Litecoin
	CoinEthereum  CoinType = 0x80000000 + 60   // Ethereum
	CoinEthereumC CoinType = 0x80000000 + 61   // Ethereum Classic
	CoinXRP       CoinType = 0x80000000 + 144  // XRP
	CoinCardano   CoinType = 0x80000000 + 1815 // Cardano
	CoinPolkadot  CoinType = 0x80000000 + 354  // Polkadot
	CoinSolana    CoinType = 0x80000000 + 501  // Solana
	CoinDogecoin  CoinType = 0x80000000 + 3    // Dogecoin
	CoinDash      CoinType = 0x80000000 + 5    // Dash
	CoinZcash     CoinType = 0x80000000 + 133  // Zcash
)

// String returns the coin name.
func (c CoinType) String() string {
	names := map[CoinType]string{
		CoinBitcoin:   "Bitcoin",
		CoinLitecoin:  "Litecoin",
		CoinEthereum:  "Ethereum",
		CoinEthereumC: "Ethereum Classic",
		CoinXRP:       "XRP",
		CoinCardano:   "Cardano",
		CoinPolkadot:  "Polkadot",
		CoinSolana:    "Solana",
		CoinDogecoin:  "Dogecoin",
		CoinDash:      "Dash",
		CoinZcash:     "Zcash",
	}
	if n, ok := names[c]; ok {
		return n
	}
	return fmt.Sprintf("coin-%d", uint32(c)&0x7FFFFFFF)
}

// AccountPath builds a BIP-44 derivation path:
//
//	m / purpose' / coin_type' / account' / change / address_index
//
// change: 0 = external (receiving), 1 = internal (change).
func AccountPath(coin CoinType, account uint32, change uint32, addressIndex uint32) string {
	return fmt.Sprintf("m/%d'/%d'/%d'/%d/%d",
		PurposeBIP44-hardenedStart, // strip hardened bit for display (44)
		uint32(coin)&0x7FFFFFFF,    // strip hardened bit for display
		account,                    // account index (displayed without hardened bit)
		change,
		addressIndex,
	)
}

// ParsePathSplit splits a BIP-32 path into its components for display.
func ParsePathSplit(path string) ([]string, error) {
	if len(path) < 2 || path[0] != 'm' || path[1] != '/' {
		return nil, fmt.Errorf("invalid path %q", path)
	}
	return splitPath(path[2:]), nil
}

// IsHardened reports whether a path component string denotes a hardened index.
func IsHardened(comp string) bool {
	return len(comp) > 0 && (comp[len(comp)-1] == '\'' || comp[len(comp)-1] == 'h' || comp[len(comp)-1] == 'H')
}
