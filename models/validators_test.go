package models

import "testing"

func TestEthcallVal(t *testing.T) {
	archiveNode := "http://localhost:8545"
	blkNumHex := "0x11"
	valAddr := "0x086119bd018ed4940e7427b9373c014f7b754ad5"

	valInfo, err := jsonrpcEthCallGetValInfo(archiveNode, blkNumHex, valAddr)
	if err != nil {
		t.Error(err)
	}
	t.Log(valInfo)
}

func TestGetActVals(t *testing.T) {
	archiveNode := "http://localhost:8545"
	blkNumHex := "0x11"
	vals, err := jsonrpcEthCallGetActVals(archiveNode, blkNumHex)
	if err != nil {
		t.Error(err)
	}
	t.Log(vals)
}

func TestGetAllVals(t *testing.T) {
	archiveNode := "http://localhost:8545"
	blkNumHex := "0x110"
	vals, err := rpcCongressGetAllVals(archiveNode, blkNumHex)
	if err != nil {
		t.Error(err)
	}
	t.Log(vals)
}