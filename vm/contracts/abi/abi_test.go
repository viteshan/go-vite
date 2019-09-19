package abi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/vm/abi"
	"strings"
	"testing"
)

func TestContractsABIInit(t *testing.T) {
	tests := []string{jsonPledge, jsonConsensusGroup, jsonMintage}
	for _, data := range tests {
		if _, err := abi.JSONToABIContract(strings.NewReader(data)); err != nil {
			t.Fatalf("json to abi failed, %v, %v", data, err)
		}
	}
}

var (
	t1 = types.TokenTypeId{0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	t2 = types.TokenTypeId{0, 0, 0, 0, 0, 0, 0, 0, 0, 2}
	t3 = types.TokenTypeId{0, 0, 0, 0, 0, 0, 0, 0, 0, 3}
)

func TestDeleteTokenId(t *testing.T) {
	tests := []struct {
		input   []types.TokenTypeId
		tokenId types.TokenTypeId
		output  []types.TokenTypeId
	}{
		{[]types.TokenTypeId{t1}, t1, []types.TokenTypeId{}},
		{[]types.TokenTypeId{t1}, t2, []types.TokenTypeId{t1}},
		{[]types.TokenTypeId{t1, t2}, t1, []types.TokenTypeId{t2}},
		{[]types.TokenTypeId{t1, t2}, t2, []types.TokenTypeId{t1}},
		{[]types.TokenTypeId{t1, t2}, t3, []types.TokenTypeId{t1, t2}},
		{[]types.TokenTypeId{t1, t2, t3}, t1, []types.TokenTypeId{t2, t3}},
		{[]types.TokenTypeId{t1, t2, t3}, t2, []types.TokenTypeId{t1, t3}},
		{[]types.TokenTypeId{t1, t2, t3}, t3, []types.TokenTypeId{t1, t2}},
	}
	for _, test := range tests {
		var idList []byte
		for _, tid := range test.input {
			idList = AppendTokenId(idList, tid)
		}
		result := DeleteTokenId(idList, test.tokenId)
		var target []byte
		for _, tid := range test.output {
			target = AppendTokenId(target, tid)
		}
		if !bytes.Equal(result, target) {
			t.Fatalf("delete token id failed, delete %v from input %v, expected %v, got %v", test.tokenId, test.input, target, result)
		}
	}
}

func TestABIContract_EventById(t *testing.T) {
	for _, e := range ABIMintage.Events {
		fmt.Printf("%v: %v\n", e.Name, e.Id().String())
	}
}

func TestABIContract_MethodById(t *testing.T) {
	for _, m := range ABIConsensusGroup.Methods {
		fmt.Printf("%v: %v\n", m.Sig(), hex.EncodeToString(m.Id()))
	}
}

func TestABIContract_CallbackById(t *testing.T) {
	for _, m := range ABIPledge.Callbacks {
		fmt.Printf("%v: %v\n", m.Sig(), hex.EncodeToString(m.Id()))
	}
}
