/**
Package vm implements the vite virtual machine
*/
package vm

import (
	"bytes"
	"errors"
	"github.com/vitelabs/go-vite/common/types"
	"math/big"
	"sync/atomic"
)

type VMConfig struct {
	Debug bool
}

type Log struct {
	// list of topics provided by the contract
	Topics []types.Hash
	// supplied by the contract, usually ABI-encoded
	Data []byte
}

type VM struct {
	VMConfig
	StateDb     Database
	createBlock CreateBlockFunc

	abort          int32
	intPool        *intPool
	instructionSet [256]operation
	logList        []*Log
	blockList      []VmBlock
	returnData     []byte
}

func NewVM(stateDb Database, createBlockFunc CreateBlockFunc, config VMConfig) *VM {
	return &VM{StateDb: stateDb, createBlock: createBlockFunc, instructionSet: simpleInstructionSet, logList: make([]*Log, 0), VMConfig: config}
}

func (vm *VM) Run(block VmBlock) (blockList []VmBlock, logList []*Log, err error) {
	switch block.TxType() {
	case TxTypeReceive, TxTypeReceiveError:
		sendBlock := vm.StateDb.AccountBlock(block.From(), block.FromHash())
		block.SetData(sendBlock.Data())
		if sendBlock.TxType() == TxTypeSendCreate {
			return vm.receiveCreate(block, vm.calcCreateQuota(sendBlock.CreateFee()))
		} else if sendBlock.TxType() == TxTypeSendCall {
			return vm.receiveCall(block)
		}
	case TxTypeSendCreate:
		block, err = vm.sendCreate(block)
		if err != nil {
			return nil, nil, err
		} else {
			return []VmBlock{block}, nil, nil
		}
	case TxTypeSendCall:
		block, err = vm.sendCall(block)
		if err != nil {
			return nil, nil, err
		} else {
			return []VmBlock{block}, nil, nil
		}
	}
	return nil, nil, errors.New("transaction type not supported")
}

func (vm *VM) Cancel() {
	atomic.StoreInt32(&vm.abort, 1)
}

// send contract create transaction, create address, sub balance and service fee
func (vm *VM) sendCreate(block VmBlock) (VmBlock, error) {
	// check can make transaction
	quotaTotal, quotaAddition := vm.quotaLeft(block.From(), block)
	quotaLeft := quotaTotal
	quotaRefund := uint64(0)
	cost, err := intrinsicGasCost(block.Data(), false)
	if err != nil {
		return nil, err
	}
	quotaLeft, err = useQuota(quotaLeft, cost)
	if err != nil {
		return nil, err
	}
	if !checkContractFee(block.CreateFee()) {
		return nil, ErrInvalidContractFee
	}
	if !vm.canTransfer(block.From(), block.TokenId(), block.Amount(), block.CreateFee()) {
		return nil, ErrInsufficientBalance
	}
	// create address
	contractAddr, err := createAddress(block.From(), block.Height(), block.Data(), block.SnapshotHash())

	if err != nil || vm.StateDb.IsExistAddress(contractAddr) {
		return nil, ErrContractAddressCreationFail
	}
	// sub balance and service fee
	vm.StateDb.SubBalance(block.From(), block.TokenId(), block.Amount())
	vm.StateDb.SubBalance(block.From(), viteTokenTypeId, block.CreateFee())
	vm.updateBlock(block, block.From(), nil, quotaUsed(quotaTotal, quotaAddition, quotaLeft, quotaRefund, nil), nil)
	block.SetTo(contractAddr)
	return block, nil
}

// receive contract create transaction, create contract account, run initialization code, set contract code, do send blocks
func (vm *VM) receiveCreate(block VmBlock, quotaLeft uint64) (blockList []VmBlock, logList []*Log, err error) {
	if vm.StateDb.IsExistAddress(block.To()) {
		return nil, nil, ErrAddressCollision
	}
	// check can make transaction
	cost, err := intrinsicGasCost(nil, true)
	if err != nil {
		return nil, nil, err
	}
	quotaLeft, err = useQuota(quotaLeft, cost)
	if err != nil {
		return nil, nil, err
	}

	vm.blockList = []VmBlock{block}

	if block.Depth() > callCreateDepth {
		vm.updateBlock(block, block.To(), ErrDepth, 0, nil)
		return vm.blockList, nil, ErrDepth
	}

	// create contract account and add balance
	vm.StateDb.CreateAccount(block.To())
	vm.StateDb.AddBalance(block.To(), block.TokenId(), block.Amount())

	// init contract state and set contract code
	c := newContract(block.From(), block.To(), block, quotaLeft, 0)
	c.setCallCode(block.To(), block.Data())
	code, err := c.run(vm)
	if err == nil {
		codeCost := uint64(len(code)) * contractCodeGas
		c.quotaLeft, err = useQuota(c.quotaLeft, codeCost)
		if err == nil {
			codeHash, _ := types.BytesToHash(code)
			vm.StateDb.SetContractCode(block.To(), code)
			vm.updateBlock(block, block.To(), nil, 0, codeHash.Bytes())
			err = vm.doSendBlockList()
			if err == nil {
				return vm.blockList, vm.logList, nil
			}
		}
	}

	vm.revert()
	vm.StateDb.CreateAccount(block.To())
	vm.updateBlock(block, block.To(), err, 0, nil)
	return vm.blockList, nil, err
}

