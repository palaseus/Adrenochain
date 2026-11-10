package utxo

import (
	"github.com/palaseus/adrenochain/pkg/block"
	"github.com/palaseus/adrenochain/pkg/crypto_utils"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestBlockSequentialSpends(t *testing.T) {
	ctu := crypto_utils.NewCryptoTestUtils(t)
	alice := ctu.GenerateTestKeyPair()
	us := NewUTXOSet()
	funding := createTestUTXO("sequential-spend", 0, 10000, alice, false, 0)
	us.AddUTXOSafe(funding)
	keys := map[string]*crypto_utils.TestKeyPair{alice.Address: alice}
	spend := func(hash []byte, value uint64) *block.Transaction {
		return ctu.CreateSignedTransaction([]*block.TxInput{{PrevTxHash: hash, Sequence: 0xffffffff}}, []*block.TxOutput{{Value: value, ScriptPubKey: funding.ScriptPubKey}}, keys, 100)
	}
	first := spend(funding.TxHash, 9900)
	second := spend(first.Hash, 9800)
	coinbase := block.NewTransaction(nil, []*block.TxOutput{{Value: 1000, ScriptPubKey: []byte("miner")}}, 0)
	b := &block.Block{Header: &block.Header{Height: 1}, Transactions: []*block.Transaction{coinbase, first, second}}
	require.NoError(t, us.ValidateBlockTransactions(b))
	require.NotNil(t, us.GetUTXO(funding.TxHash, 0), "validation must not mutate state")
	doubleSpend := spend(funding.TxHash, 9800)
	b.Transactions[2] = doubleSpend
	require.Error(t, us.ValidateBlockTransactions(b))
	b.Transactions[2] = coinbase
	require.Error(t, us.ValidateBlockTransactions(b))
}

func TestAmountsRejectOverflowAndNil(t *testing.T) {
	us := NewUTXOSet()
	tx := block.NewTransaction(nil, []*block.TxOutput{{Value: math.MaxUint64, ScriptPubKey: []byte{1}}, {Value: 1, ScriptPubKey: []byte{2}}}, 0)
	require.ErrorContains(t, us.ValidateTransaction(tx), "overflow")
	require.ErrorContains(t, us.ValidateTransactionBusinessLogic(tx), "overflow")
	_, err := us.CalculateFee(tx)
	require.Error(t, err)
	tx.Outputs = []*block.TxOutput{nil}
	require.Error(t, us.ValidateTransaction(tx))
	tx.Outputs = []*block.TxOutput{{Value: 1000, ScriptPubKey: []byte{1}}}
	tx.Inputs = []*block.TxInput{nil}
	require.Error(t, us.ValidateTransaction(tx))
}

func TestReplacingOutpointPreservesBalances(t *testing.T) {
	us := NewUTXOSet()
	us.AddUTXOSafe(NewUTXO([]byte{1}, 0, 100, []byte{1}, "alice", false, 0))
	us.AddUTXOSafe(NewUTXO([]byte{1}, 0, 200, []byte{2}, "bob", false, 0))
	require.Zero(t, us.GetBalance("alice"))
	require.EqualValues(t, 200, us.GetBalance("bob"))
	require.Equal(t, 1, us.GetUTXOCount())
}

func TestUTXOOwnershipIsIsolated(t *testing.T) {
	us := NewUTXOSet()
	original := NewUTXO([]byte{1}, 0, 100, []byte{1}, "alice", false, 0)
	us.AddUTXO(original)
	original.Value = 200
	original.ScriptPubKey[0] = 2
	copy := us.GetUTXO([]byte{1}, 0)
	require.EqualValues(t, 100, copy.Value)
	require.Equal(t, []byte{1}, copy.ScriptPubKey)
	copy.Value = 300
	copy.TxHash[0] = 9
	require.EqualValues(t, 100, us.GetBalance("alice"))
	require.NotNil(t, us.RemoveUTXO([]byte{1}, 0))
	require.Zero(t, us.GetBalance("alice"))
}

func TestBlockApplicationRollsBackOnBalanceOverflow(t *testing.T) {
	us := NewUTXOSet()
	us.AddUTXO(NewUTXO([]byte{1}, 0, math.MaxUint64, []byte{1}, "01", false, 0))
	initial := us.GetStats()
	first := block.NewTransaction(nil, []*block.TxOutput{{Value: 1000, ScriptPubKey: []byte{2}}}, 0)
	second := block.NewTransaction(nil, []*block.TxOutput{{Value: 1, ScriptPubKey: []byte{1}}}, 0)
	err := us.ProcessBlock(&block.Block{Header: &block.Header{Height: 1}, Transactions: []*block.Transaction{first, second}})
	require.ErrorContains(t, err, "overflow")
	require.Equal(t, initial, us.GetStats())
	require.Nil(t, us.GetUTXO(first.Hash, 0))
	require.EqualValues(t, uint64(math.MaxUint64), us.GetBalance("01"))
}
