package block

import (
	"encoding/binary"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestDeserializeRejectsImpossibleCounts(t *testing.T) {
	txData := make([]byte, 60)
	binary.BigEndian.PutUint32(txData[4:8], math.MaxUint32)
	require.Error(t, new(Transaction).Deserialize(txData))
	clear(txData)
	binary.BigEndian.PutUint32(txData[8:12], math.MaxUint32)
	require.Error(t, new(Transaction).Deserialize(txData))
	b := NewBlock(make([]byte, 32), 1, 1)
	data, err := b.Serialize()
	require.NoError(t, err)
	binary.BigEndian.PutUint32(data[104:108], math.MaxUint32)
	require.Error(t, new(Block).Deserialize(data))
}

func TestValidationRejectsNilNestedObjects(t *testing.T) {
	b := NewBlock(make([]byte, 32), 1, 1)
	b.Transactions = []*Transaction{nil}
	require.Error(t, b.IsValid())
	tx := &Transaction{Version: 1, Hash: make([]byte, 32), Inputs: []*TxInput{nil}, Outputs: []*TxOutput{{Value: 1, ScriptPubKey: []byte{1}}}}
	require.Error(t, tx.IsValid())
	tx.Inputs = nil
	tx.Outputs = []*TxOutput{nil}
	require.Error(t, tx.IsValid())
}

func FuzzMalformedWireData(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 108))
	malicious := make([]byte, 108)
	binary.BigEndian.PutUint32(malicious[4:8], math.MaxUint32)
	f.Add(malicious)
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		_ = new(Block).Deserialize(data)
		_ = new(Transaction).Deserialize(data)
		_ = new(Header).Deserialize(data)
		_ = new(TxInput).Deserialize(data)
		_ = new(TxOutput).Deserialize(data)
	})
}
