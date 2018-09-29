package net

import (
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/monitor"
	"github.com/vitelabs/go-vite/vite/net/message"
	"sync/atomic"
)

type Fetcher interface {
	// start is required, because we need start + count to find appropriate peer
	FetchSnapshotBlocks(start uint64, count uint64, hash *types.Hash)

	// address is optional
	FetchAccountBlocks(start types.Hash, count uint64, address *types.Address)
}

type fetcher struct {
	filter   Filter
	peers    *peerSet
	receiver Receiver
	pool     RequestPool
	ready    int32 // atomic
	log      log15.Logger
}

func newFetcher(filter Filter, peers *peerSet, receiver Receiver, pool RequestPool) *fetcher {
	return &fetcher{
		filter:   filter,
		peers:    peers,
		receiver: receiver,
		pool:     pool,
		log:      log15.New("module", "net/fetcher"),
	}
}

func (f *fetcher) FetchSnapshotBlocks(start uint64, count uint64, hash *types.Hash) {
	monitor.LogEvent("net/fetch", "s")

	if atomic.LoadInt32(&f.ready) == 0 {
		f.log.Warn("not ready")
		return
	}

	m := &message.GetSnapshotBlocks{
		From:    &ledger.HashHeight{start, *hash},
		Count:   count,
		Forward: true,
	}

	peers := f.peers.Pick(start + count)
	if len(peers) != 0 {
		p := peers[0]
		id := f.pool.MsgID()
		err := p.Send(GetAccountBlocksCode, id, m)
		if err != nil {
			f.log.Error(fmt.Sprintf("send %s to %s error: %v", GetAccountBlocksCode, p, err))
		} else {
			f.log.Info(fmt.Sprintf("send %s to %s done", GetAccountBlocksCode, p))
		}
	} else {
		f.log.Error(errNoPeer.Error())
	}
}

func (f *fetcher) FetchAccountBlocks(start types.Hash, count uint64, address *types.Address) {
	monitor.LogEvent("net/fetch", "a")

	if atomic.LoadInt32(&f.ready) == 0 {
		f.log.Warn("not ready")
		return
	}

	addr := NULL_ADDRESS
	if address != nil {
		addr = *address
	}
	m := &message.GetAccountBlocks{
		Address: addr,
		From: &ledger.HashHeight{
			Hash: start,
		},
		Count:   count,
		Forward: true,
	}

	p := f.peers.BestPeer()
	if p != nil {
		id := f.pool.MsgID()
		err := p.Send(GetAccountBlocksCode, id, m)
		if err != nil {
			f.log.Error(fmt.Sprintf("send %s to %s error: %v", GetAccountBlocksCode, p, err))
		} else {
			f.log.Info(fmt.Sprintf("send %s to %s done", GetAccountBlocksCode, p))
		}
	} else {
		f.log.Error(errNoPeer.Error())
	}
}

func (f *fetcher) listen(st SyncState) {
	if st == Syncdone || st == SyncDownloaded {
		atomic.StoreInt32(&f.ready, 1)
	}
}
