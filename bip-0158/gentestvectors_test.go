package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/btcsuite/btcd/btcutil/gcs/builder"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

// The checked-in BIP-158 vectors include serialized blocks and spent-output
// scripts, so filter generation can be checked without a live btcd RPC server.
func TestBasicFiltersMatchCheckedInVectors(t *testing.T) {
	data, err := os.ReadFile("testnet-19.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows [][]json.RawMessage
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 11 {
		t.Fatalf("expected header and 10 vectors, got %d rows", len(rows))
	}

	for _, row := range rows[1:] {
		if len(row) != 8 {
			t.Fatalf("expected 8 fields, got %d", len(row))
		}
		var height uint32
		var blockHash, blockHex, prevHeader, filterHex, header string
		var prevScriptHex []string
		for i, dest := range []interface{}{&height, &blockHash, &blockHex, &prevScriptHex, &prevHeader, &filterHex, &header} {
			if err := json.Unmarshal(row[i], dest); err != nil {
				t.Fatalf("field %d: %v", i, err)
			}
		}
		t.Run(fmt.Sprint(height), func(t *testing.T) {
			blockBytes, err := hex.DecodeString(blockHex)
			if err != nil {
				t.Fatal(err)
			}
			var block wire.MsgBlock
			if err := block.Deserialize(bytes.NewReader(blockBytes)); err != nil {
				t.Fatal(err)
			}
			if got := block.BlockHash().String(); got != blockHash {
				t.Fatalf("block hash: got %s, want %s", got, blockHash)
			}
			prevScripts := make([][]byte, len(prevScriptHex))
			for i, encoded := range prevScriptHex {
				prevScripts[i], err = hex.DecodeString(encoded)
				if err != nil {
					t.Fatal(err)
				}
			}
			filter, err := builder.BuildBasicFilter(&block, prevScripts)
			if err != nil {
				t.Fatal(err)
			}
			encodedFilter, err := filter.NBytes()
			if err != nil {
				t.Fatal(err)
			}
			if got := hex.EncodeToString(encodedFilter); got != filterHex {
				t.Fatalf("filter: got %s, want %s", got, filterHex)
			}
			previous, err := chainhash.NewHashFromStr(prevHeader)
			if err != nil {
				t.Fatal(err)
			}
			computed, err := builder.MakeHeaderForFilter(filter, *previous)
			if err != nil {
				t.Fatal(err)
			}
			if got := computed.String(); got != header {
				t.Fatalf("filter header: got %s, want %s", got, header)
			}
		})
	}
}
