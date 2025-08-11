package net

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestBlockMessage(t *testing.T) {
	t.Run("NewBlockMessage", func(t *testing.T) {
		msg := &BlockMessage{
			BlockData: []byte("test block data"),
		}
		assert.Equal(t, []byte("test block data"), msg.GetBlockData())
	})

	t.Run("BlockMessageReset", func(t *testing.T) {
		msg := &BlockMessage{
			BlockData: []byte("test data"),
		}
		msg.Reset()
		assert.Nil(t, msg.BlockData)
	})

	t.Run("BlockMessageString", func(t *testing.T) {
		msg := &BlockMessage{
			BlockData: []byte("test data"),
		}
		str := msg.String()
		assert.Contains(t, str, "block_data")
		assert.Contains(t, str, "test data")
	})

	t.Run("BlockMessageProtoMessage", func(t *testing.T) {
		msg := &BlockMessage{}
		msg.ProtoMessage() // Should not panic
	})

	t.Run("BlockMessageProtoReflect", func(t *testing.T) {
		msg := &BlockMessage{
			BlockData: []byte("test data"),
		}
		reflection := msg.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockMessageDescriptor", func(t *testing.T) {
		msg := &BlockMessage{}
		desc, _ := msg.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestTransactionMessage(t *testing.T) {
	t.Run("NewTransactionMessage", func(t *testing.T) {
		msg := &TransactionMessage{
			TransactionData: []byte("test transaction data"),
		}
		assert.Equal(t, []byte("test transaction data"), msg.GetTransactionData())
	})

	t.Run("TransactionMessageReset", func(t *testing.T) {
		msg := &TransactionMessage{
			TransactionData: []byte("test data"),
		}
		msg.Reset()
		assert.Nil(t, msg.TransactionData)
	})

	t.Run("TransactionMessageString", func(t *testing.T) {
		msg := &TransactionMessage{
			TransactionData: []byte("test data"),
		}
		str := msg.String()
		assert.Contains(t, str, "transaction_data")
	})

	t.Run("TransactionMessageProtoMessage", func(t *testing.T) {
		msg := &TransactionMessage{}
		msg.ProtoMessage() // Should not panic
	})

	t.Run("TransactionMessageProtoReflect", func(t *testing.T) {
		msg := &TransactionMessage{
			TransactionData: []byte("test data"),
		}
		reflection := msg.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("TransactionMessageDescriptor", func(t *testing.T) {
		msg := &TransactionMessage{}
		desc, _ := msg.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestBlockHeader(t *testing.T) {
	t.Run("NewBlockHeader", func(t *testing.T) {
		header := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     1234567890,
			Difficulty:    1000,
			Nonce:         42,
			Height:        100,
			Hash:          []byte("block_hash"),
		}

		assert.Equal(t, uint32(1), header.GetVersion())
		assert.Equal(t, []byte("prev_hash"), header.GetPrevBlockHash())
		assert.Equal(t, []byte("merkle_root"), header.GetMerkleRoot())
		assert.Equal(t, int64(1234567890), header.GetTimestamp())
		assert.Equal(t, uint64(1000), header.GetDifficulty())
		assert.Equal(t, uint64(42), header.GetNonce())
		assert.Equal(t, uint64(100), header.GetHeight())
		assert.Equal(t, []byte("block_hash"), header.GetHash())
	})

	t.Run("BlockHeaderReset", func(t *testing.T) {
		header := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     1234567890,
			Difficulty:    1000,
			Nonce:         42,
			Height:        100,
			Hash:          []byte("block_hash"),
		}
		header.Reset()
		assert.Equal(t, uint32(0), header.Version)
		assert.Nil(t, header.PrevBlockHash)
		assert.Nil(t, header.MerkleRoot)
		assert.Equal(t, int64(0), header.Timestamp)
		assert.Equal(t, uint64(0), header.Difficulty)
		assert.Equal(t, uint64(0), header.Nonce)
		assert.Equal(t, uint64(0), header.Height)
		assert.Nil(t, header.Hash)
	})

	t.Run("BlockHeaderString", func(t *testing.T) {
		header := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     1234567890,
			Difficulty:    1000,
			Nonce:         42,
			Height:        100,
			Hash:          []byte("block_hash"),
		}
		str := header.String()
		assert.Contains(t, str, "version")
	})

	t.Run("BlockHeaderProtoMessage", func(t *testing.T) {
		header := &BlockHeader{}
		header.ProtoMessage() // Should not panic
	})

	t.Run("BlockHeaderProtoReflect", func(t *testing.T) {
		header := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
		}
		reflection := header.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockHeaderDescriptor", func(t *testing.T) {
		header := &BlockHeader{}
		desc, _ := header.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestBlockHeadersRequest(t *testing.T) {
	t.Run("NewBlockHeadersRequest", func(t *testing.T) {
		req := &BlockHeadersRequest{
			StartHeight: 100,
			Count:       50,
			StopHash:    []byte("stop_hash"),
		}

		assert.Equal(t, uint64(100), req.GetStartHeight())
		assert.Equal(t, uint64(50), req.GetCount())
		assert.Equal(t, []byte("stop_hash"), req.GetStopHash())
	})

	t.Run("BlockHeadersRequestReset", func(t *testing.T) {
		req := &BlockHeadersRequest{
			StartHeight: 100,
			Count:       50,
			StopHash:    []byte("stop_hash"),
		}
		req.Reset()
		assert.Equal(t, uint64(0), req.StartHeight)
		assert.Equal(t, uint64(0), req.Count)
		assert.Nil(t, req.StopHash)
	})

	t.Run("BlockHeadersRequestString", func(t *testing.T) {
		req := &BlockHeadersRequest{
			StartHeight: 100,
			Count:       50,
		}
		str := req.String()
		assert.Contains(t, str, "start_height")
	})

	t.Run("BlockHeadersRequestProtoMessage", func(t *testing.T) {
		req := &BlockHeadersRequest{}
		req.ProtoMessage() // Should not panic
	})

	t.Run("BlockHeadersRequestProtoReflect", func(t *testing.T) {
		req := &BlockHeadersRequest{
			StartHeight: 100,
			Count:       50,
		}
		reflection := req.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockHeadersRequestDescriptor", func(t *testing.T) {
		req := &BlockHeadersRequest{}
		desc, _ := req.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestBlockHeadersResponse(t *testing.T) {
	t.Run("NewBlockHeadersResponse", func(t *testing.T) {
		headers := []*BlockHeader{
			{Version: 1, Height: 100},
			{Version: 1, Height: 101},
		}
		resp := &BlockHeadersResponse{
			Headers: headers,
			HasMore: true,
		}

		assert.Equal(t, headers, resp.GetHeaders())
		assert.Equal(t, true, resp.GetHasMore())
	})

	t.Run("BlockHeadersResponseReset", func(t *testing.T) {
		headers := []*BlockHeader{
			{Version: 1, Height: 100},
		}
		resp := &BlockHeadersResponse{
			Headers: headers,
			HasMore: true,
		}
		resp.Reset()
		assert.Nil(t, resp.Headers)
		assert.Equal(t, false, resp.HasMore)
	})

	t.Run("BlockHeadersResponseString", func(t *testing.T) {
		resp := &BlockHeadersResponse{
			Headers: []*BlockHeader{{Version: 1}},
			HasMore: true,
		}
		str := resp.String()
		assert.Contains(t, str, "headers")
	})

	t.Run("BlockHeadersResponseProtoMessage", func(t *testing.T) {
		resp := &BlockHeadersResponse{}
		resp.ProtoMessage() // Should not panic
	})

	t.Run("BlockHeadersResponseProtoReflect", func(t *testing.T) {
		resp := &BlockHeadersResponse{
			Headers: []*BlockHeader{{Version: 1}},
		}
		reflection := resp.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockHeadersResponseDescriptor", func(t *testing.T) {
		resp := &BlockHeadersResponse{}
		desc, _ := resp.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestBlockRequest(t *testing.T) {
	t.Run("NewBlockRequest", func(t *testing.T) {
		req := &BlockRequest{
			BlockHash: []byte("block_hash"),
			Height:    100,
		}

		assert.Equal(t, []byte("block_hash"), req.GetBlockHash())
		assert.Equal(t, uint64(100), req.GetHeight())
	})

	t.Run("BlockRequestReset", func(t *testing.T) {
		req := &BlockRequest{
			BlockHash: []byte("block_hash"),
			Height:    100,
		}
		req.Reset()
		assert.Nil(t, req.BlockHash)
		assert.Equal(t, uint64(0), req.Height)
	})

	t.Run("BlockRequestString", func(t *testing.T) {
		req := &BlockRequest{
			BlockHash: []byte("block_hash"),
			Height:    100,
		}
		str := req.String()
		assert.Contains(t, str, "block_hash")
	})

	t.Run("BlockRequestProtoMessage", func(t *testing.T) {
		req := &BlockRequest{}
		req.ProtoMessage() // Should not panic
	})

	t.Run("BlockRequestProtoReflect", func(t *testing.T) {
		req := &BlockRequest{
			BlockHash: []byte("block_hash"),
		}
		reflection := req.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockRequestDescriptor", func(t *testing.T) {
		req := &BlockRequest{}
		desc, _ := req.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestBlockResponse(t *testing.T) {
	t.Run("NewBlockResponse", func(t *testing.T) {
		resp := &BlockResponse{
			BlockData: []byte("block_data"),
			Found:     true,
		}

		assert.Equal(t, []byte("block_data"), resp.GetBlockData())
		assert.Equal(t, true, resp.GetFound())
	})

	t.Run("BlockResponseReset", func(t *testing.T) {
		resp := &BlockResponse{
			BlockData: []byte("block_data"),
			Found:     true,
		}
		resp.Reset()
		assert.Nil(t, resp.BlockData)
		assert.Equal(t, false, resp.Found)
	})

	t.Run("BlockResponseString", func(t *testing.T) {
		resp := &BlockResponse{
			BlockData: []byte("block_data"),
			Found:     true,
		}
		str := resp.String()
		assert.Contains(t, str, "block_data")
	})

	t.Run("BlockResponseProtoMessage", func(t *testing.T) {
		resp := &BlockResponse{}
		resp.ProtoMessage() // Should not panic
	})

	t.Run("BlockResponseProtoReflect", func(t *testing.T) {
		resp := &BlockResponse{
			BlockData: []byte("block_data"),
		}
		reflection := resp.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockResponseDescriptor", func(t *testing.T) {
		resp := &BlockResponse{}
		desc, _ := resp.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestSyncRequest(t *testing.T) {
	t.Run("NewSyncRequest", func(t *testing.T) {
		req := &SyncRequest{
			CurrentHeight: 100,
			BestBlockHash: []byte("best_hash"),
			KnownHeaders:  [][]byte{[]byte("header1"), []byte("header2")},
		}

		assert.Equal(t, uint64(100), req.GetCurrentHeight())
		assert.Equal(t, []byte("best_hash"), req.GetBestBlockHash())
		assert.Equal(t, [][]byte{[]byte("header1"), []byte("header2")}, req.GetKnownHeaders())
	})

	t.Run("SyncRequestReset", func(t *testing.T) {
		req := &SyncRequest{
			CurrentHeight: 100,
			BestBlockHash: []byte("best_hash"),
			KnownHeaders:  [][]byte{[]byte("header1")},
		}
		req.Reset()
		assert.Equal(t, uint64(0), req.CurrentHeight)
		assert.Nil(t, req.BestBlockHash)
		assert.Nil(t, req.KnownHeaders)
	})

	t.Run("SyncRequestString", func(t *testing.T) {
		req := &SyncRequest{
			CurrentHeight: 100,
			BestBlockHash: []byte("best_hash"),
		}
		str := req.String()
		assert.Contains(t, str, "current_height")
	})

	t.Run("SyncRequestProtoMessage", func(t *testing.T) {
		req := &SyncRequest{}
		req.ProtoMessage() // Should not panic
	})

	t.Run("SyncRequestProtoReflect", func(t *testing.T) {
		req := &SyncRequest{
			CurrentHeight: 100,
		}
		reflection := req.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("SyncRequestDescriptor", func(t *testing.T) {
		req := &SyncRequest{}
		desc, _ := req.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestSyncResponse(t *testing.T) {
	t.Run("NewSyncResponse", func(t *testing.T) {
		headers := []*BlockHeader{
			{Version: 1, Height: 100},
			{Version: 1, Height: 101},
		}
		resp := &SyncResponse{
			BestHeight:    200,
			BestBlockHash: []byte("best_hash"),
			Headers:       headers,
			NeedsSync:     true,
		}

		assert.Equal(t, uint64(200), resp.GetBestHeight())
		assert.Equal(t, []byte("best_hash"), resp.GetBestBlockHash())
		assert.Equal(t, headers, resp.GetHeaders())
		assert.Equal(t, true, resp.GetNeedsSync())
	})

	t.Run("SyncResponseReset", func(t *testing.T) {
		headers := []*BlockHeader{{Version: 1}}
		resp := &SyncResponse{
			BestHeight:    200,
			BestBlockHash: []byte("best_hash"),
			Headers:       headers,
			NeedsSync:     true,
		}
		resp.Reset()
		assert.Equal(t, uint64(0), resp.BestHeight)
		assert.Nil(t, resp.BestBlockHash)
		assert.Nil(t, resp.Headers)
		assert.Equal(t, false, resp.NeedsSync)
	})

	t.Run("SyncResponseString", func(t *testing.T) {
		resp := &SyncResponse{
			BestHeight: 200,
			NeedsSync:  true,
		}
		str := resp.String()
		assert.Contains(t, str, "best_height")
	})

	t.Run("SyncResponseProtoMessage", func(t *testing.T) {
		resp := &SyncResponse{}
		resp.ProtoMessage() // Should not panic
	})

	t.Run("SyncResponseProtoReflect", func(t *testing.T) {
		resp := &SyncResponse{
			BestHeight: 200,
		}
		reflection := resp.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("SyncResponseDescriptor", func(t *testing.T) {
		resp := &SyncResponse{}
		desc, _ := resp.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestStateRequest(t *testing.T) {
	t.Run("NewStateRequest", func(t *testing.T) {
		req := &StateRequest{
			Height:    100,
			StateRoot: []byte("state_root"),
		}

		assert.Equal(t, uint64(100), req.GetHeight())
		assert.Equal(t, []byte("state_root"), req.GetStateRoot())
	})

	t.Run("StateRequestReset", func(t *testing.T) {
		req := &StateRequest{
			Height:    100,
			StateRoot: []byte("state_root"),
		}
		req.Reset()
		assert.Equal(t, uint64(0), req.Height)
		assert.Nil(t, req.StateRoot)
	})

	t.Run("StateRequestString", func(t *testing.T) {
		req := &StateRequest{
			Height:    100,
			StateRoot: []byte("state_root"),
		}
		str := req.String()
		assert.Contains(t, str, "height")
	})

	t.Run("StateRequestProtoMessage", func(t *testing.T) {
		req := &StateRequest{}
		req.ProtoMessage() // Should not panic
	})

	t.Run("StateRequestProtoReflect", func(t *testing.T) {
		req := &StateRequest{
			Height: 100,
		}
		reflection := req.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("StateRequestDescriptor", func(t *testing.T) {
		req := &StateRequest{}
		desc, _ := req.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestStateResponse(t *testing.T) {
	t.Run("NewStateResponse", func(t *testing.T) {
		resp := &StateResponse{
			StateData: []byte("state_data"),
			Height:    100,
			StateRoot: []byte("state_root"),
			Found:     true,
		}

		assert.Equal(t, []byte("state_data"), resp.GetStateData())
		assert.Equal(t, uint64(100), resp.GetHeight())
		assert.Equal(t, []byte("state_root"), resp.GetStateRoot())
		assert.Equal(t, true, resp.GetFound())
	})

	t.Run("StateResponseReset", func(t *testing.T) {
		resp := &StateResponse{
			StateData: []byte("state_data"),
			Height:    100,
			StateRoot: []byte("state_root"),
			Found:     true,
		}
		resp.Reset()
		assert.Nil(t, resp.StateData)
		assert.Equal(t, uint64(0), resp.Height)
		assert.Nil(t, resp.StateRoot)
		assert.Equal(t, false, resp.Found)
	})

	t.Run("StateResponseString", func(t *testing.T) {
		resp := &StateResponse{
			StateData: []byte("state_data"),
			Height:    100,
		}
		str := resp.String()
		assert.Contains(t, str, "state_data")
	})

	t.Run("StateResponseProtoMessage", func(t *testing.T) {
		resp := &StateResponse{}
		resp.ProtoMessage() // Should not panic
	})

	t.Run("StateResponseProtoReflect", func(t *testing.T) {
		resp := &StateResponse{
			StateData: []byte("state_data"),
		}
		reflection := resp.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("StateResponseDescriptor", func(t *testing.T) {
		resp := &StateResponse{}
		desc, _ := resp.Descriptor()
		assert.NotNil(t, desc)
	})
}

func TestMessage(t *testing.T) {
	t.Run("NewMessage", func(t *testing.T) {
		now := time.Now().UnixNano()
		msg := &Message{
			TimestampUnixNano: now,
			FromPeerId:        []byte("peer_id"),
			Signature:         []byte("signature"),
		}

		assert.Equal(t, now, msg.GetTimestampUnixNano())
		assert.Equal(t, []byte("peer_id"), msg.GetFromPeerId())
		assert.Equal(t, []byte("signature"), msg.GetSignature())
	})

	t.Run("MessageWithBlockMessage", func(t *testing.T) {
		blockMsg := &BlockMessage{BlockData: []byte("block data")}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_BlockMessage{
				BlockMessage: blockMsg,
			},
		}

		assert.Equal(t, blockMsg, msg.GetBlockMessage())
		assert.Nil(t, msg.GetTransactionMessage())
		assert.Nil(t, msg.GetHeadersRequest())
		assert.Nil(t, msg.GetHeadersResponse())
		assert.Nil(t, msg.GetBlockRequest())
		assert.Nil(t, msg.GetBlockResponse())
		assert.Nil(t, msg.GetSyncRequest())
		assert.Nil(t, msg.GetSyncResponse())
		assert.Nil(t, msg.GetStateRequest())
		assert.Nil(t, msg.GetStateResponse())
	})

	t.Run("MessageWithTransactionMessage", func(t *testing.T) {
		txMsg := &TransactionMessage{TransactionData: []byte("tx data")}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_TransactionMessage{
				TransactionMessage: txMsg,
			},
		}

		assert.Equal(t, txMsg, msg.GetTransactionMessage())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithHeadersRequest", func(t *testing.T) {
		headersReq := &BlockHeadersRequest{StartHeight: 100, Count: 50}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_HeadersRequest{
				HeadersRequest: headersReq,
			},
		}

		assert.Equal(t, headersReq, msg.GetHeadersRequest())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithHeadersResponse", func(t *testing.T) {
		headersResp := &BlockHeadersResponse{Headers: []*BlockHeader{{Version: 1}}}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_HeadersResponse{
				HeadersResponse: headersResp,
			},
		}

		assert.Equal(t, headersResp, msg.GetHeadersResponse())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithBlockRequest", func(t *testing.T) {
		blockReq := &BlockRequest{Height: 100}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_BlockRequest{
				BlockRequest: blockReq,
			},
		}

		assert.Equal(t, blockReq, msg.GetBlockRequest())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithBlockResponse", func(t *testing.T) {
		blockResp := &BlockResponse{Found: true}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_BlockResponse{
				BlockResponse: blockResp,
			},
		}

		assert.Equal(t, blockResp, msg.GetBlockResponse())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithSyncRequest", func(t *testing.T) {
		syncReq := &SyncRequest{CurrentHeight: 100}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_SyncRequest{
				SyncRequest: syncReq,
			},
		}

		assert.Equal(t, syncReq, msg.GetSyncRequest())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithSyncResponse", func(t *testing.T) {
		syncResp := &SyncResponse{BestHeight: 200}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_SyncResponse{
				SyncResponse: syncResp,
			},
		}

		assert.Equal(t, syncResp, msg.GetSyncResponse())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithStateRequest", func(t *testing.T) {
		stateReq := &StateRequest{Height: 100}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_StateRequest{
				StateRequest: stateReq,
			},
		}

		assert.Equal(t, stateReq, msg.GetStateRequest())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithStateResponse", func(t *testing.T) {
		stateResp := &StateResponse{Height: 100, Found: true}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			Content: &Message_StateResponse{
				StateResponse: stateResp,
			},
		}

		assert.Equal(t, stateResp, msg.GetStateResponse())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageReset", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Signature:         []byte("signature"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{BlockData: []byte("data")},
			},
		}
		msg.Reset()
		assert.Equal(t, int64(0), msg.TimestampUnixNano)
		assert.Nil(t, msg.FromPeerId)
		assert.Nil(t, msg.Signature)
		assert.Nil(t, msg.Content)
	})

	t.Run("MessageString", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
		}
		str := msg.String()
		assert.Contains(t, str, "timestamp_unix_nano")
	})

	t.Run("MessageProtoMessage", func(t *testing.T) {
		msg := &Message{}
		msg.ProtoMessage() // Should not panic
	})

	t.Run("MessageProtoReflect", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
		}
		reflection := msg.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("MessageDescriptor", func(t *testing.T) {
		msg := &Message{}
		desc, _ := msg.Descriptor()
		assert.NotNil(t, desc)
	})

	t.Run("MessageGetContent", func(t *testing.T) {
		msg := &Message{
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{BlockData: []byte("data")},
			},
		}
		content := msg.GetContent()
		assert.NotNil(t, content)
	})
}

func TestMessageContentInterface(t *testing.T) {
	t.Run("BlockMessageIsContent", func(t *testing.T) {
		content := &Message_BlockMessage{
			BlockMessage: &BlockMessage{BlockData: []byte("data")},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("TransactionMessageIsContent", func(t *testing.T) {
		content := &Message_TransactionMessage{
			TransactionMessage: &TransactionMessage{TransactionData: []byte("data")},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("HeadersRequestIsContent", func(t *testing.T) {
		content := &Message_HeadersRequest{
			HeadersRequest: &BlockHeadersRequest{StartHeight: 100},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("HeadersResponseIsContent", func(t *testing.T) {
		content := &Message_HeadersResponse{
			HeadersResponse: &BlockHeadersResponse{Headers: []*BlockHeader{}},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("BlockRequestIsContent", func(t *testing.T) {
		content := &Message_BlockRequest{
			BlockRequest: &BlockRequest{Height: 100},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("BlockResponseIsContent", func(t *testing.T) {
		content := &Message_BlockResponse{
			BlockResponse: &BlockResponse{Found: true},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("SyncRequestIsContent", func(t *testing.T) {
		content := &Message_SyncRequest{
			SyncRequest: &SyncRequest{CurrentHeight: 100},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("SyncResponseIsContent", func(t *testing.T) {
		content := &Message_SyncResponse{
			SyncResponse: &SyncResponse{BestHeight: 200},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("StateRequestIsContent", func(t *testing.T) {
		content := &Message_StateRequest{
			StateRequest: &StateRequest{Height: 100},
		}
		content.isMessage_Content() // Should not panic
	})

	t.Run("StateResponseIsContent", func(t *testing.T) {
		content := &Message_StateResponse{
			StateResponse: &StateResponse{Height: 100, Found: true},
		}
		content.isMessage_Content() // Should not panic
	})
}

func TestProtoSerialization(t *testing.T) {
	t.Run("BlockMessageSerialization", func(t *testing.T) {
		original := &BlockMessage{
			BlockData: []byte("test block data"),
		}

		data, err := proto.Marshal(original)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		unmarshaled := &BlockMessage{}
		err = proto.Unmarshal(data, unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, original.BlockData, unmarshaled.BlockData)
	})

	t.Run("TransactionMessageSerialization", func(t *testing.T) {
		original := &TransactionMessage{
			TransactionData: []byte("test transaction data"),
		}

		data, err := proto.Marshal(original)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		unmarshaled := &TransactionMessage{}
		err = proto.Unmarshal(data, unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, original.TransactionData, unmarshaled.TransactionData)
	})

	t.Run("BlockHeaderSerialization", func(t *testing.T) {
		original := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     1234567890,
			Difficulty:    1000,
			Nonce:         42,
			Height:        100,
			Hash:          []byte("block_hash"),
		}

		data, err := proto.Marshal(original)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		unmarshaled := &BlockHeader{}
		err = proto.Unmarshal(data, unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, original.Version, unmarshaled.Version)
		assert.Equal(t, original.PrevBlockHash, unmarshaled.PrevBlockHash)
		assert.Equal(t, original.MerkleRoot, unmarshaled.MerkleRoot)
		assert.Equal(t, original.Timestamp, unmarshaled.Timestamp)
		assert.Equal(t, original.Difficulty, unmarshaled.Difficulty)
		assert.Equal(t, original.Nonce, unmarshaled.Nonce)
		assert.Equal(t, original.Height, unmarshaled.Height)
		assert.Equal(t, original.Hash, unmarshaled.Hash)
	})

	t.Run("MessageSerialization", func(t *testing.T) {
		original := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Signature:         []byte("signature"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{BlockData: []byte("block data")},
			},
		}

		data, err := proto.Marshal(original)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		unmarshaled := &Message{}
		err = proto.Unmarshal(data, unmarshaled)
		assert.NoError(t, err)
		assert.Equal(t, original.TimestampUnixNano, unmarshaled.TimestampUnixNano)
		assert.Equal(t, original.FromPeerId, unmarshaled.FromPeerId)
		assert.Equal(t, original.Signature, unmarshaled.Signature)
		assert.NotNil(t, unmarshaled.GetBlockMessage())
		assert.Equal(t, original.GetBlockMessage().BlockData, unmarshaled.GetBlockMessage().BlockData)
	})
}
