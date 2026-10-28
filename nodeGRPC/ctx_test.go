package nodeGRPC

import (
	"context"
	"errors"
	"testing"

	"github.com/Jagtech-Services/go-tari-grpc-lib/v3/tari_generated"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The fakeBaseNodeServer type and startFakeNode helper live in client_test.go. The methods
// below extend that same type with just enough of the BaseNodeServer interface to exercise
// GetTipInfoCtx, SubmitBlockCtx, and GetNewBlockTemplateWithCoinbasesCtx over the in-memory
// bufconn listener already wired up by startFakeNode.

func (f *fakeBaseNodeServer) GetTipInfo(ctx context.Context, in *tari_generated.Empty) (*tari_generated.TipInfoResponse, error) {
	return &tari_generated.TipInfoResponse{
		Metadata: &tari_generated.MetaData{
			BestBlockHeight: 42,
		},
		InitialSyncAchieved: true,
	}, nil
}

func (f *fakeBaseNodeServer) SubmitBlock(ctx context.Context, in *tari_generated.Block) (*tari_generated.SubmitBlockResponse, error) {
	return &tari_generated.SubmitBlockResponse{
		BlockHash: []byte("fake-block-hash"),
	}, nil
}

func (f *fakeBaseNodeServer) GetNewBlockTemplateWithCoinbases(ctx context.Context, in *tari_generated.GetNewBlockTemplateWithCoinbasesRequest) (*tari_generated.GetNewBlockResult, error) {
	return &tari_generated.GetNewBlockResult{
		BlockHash: []byte("fake-new-block-hash"),
	}, nil
}

// TestInitNodeGRPC_ReturnsNoErrorForValidAddress verifies InitNodeGRPC's new error return
// value is nil for a syntactically valid address, and that it still populates the
// package-level grpcConn as before.
func TestInitNodeGRPC_ReturnsNoErrorForValidAddress(t *testing.T) {
	if err := InitNodeGRPC("127.0.0.1:18142"); err != nil {
		t.Fatalf("InitNodeGRPC returned unexpected error: %v", err)
	}
	if grpcConn == nil {
		t.Fatal("InitNodeGRPC did not populate the package-level grpcConn")
	}
}

func TestGetTipInfoCtx_Success(t *testing.T) {
	startFakeNode(t)

	resp, err := GetTipInfoCtx(context.Background())
	if err != nil {
		t.Fatalf("GetTipInfoCtx returned unexpected error: %v", err)
	}
	if resp.GetMetadata().GetBestBlockHeight() != 42 {
		t.Errorf("BestBlockHeight = %d, want 42", resp.GetMetadata().GetBestBlockHeight())
	}
	if !resp.GetInitialSyncAchieved() {
		t.Errorf("InitialSyncAchieved = false, want true")
	}
}

// TestGetTipInfoCtx_PropagatesCancellation proves GetTipInfoCtx actually threads the
// caller-supplied ctx through to the underlying GRPC call (rather than silently using
// context.Background() internally): a pre-cancelled ctx must cause the call to fail with a
// Canceled status instead of succeeding.
func TestGetTipInfoCtx_PropagatesCancellation(t *testing.T) {
	startFakeNode(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GetTipInfoCtx(ctx)
	if err == nil {
		t.Fatal("GetTipInfoCtx with a cancelled ctx returned no error, want a Canceled error")
	}
	if !errors.Is(err, context.Canceled) && status.Code(err) != codes.Canceled {
		t.Errorf("GetTipInfoCtx error = %v, want context.Canceled or a Canceled status", err)
	}
}

func TestSubmitBlockCtx_Success(t *testing.T) {
	startFakeNode(t)

	resp, err := SubmitBlockCtx(context.Background(), &tari_generated.Block{})
	if err != nil {
		t.Fatalf("SubmitBlockCtx returned unexpected error: %v", err)
	}
	if string(resp.GetBlockHash()) != "fake-block-hash" {
		t.Errorf("BlockHash = %q, want %q", resp.GetBlockHash(), "fake-block-hash")
	}
}

// TestSubmitBlockCtx_PropagatesCancellation proves SubmitBlockCtx threads the caller-supplied
// ctx through to the underlying GRPC call.
func TestSubmitBlockCtx_PropagatesCancellation(t *testing.T) {
	startFakeNode(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := SubmitBlockCtx(ctx, &tari_generated.Block{})
	if err == nil {
		t.Fatal("SubmitBlockCtx with a cancelled ctx returned no error, want a Canceled error")
	}
	if !errors.Is(err, context.Canceled) && status.Code(err) != codes.Canceled {
		t.Errorf("SubmitBlockCtx error = %v, want context.Canceled or a Canceled status", err)
	}
}

func TestGetNewBlockTemplateWithCoinbasesCtx_Success(t *testing.T) {
	startFakeNode(t)

	resp, err := GetNewBlockTemplateWithCoinbasesCtx(context.Background(), &tari_generated.GetNewBlockTemplateWithCoinbasesRequest{})
	if err != nil {
		t.Fatalf("GetNewBlockTemplateWithCoinbasesCtx returned unexpected error: %v", err)
	}
	if string(resp.GetBlockHash()) != "fake-new-block-hash" {
		t.Errorf("BlockHash = %q, want %q", resp.GetBlockHash(), "fake-new-block-hash")
	}
}

// TestGetNewBlockTemplateWithCoinbasesCtx_PropagatesCancellation proves
// GetNewBlockTemplateWithCoinbasesCtx threads the caller-supplied ctx through to the
// underlying GRPC call.
func TestGetNewBlockTemplateWithCoinbasesCtx_PropagatesCancellation(t *testing.T) {
	startFakeNode(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := GetNewBlockTemplateWithCoinbasesCtx(ctx, &tari_generated.GetNewBlockTemplateWithCoinbasesRequest{})
	if err == nil {
		t.Fatal("GetNewBlockTemplateWithCoinbasesCtx with a cancelled ctx returned no error, want a Canceled error")
	}
	if !errors.Is(err, context.Canceled) && status.Code(err) != codes.Canceled {
		t.Errorf("GetNewBlockTemplateWithCoinbasesCtx error = %v, want context.Canceled or a Canceled status", err)
	}
}
