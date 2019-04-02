package chain_state

import (
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vitelabs/go-vite/chain/utils"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"math/big"
)

// TODO
func (sDB *StateDB) Rollback(deletedSnapshotSegments []*ledger.SnapshotChunk) error {
	if len(deletedSnapshotSegments) <= 0 {
		return nil
	}
	batch := sDB.store.NewBatch()
	//blockHashList := make([]*types.Hash, 0, size)

	allBalanceMap := make(map[types.Address]map[types.TokenTypeId]*big.Int)
	getBalance := func(addr types.Address, tokenTypeId types.TokenTypeId) (*big.Int, error) {
		balanceMap, ok := allBalanceMap[addr]
		if !ok {
			balanceMap = make(map[types.TokenTypeId]*big.Int)
			allBalanceMap[addr] = balanceMap
		}

		balance, ok := balanceMap[tokenTypeId]
		if !ok {
			var err error
			balance, err = sDB.chain.GetBalance(addr, tokenTypeId)
			if err != nil {
				return nil, err
			}
			balanceMap[tokenTypeId] = balance

		}
		return balance, nil
	}
	firstSb := deletedSnapshotSegments[0].SnapshotBlock
	isDeleteSnapshotBlock := false

	//rollbackStorageKeys := make(map[types.Address][]byte)
	for _, seg := range deletedSnapshotSegments {
		snapshotBlock := seg.SnapshotBlock

		kvLogMap := make(map[types.Hash][]byte, 0)
		if snapshotBlock != nil {
			isDeleteSnapshotBlock = true
			if snapshotBlock.Hash != firstSb.Hash {
				// rollback record
				var err error
				kvLogMap, err = sDB.storageRedo.QueryLog(snapshotBlock.Height)
				if err != nil {
					return err
				}
				//sDB.storageRedo.Rollback(snapshotBlock.Height)
			}
		}

		deleteHistoryBalanceKeys := make(map[string]struct{})

		for _, accountBlock := range seg.AccountBlocks {
			if kvLog, ok := kvLogMap[accountBlock.Hash]; ok {
				var kvList [][2]byte
				if err := rlp.DecodeBytes(kvLog, kvList); err != nil {
					return err
				}

				// TODO
			}
			//for hash, kvLog := range kvLogMap {
			//	var kvList [][2]byte
			//	if err := rlp.DecodeBytes(kvLog, kvList); err != nil {
			//		return err
			//	}
			//
			//	for  := range kvList {
			//
			//	}
			//}

			// rollback balance
			addr := accountBlock.AccountAddress
			tokenId := accountBlock.TokenId

			var sendBlock *ledger.AccountBlock

			if accountBlock.IsReceiveBlock() {
				sendBlock, err := sDB.chain.GetAccountBlockByHash(accountBlock.FromBlockHash)
				if err != nil {
					return err
				}
				tokenId = sendBlock.TokenId
			}

			balance, err := getBalance(addr, tokenId)
			if err != nil {
				return err
			}

			if accountBlock.IsReceiveBlock() {
				balance.Add(balance, sendBlock.Amount)
			} else {
				balance.Sub(balance, accountBlock.Amount)
			}

			if accountBlock.Fee != nil {
				balance.Add(balance, accountBlock.Fee)
			}

			allBalanceMap[addr][tokenId] = balance

			// delete history balance
			if snapshotBlock != nil {
				deleteHistoryBalanceKeys[string(chain_utils.CreateHistoryBalanceKey(addr, tokenId, snapshotBlock.Height))] = struct{}{}
			}

			// delete code
			if accountBlock.Height <= 1 {
				batch.Delete(chain_utils.CreateCodeKey(accountBlock.AccountAddress))
			}

			// delete contract meta
			if accountBlock.BlockType == ledger.BlockTypeSendCreate {
				batch.Delete(chain_utils.CreateContractMetaKey(accountBlock.AccountAddress))
			}

			// delete log hash
			if accountBlock.LogHash != nil {
				batch.Delete(chain_utils.CreateVmLogListKey(accountBlock.LogHash))
			}

			// delete call depth
			if accountBlock.IsReceiveBlock() {
				for _, sendBlock := range accountBlock.SendBlockList {
					batch.Delete(chain_utils.CreateCallDepthKey(&sendBlock.Hash))
				}
			}
		}

		for key := range deleteHistoryBalanceKeys {
			batch.Delete([]byte(key))
		}
	}

	// reset index
	for addr, balanceMap := range allBalanceMap {
		for tokenTypeId, balance := range balanceMap {
			balanceBytes := balance.Bytes()
			batch.Put(chain_utils.CreateBalanceKey(addr, tokenTypeId), balanceBytes)
			if !isDeleteSnapshotBlock {
				batch.Put(chain_utils.CreateHistoryBalanceKey(addr, tokenTypeId, sDB.chain.GetLatestSnapshotBlock().Height+1), balanceBytes)
			}
		}
	}

	sDB.store.Write(batch)
	return nil
}

//func (sDB *StateDB) RecoverUnconfirmed(accountBlocks []*ledger.AccountBlock) error {
//	batch := sDB.store.NewBatch()
//
//	for _, accountBlock := range accountBlocks {
//		// rollback storage redo key
//		batch.Delete(chain_utils.CreateStorageRedoKey(accountBlock.Hash))
//
//		// rollback balance
//		addr := accountBlock.AccountAddress
//		tokenId := accountBlock.TokenId
//
//		var sendBlock *ledger.AccountBlock
//
//		if accountBlock.IsReceiveBlock() {
//			sendBlock, err := sDB.chain.GetAccountBlockByHash(accountBlock.FromBlockHash)
//			if err != nil {
//				return err
//			}
//			tokenId = sendBlock.TokenId
//		}
//		balance, err := getBalance(addr, tokenId)
//		if err != nil {
//			return err
//		}
//		if accountBlock.IsReceiveBlock() {
//			balance.Add(balance, sendBlock.Amount)
//		} else {
//			balance.Sub(balance, accountBlock.Amount)
//
//		}
//		allBalanceMap[addr][tokenId] = balance
//
//		// delete history balance
//		if snapshotBlock != nil {
//			deleteKey[string(chain_utils.CreateHistoryBalanceKey(addr, tokenId, snapshotBlock.Height))] = struct{}{}
//		}
//
//		// delete code
//		if accountBlock.Height <= 1 {
//			batch.Delete(chain_utils.CreateCodeKey(accountBlock.AccountAddress))
//		}
//
//		// delete contract meta
//		if accountBlock.BlockType == ledger.BlockTypeSendCreate {
//			batch.Delete(chain_utils.CreateContractMetaKey(accountBlock.AccountAddress))
//		}
//
//		// delete log hash
//		if accountBlock.LogHash != nil {
//			batch.Delete(chain_utils.CreateVmLogListKey(accountBlock.LogHash))
//		}
//
//		// delete call depth
//		if accountBlock.IsReceiveBlock() {
//			for _, sendBlock := range accountBlock.SendBlockList {
//				batch.Delete(chain_utils.CreateCallDepthKey(&sendBlock.Hash))
//			}
//		}
//	}
//	sDB.store.Write(batch)
//
//}
