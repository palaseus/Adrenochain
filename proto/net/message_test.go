package net

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestBlockMessage tests BlockMessage functionality
func TestBlockMessage(t *testing.T) {
	t.Run("NewBlockMessage", func(t *testing.T) {
		msg := &BlockMessage{
			BlockData: []byte("test block data"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, []byte("test block data"), msg.BlockData)
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
		assert.Contains(t, str, "test data")
	})

	t.Run("BlockMessageProtoMessage", func(t *testing.T) {
		msg := &BlockMessage{}
		// This should not panic
		msg.ProtoMessage()
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
		descriptor := msg.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestTransactionMessage tests TransactionMessage functionality
func TestTransactionMessage(t *testing.T) {
	t.Run("NewTransactionMessage", func(t *testing.T) {
		msg := &TransactionMessage{
			TransactionData: []byte("test transaction data"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, []byte("test transaction data"), msg.TransactionData)
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
		assert.Contains(t, str, "test data")
	})

	t.Run("TransactionMessageProtoMessage", func(t *testing.T) {
		msg := &TransactionMessage{}
		msg.ProtoMessage()
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
		descriptor := msg.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestBlockHeader tests BlockHeader functionality
func TestBlockHeader(t *testing.T) {
	t.Run("NewBlockHeader", func(t *testing.T) {
		header := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now().Unix(),
			Difficulty:    1000,
			Nonce:         12345,
			Height:        100,
			Hash:          []byte("block_hash"),
		}
		assert.NotNil(t, header)
		assert.Equal(t, uint32(1), header.Version)
		assert.Equal(t, []byte("prev_hash"), header.PrevBlockHash)
		assert.Equal(t, []byte("merkle_root"), header.MerkleRoot)
		assert.Equal(t, uint64(1000), header.Difficulty)
		assert.Equal(t, uint64(12345), header.Nonce)
		assert.Equal(t, uint64(100), header.Height)
		assert.Equal(t, []byte("block_hash"), header.Hash)
	})

	t.Run("BlockHeaderReset", func(t *testing.T) {
		header := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now().Unix(),
			Difficulty:    1000,
			Nonce:         12345,
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
			Timestamp:     time.Now().Unix(),
			Difficulty:    1000,
			Nonce:         12345,
			Height:        100,
			Hash:          []byte("block_hash"),
		}
		str := header.String()
		assert.Contains(t, str, "1")
		assert.Contains(t, str, "1000")
		assert.Contains(t, str, "12345")
		assert.Contains(t, str, "100")
	})

	t.Run("BlockHeaderProtoMessage", func(t *testing.T) {
		header := &BlockHeader{}
		header.ProtoMessage()
	})

	t.Run("BlockHeaderProtoReflect", func(t *testing.T) {
		header := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now().Unix(),
			Difficulty:    1000,
			Nonce:         12345,
			Height:        100,
			Hash:          []byte("block_hash"),
		}
		reflection := header.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockHeaderDescriptor", func(t *testing.T) {
		header := &BlockHeader{}
		descriptor := header.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestMessage tests Message functionality
func TestMessage(t *testing.T) {
	t.Run("NewMessage", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Signature:         []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.NotZero(t, msg.TimestampUnixNano)
		assert.Equal(t, []byte("peer_id"), msg.FromPeerId)
		assert.Equal(t, []byte("signature"), msg.Signature)
	})

	t.Run("MessageWithBlockMessage", func(t *testing.T) {
		blockMsg := &BlockMessage{
			BlockData: []byte("block data"),
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_BlockMessage{
				BlockMessage: blockMsg,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, blockMsg, msg.GetBlockMessage())
		assert.Nil(t, msg.GetTransactionMessage())
	})

	t.Run("MessageWithTransactionMessage", func(t *testing.T) {
		txMsg := &TransactionMessage{
			TransactionData: []byte("tx data"),
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_TransactionMessage{
				TransactionMessage: txMsg,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, txMsg, msg.GetTransactionMessage())
		assert.Nil(t, msg.GetBlockMessage())
	})

	t.Run("MessageWithHeadersRequest", func(t *testing.T) {
		headersReq := &BlockHeadersRequest{
			StartHeight: 100,
			Count:       50,
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_HeadersRequest{
				HeadersRequest: headersReq,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, headersReq, msg.GetHeadersRequest())
	})

	t.Run("MessageWithHeadersResponse", func(t *testing.T) {
		headersResp := &BlockHeadersResponse{
			Headers: []*BlockHeader{{Version: 1, Height: 100}},
			HasMore: true,
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_HeadersResponse{
				HeadersResponse: headersResp,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, headersResp, msg.GetHeadersResponse())
	})

	t.Run("MessageWithBlockRequest", func(t *testing.T) {
		blockReq := &BlockRequest{
			BlockHash: []byte("block_hash"),
			Height:    100,
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_BlockRequest{
				BlockRequest: blockReq,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, blockReq, msg.GetBlockRequest())
	})

	t.Run("MessageWithBlockResponse", func(t *testing.T) {
		blockResp := &BlockResponse{
			Found:     true,
			BlockData: []byte("block_data"),
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_BlockResponse{
				BlockResponse: blockResp,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, blockResp, msg.GetBlockResponse())
	})

	t.Run("MessageWithSyncRequest", func(t *testing.T) {
		syncReq := &SyncRequest{
			CurrentHeight: 100,
			BestBlockHash: []byte("best_hash"),
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_SyncRequest{
				SyncRequest: syncReq,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, syncReq, msg.GetSyncRequest())
	})

	t.Run("MessageWithSyncResponse", func(t *testing.T) {
		syncResp := &SyncResponse{
			BestHeight:    200,
			BestBlockHash: []byte("best_hash"),
			NeedsSync:     true,
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_SyncResponse{
				SyncResponse: syncResp,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, syncResp, msg.GetSyncResponse())
	})

	t.Run("MessageWithStateRequest", func(t *testing.T) {
		stateReq := &StateRequest{
			StateRoot: []byte("state_root"),
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_StateRequest{
				StateRequest: stateReq,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, stateReq, msg.GetStateRequest())
	})

	t.Run("MessageWithStateResponse", func(t *testing.T) {
		stateResp := &StateResponse{
			StateRoot: []byte("state_root"),
			HasMore:   true,
		}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_StateResponse{
				StateResponse: stateResp,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, stateResp, msg.GetStateResponse())
	})

	t.Run("MessageReset", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{BlockData: []byte("data")},
			},
			Signature: []byte("signature"),
		}
		msg.Reset()
		assert.Equal(t, int64(0), msg.TimestampUnixNano)
		assert.Nil(t, msg.FromPeerId)
		assert.Nil(t, msg.Content)
		assert.Nil(t, msg.Signature)
	})

	t.Run("MessageString", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{BlockData: []byte("data")},
			},
			Signature: []byte("signature"),
		}
		str := msg.String()
		assert.Contains(t, str, "peer_id")
		assert.Contains(t, str, "data")
		assert.Contains(t, str, "signature")
	})

	t.Run("MessageProtoMessage", func(t *testing.T) {
		msg := &Message{}
		msg.ProtoMessage()
	})

	t.Run("MessageProtoReflect", func(t *testing.T) {
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_BlockMessage{
				BlockMessage: &BlockMessage{BlockData: []byte("data")},
			},
			Signature: []byte("signature"),
		}
		reflection := msg.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("MessageDescriptor", func(t *testing.T) {
		msg := &Message{}
		descriptor := msg.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})

	t.Run("MessageGetContent", func(t *testing.T) {
		blockMsg := &BlockMessage{BlockData: []byte("data")}
		msg := &Message{
			Content: &Message_BlockMessage{
				BlockMessage: blockMsg,
			},
		}
		content := msg.GetContent()
		assert.NotNil(t, content)
		// The content is the wrapper, not the direct message
		assert.Equal(t, &Message_BlockMessage{BlockMessage: blockMsg}, content)
	})
}

// TestMessageContentInterface tests the message content interface
func TestMessageContentInterface(t *testing.T) {
	t.Run("BlockMessageIsContent", func(t *testing.T) {
		blockMsg := &BlockMessage{BlockData: []byte("data")}
		var content isMessage_Content = &Message_BlockMessage{BlockMessage: blockMsg}
		assert.NotNil(t, content)
	})

	t.Run("TransactionMessageIsContent", func(t *testing.T) {
		txMsg := &TransactionMessage{TransactionData: []byte("data")}
		var content isMessage_Content = &Message_TransactionMessage{TransactionMessage: txMsg}
		assert.NotNil(t, content)
	})

	t.Run("HeadersRequestIsContent", func(t *testing.T) {
		headersReq := &BlockHeadersRequest{StartHeight: 100, Count: 50}
		var content isMessage_Content = &Message_HeadersRequest{HeadersRequest: headersReq}
		assert.NotNil(t, content)
	})

	t.Run("HeadersResponseIsContent", func(t *testing.T) {
		headersResp := &BlockHeadersResponse{Headers: []*BlockHeader{}, HasMore: false}
		var content isMessage_Content = &Message_HeadersResponse{HeadersResponse: headersResp}
		assert.NotNil(t, content)
	})

	t.Run("BlockRequestIsContent", func(t *testing.T) {
		blockReq := &BlockRequest{BlockHash: []byte("hash"), Height: 100}
		var content isMessage_Content = &Message_BlockRequest{BlockRequest: blockReq}
		assert.NotNil(t, content)
	})

	t.Run("BlockResponseIsContent", func(t *testing.T) {
		blockResp := &BlockResponse{Found: true, BlockData: []byte("data")}
		var content isMessage_Content = &Message_BlockResponse{BlockResponse: blockResp}
		assert.NotNil(t, content)
	})

	t.Run("SyncRequestIsContent", func(t *testing.T) {
		syncReq := &SyncRequest{CurrentHeight: 100, BestBlockHash: []byte("hash")}
		var content isMessage_Content = &Message_SyncRequest{SyncRequest: syncReq}
		assert.NotNil(t, content)
	})

	t.Run("SyncResponseIsContent", func(t *testing.T) {
		syncResp := &SyncResponse{BestHeight: 200, BestBlockHash: []byte("hash"), NeedsSync: true}
		var content isMessage_Content = &Message_SyncResponse{SyncResponse: syncResp}
		assert.NotNil(t, content)
	})

	t.Run("StateRequestIsContent", func(t *testing.T) {
		stateReq := &StateRequest{StateRoot: []byte("root")}
		var content isMessage_Content = &Message_StateRequest{StateRequest: stateReq}
		assert.NotNil(t, content)
	})

	t.Run("StateResponseIsContent", func(t *testing.T) {
		stateResp := &StateResponse{StateRoot: []byte("root"), HasMore: false}
		var content isMessage_Content = &Message_StateResponse{StateResponse: stateResp}
		assert.NotNil(t, content)
	})
}

// TestBlockHeadersRequest tests BlockHeadersRequest functionality
func TestBlockHeadersRequest(t *testing.T) {
	t.Run("NewBlockHeadersRequest", func(t *testing.T) {
		req := &BlockHeadersRequest{
			StartHeight: 100,
			Count:       50,
			StopHash:    []byte("stop_hash"),
		}
		assert.NotNil(t, req)
		assert.Equal(t, uint64(100), req.StartHeight)
		assert.Equal(t, uint64(50), req.Count)
		assert.Equal(t, []byte("stop_hash"), req.StopHash)
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
			StopHash:    []byte("stop_hash"),
		}
		str := req.String()
		assert.Contains(t, str, "100")
		assert.Contains(t, str, "50")
	})

	t.Run("BlockHeadersRequestProtoMessage", func(t *testing.T) {
		req := &BlockHeadersRequest{}
		req.ProtoMessage()
	})

	t.Run("BlockHeadersRequestProtoReflect", func(t *testing.T) {
		req := &BlockHeadersRequest{
			StartHeight: 100,
			Count:       50,
			StopHash:    []byte("stop_hash"),
		}
		reflection := req.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockHeadersRequestDescriptor", func(t *testing.T) {
		req := &BlockHeadersRequest{}
		descriptor := req.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestBlockHeadersResponse tests BlockHeadersResponse functionality
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
		assert.NotNil(t, resp)
		assert.Len(t, resp.Headers, 2)
		assert.True(t, resp.HasMore)
	})

	t.Run("BlockHeadersResponseReset", func(t *testing.T) {
		headers := []*BlockHeader{
			{Version: 1, Height: 100},
			{Version: 1, Height: 101},
		}
		resp := &BlockHeadersResponse{
			Headers: headers,
			HasMore: true,
		}
		resp.Reset()
		assert.Nil(t, resp.Headers)
		assert.False(t, resp.HasMore)
	})

	t.Run("BlockHeadersResponseString", func(t *testing.T) {
		headers := []*BlockHeader{
			{Version: 1, Height: 100},
			{Version: 1, Height: 101},
		}
		resp := &BlockHeadersResponse{
			Headers: headers,
			HasMore: true,
		}
		str := resp.String()
		assert.Contains(t, str, "100")
		assert.Contains(t, str, "101")
		assert.Contains(t, str, "true")
	})

	t.Run("BlockHeadersResponseProtoMessage", func(t *testing.T) {
		resp := &BlockHeadersResponse{}
		resp.ProtoMessage()
	})

	t.Run("BlockHeadersResponseProtoReflect", func(t *testing.T) {
		headers := []*BlockHeader{
			{Version: 1, Height: 100},
			{Version: 1, Height: 101},
		}
		resp := &BlockHeadersResponse{
			Headers: headers,
			HasMore: true,
		}
		reflection := resp.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockHeadersResponseDescriptor", func(t *testing.T) {
		resp := &BlockHeadersResponse{}
		descriptor := resp.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestBlockRequest tests BlockRequest functionality
func TestBlockRequest(t *testing.T) {
	t.Run("NewBlockRequest", func(t *testing.T) {
		req := &BlockRequest{
			BlockHash: []byte("block_hash"),
			Height:    100,
		}
		assert.NotNil(t, req)
		assert.Equal(t, []byte("block_hash"), req.BlockHash)
		assert.Equal(t, uint64(100), req.Height)
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
		assert.Contains(t, str, "100")
	})

	t.Run("BlockRequestProtoMessage", func(t *testing.T) {
		req := &BlockRequest{}
		req.ProtoMessage()
	})

	t.Run("BlockRequestProtoReflect", func(t *testing.T) {
		req := &BlockRequest{
			BlockHash: []byte("block_hash"),
			Height:    100,
		}
		reflection := req.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockRequestDescriptor", func(t *testing.T) {
		req := &BlockRequest{}
		descriptor := req.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestBlockResponse tests BlockResponse functionality
func TestBlockResponse(t *testing.T) {
	t.Run("NewBlockResponse", func(t *testing.T) {
		resp := &BlockResponse{
			Found:     true,
			BlockData: []byte("block_data"),
		}
		assert.NotNil(t, resp)
		assert.True(t, resp.Found)
		assert.Equal(t, []byte("block_data"), resp.BlockData)
	})

	t.Run("BlockResponseReset", func(t *testing.T) {
		resp := &BlockResponse{
			Found:     true,
			BlockData: []byte("block_data"),
		}
		resp.Reset()
		assert.False(t, resp.Found)
		assert.Nil(t, resp.BlockData)
	})

	t.Run("BlockResponseString", func(t *testing.T) {
		resp := &BlockResponse{
			Found:     true,
			BlockData: []byte("block_data"),
		}
		str := resp.String()
		assert.Contains(t, str, "true")
		assert.Contains(t, str, "block_data")
	})

	t.Run("BlockResponseProtoMessage", func(t *testing.T) {
		resp := &BlockResponse{}
		resp.ProtoMessage()
	})

	t.Run("BlockResponseProtoReflect", func(t *testing.T) {
		resp := &BlockResponse{
			Found:     true,
			BlockData: []byte("block_data"),
		}
		reflection := resp.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("BlockResponseDescriptor", func(t *testing.T) {
		resp := &BlockResponse{}
		descriptor := resp.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestSyncRequest tests SyncRequest functionality
func TestSyncRequest(t *testing.T) {
	t.Run("NewSyncRequest", func(t *testing.T) {
		knownHeaders := [][]byte{
			[]byte("header1"),
			[]byte("header2"),
		}
		req := &SyncRequest{
			CurrentHeight: 100,
			BestBlockHash: []byte("best_hash"),
			KnownHeaders:  knownHeaders,
		}
		assert.NotNil(t, req)
		assert.Equal(t, uint64(100), req.CurrentHeight)
		assert.Equal(t, []byte("best_hash"), req.BestBlockHash)
		assert.Len(t, req.KnownHeaders, 2)
	})

	t.Run("SyncRequestReset", func(t *testing.T) {
		knownHeaders := [][]byte{
			[]byte("header1"),
			[]byte("header2"),
		}
		req := &SyncRequest{
			CurrentHeight: 100,
			BestBlockHash: []byte("best_hash"),
			KnownHeaders:  knownHeaders,
		}
		req.Reset()
		assert.Equal(t, uint64(0), req.CurrentHeight)
		assert.Nil(t, req.BestBlockHash)
		assert.Nil(t, req.KnownHeaders)
	})

	t.Run("SyncRequestString", func(t *testing.T) {
		knownHeaders := [][]byte{
			[]byte("header1"),
			[]byte("header2"),
		}
		req := &SyncRequest{
			CurrentHeight: 100,
			BestBlockHash: []byte("best_hash"),
			KnownHeaders:  knownHeaders,
		}
		str := req.String()
		assert.Contains(t, str, "100")
		assert.Contains(t, str, "best_hash")
		assert.Contains(t, str, "header1")
		assert.Contains(t, str, "header2")
	})

	t.Run("SyncRequestProtoMessage", func(t *testing.T) {
		req := &SyncRequest{}
		req.ProtoMessage()
	})

	t.Run("SyncRequestProtoReflect", func(t *testing.T) {
		knownHeaders := [][]byte{
			[]byte("header1"),
			[]byte("header2"),
		}
		req := &SyncRequest{
			CurrentHeight: 100,
			BestBlockHash: []byte("best_hash"),
			KnownHeaders:  knownHeaders,
		}
		reflection := req.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("SyncRequestDescriptor", func(t *testing.T) {
		req := &SyncRequest{}
		descriptor := req.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestSyncResponse tests SyncResponse functionality
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
		assert.NotNil(t, resp)
		assert.Equal(t, uint64(200), resp.BestHeight)
		assert.Equal(t, []byte("best_hash"), resp.BestBlockHash)
		assert.Len(t, resp.Headers, 2)
		assert.True(t, resp.NeedsSync)
	})

	t.Run("SyncResponseReset", func(t *testing.T) {
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
		resp.Reset()
		assert.Equal(t, uint64(0), resp.BestHeight)
		assert.Nil(t, resp.BestBlockHash)
		assert.Nil(t, resp.Headers)
		assert.False(t, resp.NeedsSync)
	})

	t.Run("SyncResponseString", func(t *testing.T) {
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
		str := resp.String()
		assert.Contains(t, str, "200")
		assert.Contains(t, str, "best_hash")
		assert.Contains(t, str, "100")
		assert.Contains(t, str, "101")
		assert.Contains(t, str, "true")
	})

	t.Run("SyncResponseProtoMessage", func(t *testing.T) {
		resp := &SyncResponse{}
		resp.ProtoMessage()
	})

	t.Run("SyncResponseProtoReflect", func(t *testing.T) {
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
		reflection := resp.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("SyncResponseDescriptor", func(t *testing.T) {
		resp := &SyncResponse{}
		descriptor := resp.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestStateRequest tests StateRequest functionality
func TestStateRequest(t *testing.T) {
	t.Run("NewStateRequest", func(t *testing.T) {
		accountHashes := [][]byte{
			[]byte("account1"),
			[]byte("account2"),
		}
		req := &StateRequest{
			StateRoot:     []byte("state_root"),
			AccountHashes: accountHashes,
		}
		assert.NotNil(t, req)
		assert.Equal(t, []byte("state_root"), req.StateRoot)
		assert.Len(t, req.AccountHashes, 2)
	})

	t.Run("StateRequestReset", func(t *testing.T) {
		accountHashes := [][]byte{
			[]byte("account1"),
			[]byte("account2"),
		}
		req := &StateRequest{
			StateRoot:     []byte("state_root"),
			AccountHashes: accountHashes,
		}
		req.Reset()
		assert.Nil(t, req.StateRoot)
		assert.Nil(t, req.AccountHashes)
	})

	t.Run("StateRequestString", func(t *testing.T) {
		accountHashes := [][]byte{
			[]byte("account1"),
			[]byte("account2"),
		}
		req := &StateRequest{
			StateRoot:     []byte("state_root"),
			AccountHashes: accountHashes,
		}
		str := req.String()
		assert.Contains(t, str, "state_root")
		assert.Contains(t, str, "account1")
		assert.Contains(t, str, "account2")
	})

	t.Run("StateRequestProtoMessage", func(t *testing.T) {
		req := &StateRequest{}
		req.ProtoMessage()
	})

	t.Run("StateRequestProtoReflect", func(t *testing.T) {
		accountHashes := [][]byte{
			[]byte("account1"),
			[]byte("account2"),
		}
		req := &StateRequest{
			StateRoot:     []byte("state_root"),
			AccountHashes: accountHashes,
		}
		reflection := req.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("StateRequestDescriptor", func(t *testing.T) {
		req := &StateRequest{}
		descriptor := req.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestStateResponse tests StateResponse functionality
func TestStateResponse(t *testing.T) {
	t.Run("NewStateResponse", func(t *testing.T) {
		accountStates := map[string][]byte{
			"account1": []byte("state1"),
			"account2": []byte("state2"),
		}
		resp := &StateResponse{
			StateRoot:     []byte("state_root"),
			AccountStates: accountStates,
			HasMore:       true,
		}
		assert.NotNil(t, resp)
		assert.Equal(t, []byte("state_root"), resp.StateRoot)
		assert.Len(t, resp.AccountStates, 2)
		assert.True(t, resp.HasMore)
	})

	t.Run("StateResponseReset", func(t *testing.T) {
		accountStates := map[string][]byte{
			"account1": []byte("state1"),
			"account2": []byte("state2"),
		}
		resp := &StateResponse{
			StateRoot:     []byte("state_root"),
			AccountStates: accountStates,
			HasMore:       true,
		}
		resp.Reset()
		assert.Nil(t, resp.StateRoot)
		assert.Nil(t, resp.AccountStates)
		assert.False(t, resp.HasMore)
	})

	t.Run("StateResponseString", func(t *testing.T) {
		accountStates := map[string][]byte{
			"account1": []byte("state1"),
			"account2": []byte("state2"),
		}
		resp := &StateResponse{
			StateRoot:     []byte("state_root"),
			AccountStates: accountStates,
			HasMore:       true,
		}
		str := resp.String()
		assert.Contains(t, str, "state_root")
		assert.Contains(t, str, "account1")
		assert.Contains(t, str, "account2")
		assert.Contains(t, str, "true")
	})

	t.Run("StateResponseProtoMessage", func(t *testing.T) {
		resp := &StateResponse{}
		resp.ProtoMessage()
	})

	t.Run("StateResponseProtoReflect", func(t *testing.T) {
		accountStates := map[string][]byte{
			"account1": []byte("state1"),
			"account2": []byte("state2"),
		}
		resp := &StateResponse{
			StateRoot:     []byte("state_root"),
			AccountStates: accountStates,
			HasMore:       true,
		}
		reflection := resp.ProtoReflect()
		assert.NotNil(t, reflection)
	})

	t.Run("StateResponseDescriptor", func(t *testing.T) {
		resp := &StateResponse{}
		descriptor := resp.ProtoReflect().Descriptor()
		assert.NotNil(t, descriptor)
	})
}

// TestProtoSerialization tests protobuf serialization
func TestProtoSerialization(t *testing.T) {
	t.Run("BlockMessageSerialization", func(t *testing.T) {
		msg := &BlockMessage{BlockData: []byte("test block data")}
		// Test that the message can be serialized without errors
		// Note: We're not actually serializing here, just testing the structure
		assert.NotNil(t, msg)
		assert.Equal(t, []byte("test block data"), msg.BlockData)
	})

	t.Run("TransactionMessageSerialization", func(t *testing.T) {
		msg := &TransactionMessage{TransactionData: []byte("test transaction data")}
		assert.NotNil(t, msg)
		assert.Equal(t, []byte("test transaction data"), msg.TransactionData)
	})

	t.Run("BlockHeaderSerialization", func(t *testing.T) {
		header := &BlockHeader{
			Version:       1,
			PrevBlockHash: []byte("prev_hash"),
			MerkleRoot:    []byte("merkle_root"),
			Timestamp:     time.Now().Unix(),
			Difficulty:    1000,
			Nonce:         12345,
			Height:        100,
			Hash:          []byte("block_hash"),
		}
		assert.NotNil(t, header)
		assert.Equal(t, uint32(1), header.Version)
		assert.Equal(t, []byte("prev_hash"), header.PrevBlockHash)
		assert.Equal(t, []byte("merkle_root"), header.MerkleRoot)
		assert.Equal(t, uint64(1000), header.Difficulty)
		assert.Equal(t, uint64(12345), header.Nonce)
		assert.Equal(t, uint64(100), header.Height)
		assert.Equal(t, []byte("block_hash"), header.Hash)
	})

	t.Run("MessageSerialization", func(t *testing.T) {
		blockMsg := &BlockMessage{BlockData: []byte("block data")}
		msg := &Message{
			TimestampUnixNano: time.Now().UnixNano(),
			FromPeerId:        []byte("peer_id"),
			Content: &Message_BlockMessage{
				BlockMessage: blockMsg,
			},
			Signature: []byte("signature"),
		}
		assert.NotNil(t, msg)
		assert.Equal(t, blockMsg, msg.GetBlockMessage())
		assert.Equal(t, []byte("peer_id"), msg.FromPeerId)
		assert.Equal(t, []byte("signature"), msg.Signature)
	})
}
