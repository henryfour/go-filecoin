package porcelain

import (
	"context"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-filecoin/internal/pkg/block"
	"github.com/filecoin-project/go-filecoin/internal/pkg/types"
	"github.com/filecoin-project/go-filecoin/internal/pkg/util/moresync"
)

type waitPlumbing interface {
	MessageWait(context.Context, cid.Cid, func(*block.Block, *types.SignedMessage, *types.MessageReceipt) error) error
}

// MessageWaitDone blocks until the given message cid appears on chain
func MessageWaitDone(ctx context.Context, plumbing waitPlumbing, msgCid cid.Cid) (*types.MessageReceipt, error) {
	l := moresync.NewLatch(1)
	var ret *types.MessageReceipt
	err := plumbing.MessageWait(ctx, msgCid, func(_ *block.Block, _ *types.SignedMessage, rcpt *types.MessageReceipt) error {
		ret = rcpt
		l.Done()
		return nil
	})
	if err != nil {
		return nil, err
	}
	l.Wait()
	return ret, nil
}
