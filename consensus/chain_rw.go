package consensus

import (
	"math/big"

	"time"

	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/vm/contracts"
)

type ch interface {
	GetLatestSnapshotBlock() *ledger.SnapshotBlock
	GetConsensusGroupList(snapshotHash types.Hash) []*contracts.ConsensusGroupInfo                                                 // 获取所有的共识组
	GetRegisterList(snapshotHash types.Hash, gid types.Gid) []*contracts.Registration                                              // 获取共识组下的参与竞选的候选人
	GetVoteMap(snapshotHash types.Hash, gid types.Gid) []*contracts.VoteInfo                                                       // 获取候选人的投票
	GetBalanceList(snapshotHash types.Hash, tokenTypeId types.TokenTypeId, addressList []types.Address) map[types.Address]*big.Int // 获取所有用户的余额
	GetSnapshotBlockBeforeTime(timestamp *time.Time) (*ledger.SnapshotBlock, error)
	GetContractGidByAccountBlock(block *ledger.AccountBlock) (*types.Gid, error)
}
type chainRw struct {
	rw ch
}

type Vote struct {
	name    string
	addr    types.Address
	balance *big.Int
}

func (self *chainRw) GetSnapshotBeforeTime(t time.Time) (*ledger.SnapshotBlock, error) {
	block, e := self.rw.GetSnapshotBlockBeforeTime(&t)

	if e != nil {
		return nil, e
	}

	if block == nil {
		return nil, errors.New("before time[" + t.String() + "] block not exist")
	}
	return block, nil
}

func (self *chainRw) CalVotes(gid types.Gid, info *membersInfo, block ledger.HashHeight) ([]*Vote, error) {

	// query register info
	registerList := self.rw.GetRegisterList(block.Hash, gid)
	// query vote info
	votes := self.rw.GetVoteMap(block.Hash, gid)

	var registers []*Vote

	// cal candidate
	for _, v := range registerList {
		registers = append(registers, self.GenVote(block.Hash, v, votes, info.countingTokenId))
	}
	return registers, nil
}
func (self *chainRw) GenVote(snapshotHash types.Hash, registration *contracts.Registration, infos []*contracts.VoteInfo, id types.TokenTypeId) *Vote {
	var addrs []types.Address
	for _, v := range infos {
		if v.NodeName == registration.Name {
			addrs = append(addrs, v.VoterAddr)
		}
	}
	balanceMap := self.rw.GetBalanceList(snapshotHash, id, addrs)

	result := &Vote{balance: big.NewInt(0), name: registration.Name, addr: registration.NodeAddr}
	for _, v := range balanceMap {
		result.balance.Add(result.balance, v)
	}
	return result
}
func (self *chainRw) GetMemberInfo(gid types.Gid, genesis time.Time) *membersInfo {
	// todo consensus group maybe change ??
	var result *membersInfo
	head := self.rw.GetLatestSnapshotBlock()
	consensusGroupList := self.rw.GetConsensusGroupList(head.Hash)
	for _, v := range consensusGroupList {
		if v.Gid == gid {
			result = &membersInfo{
				genesisTime:     genesis,
				interval:        int32(v.Interval),
				memberCnt:       int32(v.NodeCount),
				seed:            new(big.Int).SetBytes(v.Gid.Bytes()),
				perCnt:          int32(v.PerCount),
				randCnt:         int32(v.RandCount),
				randRange:       int32(v.RandRank),
				countingTokenId: v.CountingTokenId,
			}
		}
	}

	return result
}

func (self *chainRw) getGid(block *ledger.AccountBlock) (types.Gid, error) {
	gid, e := self.rw.GetContractGidByAccountBlock(block)
	return *gid, e
}
func (self *chainRw) GetLatestSnapshotBlock() *ledger.SnapshotBlock {
	return self.rw.GetLatestSnapshotBlock()
}