func (vm *VM) sendCall(block VmBlock) (VmBlock, error) {
	// check can make transaction
	quotaTotal, quotaAddition := vm.quotaLeft(block.From(), block)
	quotaLeft := quotaTotal
	quotaRefund := uint64(0)
	cost, err := intrinsicGasCost(block.Data(), false)
	if err != nil {
		return nil, err
	}
	quotaLeft, err = useQuota(quotaLeft, cost)
	if err != nil {
		return nil, err
	}
	if !vm.canTransfer(block.From(), block.TokenId(), block.Amount(), big0) {
		return nil, ErrInsufficientBalance
	}
	// sub balance
	vm.StateDb.SubBalance(block.From(), block.TokenId(), block.Amount())
	vm.updateBlock(block, block.From(), nil, quotaUsed(quotaTotal, quotaAddition, quotaLeft, quotaRefund, nil), nil)
	return block, nil
}

func (vm *VM) receiveCall(block VmBlock) (blockList []VmBlock, logList []*Log, err error) {
	// check can make transaction
	quotaTotal, quotaAddition := vm.quotaLeft(block.To(), block)
	quotaLeft := quotaTotal
	quotaRefund := uint64(0)
	cost, err := intrinsicGasCost(block.Data(), false)
	if err != nil {
		return nil, nil, err
	}
	quotaLeft, err = useQuota(quotaLeft, cost)
	if err != nil {
		return nil, nil, err
	}
	vm.blockList = []VmBlock{block}
	// create genesis block when accepting first receive transaction
	if !vm.StateDb.IsExistAddress(block.To()) {
		vm.StateDb.CreateAccount(block.To())
	}
	if block.Depth() > callCreateDepth {
		vm.updateBlock(block, block.To(), ErrDepth, quotaUsed(quotaTotal, quotaAddition, quotaLeft, quotaRefund, ErrDepth), nil)
		return vm.blockList, nil, ErrDepth
	}
	vm.StateDb.AddBalance(block.To(), block.TokenId(), block.Amount())
	// do transfer transaction if account code size is zero
	if vm.StateDb.ContractCodeSize(block.To()) == 0 {
		vm.updateBlock(block, block.To(), nil, quotaUsed(quotaTotal, quotaAddition, quotaLeft, quotaRefund, nil), nil)
		return vm.blockList, nil, nil
	}
	// run code
	c := newContract(block.From(), block.To(), block, quotaLeft, quotaRefund)
	c.setCallCode(block.To(), vm.StateDb.ContractCode(block.To()))
	_, err = c.run(vm)
	if err == nil {
		vm.updateBlock(block, block.To(), nil, quotaUsed(quotaTotal, quotaAddition, c.quotaLeft, c.quotaRefund, nil), nil)
		err = vm.doSendBlockList()
		if err == nil {
			return vm.blockList, vm.logList, nil
		}
	}

	vm.revert()
	if !vm.StateDb.IsExistAddress(block.To()) {
		vm.StateDb.CreateAccount(block.To())
	}
	vm.updateBlock(block, block.To(), err, quotaUsed(quotaTotal, quotaAddition, c.quotaLeft, c.quotaRefund, err), nil)
	return vm.blockList, nil, err
}

func (vm *VM) delegateCall(contractAddr types.Address, data []byte, c *contract) (ret []byte, err error) {
	cNew := newContract(c.caller, c.address, c.block, c.quotaLeft, c.quotaRefund)
	cNew.setCallCode(contractAddr, vm.StateDb.ContractCode(contractAddr))
	ret, err = cNew.run(vm)
	c.quotaLeft, c.quotaRefund = cNew.quotaLeft, cNew.quotaRefund
	return ret, err
}

func (vm *VM) calcCreateQuota(fee *big.Int) uint64 {
	// TODO calculate quota for create contract receive transaction
	return quotaLimit
}

func (vm *VM) quotaLeft(addr types.Address, block VmBlock) (quotaInit, quotaAddition uint64) {
	// TODO calculate quota, use max for test
	// TODO calculate quota addition
	quotaInit = maxUint64
	quotaAddition = 0
	for _, block := range vm.blockList {
		if quotaInit <= block.Quota() {
			return 0, 0
		} else {
			quotaInit = quotaInit - block.Quota()
		}
	}
	prevHash := block.PrevHash()
	if len(vm.blockList) > 0 {
		prevHash = vm.blockList[0].PrevHash()
	}
	for {
		prevBlock := vm.StateDb.AccountBlock(addr, prevHash)
		if prevBlock != nil && bytes.Equal(block.SnapshotHash().Bytes(), prevBlock.SnapshotHash().Bytes()) {
			quotaInit = quotaInit - prevBlock.Quota()
		} else {
			if maxUint64-quotaAddition > quotaInit {
				quotaAddition = maxUint64 - quotaInit
				quotaInit = maxUint64
			} else {
				quotaInit = quotaInit + quotaAddition
			}
			return min(quotaInit, quotaLimit), min(quotaAddition, quotaLimit)
		}
	}
}

func quotaUsed(quotaTotal, quotaAddition, quotaLeft, quotaRefund uint64, err error) uint64 {
	if err == ErrOutOfQuota {
		return quotaTotal - quotaAddition
	} else if err != nil {
		return quotaTotal - quotaAddition - quotaLeft
	} else {
		return quotaTotal - quotaAddition - quotaLeft - min(quotaRefund, (quotaTotal-quotaAddition-quotaLeft)/2)
	}
}

func (vm *VM) updateBlock(block VmBlock, addr types.Address, err error, quota uint64, result []byte) {
	block.SetQuota(quota)
	// TODO data = fixed byte of err + result
	block.SetData(result)
	block.SetStateHash(vm.StateDb.StateHash(addr))
	if block.TxType() == TxTypeReceive || block.TxType() == TxTypeReceiveError {
		if err == ErrOutOfQuota || err == ErrContractAddressCreationFail {
			block.SetTxType(TxTypeReceiveError)
		} else {
			block.SetTxType(TxTypeReceive)
		}
		if len(vm.blockList) > 1 {
			for _, sendBlock := range vm.blockList[1:] {
				block.AppendSummaryHash(sendBlock.SummaryHash())
			}
		}
	}
}

func (vm *VM) doSendBlockList() (err error) {
	for i, block := range vm.blockList[1:] {
		if block.To != nil {
			vm.blockList[i], err = vm.sendCall(block)
			if err != nil {
				return err
			}
		} else {
			vm.blockList[i], err = vm.sendCreate(block)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (vm *VM) revert() {
	vm.blockList = vm.blockList[:1]
	vm.logList = nil
	vm.returnData = nil
	vm.StateDb.Revert()
}

func checkContractFee(fee *big.Int) bool {
	return ContractFeeMin.Cmp(fee) <= 0 && ContractFeeMax.Cmp(fee) >= 0
}

// TODO set vite token type id
var viteTokenTypeId = types.TokenTypeId{}

func (vm *VM) canTransfer(addr types.Address, tokenTypeId types.TokenTypeId, tokenAmount *big.Int, feeAmount *big.Int) bool {
	if bytes.Equal(tokenTypeId.Bytes(), viteTokenTypeId.Bytes()) {
		balance := new(big.Int).Add(tokenAmount, feeAmount)
		return balance.Cmp(vm.StateDb.Balance(addr, tokenTypeId)) <= 0
	} else {
		return tokenAmount.Cmp(vm.StateDb.Balance(addr, tokenTypeId)) <= 0 && feeAmount.Cmp(vm.StateDb.Balance(addr, viteTokenTypeId)) <= 0
	}
}

func createAddress(addr types.Address, height *big.Int, code []byte, snapshotHash types.Hash) (types.Address, error) {
	var a types.Address
	dataBytes := append(addr.Bytes(), height.Bytes()...)
	dataBytes = append(dataBytes, code...)
	dataBytes = append(dataBytes, snapshotHash.Bytes()...)
	addressHash := types.DataHash(dataBytes)
	err := a.SetBytes(addressHash[12:])
	return a, err
}
