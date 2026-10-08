// Copyright 2026 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/urfave/cli/v2"
)

func TestTransactionHistoryConfig(t *testing.T) {
	defaults := ethconfig.Defaults.TransactionHistory
	tests := []struct {
		name   string
		config string
		args   []string
		want   uint64
	}{
		{name: "full default", want: defaults},
		{name: "archive default", args: []string{"--gcmode=archive"}},
		{name: "archive CLI bounded", args: []string{"--gcmode=archive", "--history.transactions=250000"}, want: 250000},
		{name: "archive CLI unlimited", args: []string{"--gcmode=archive", "--history.transactions=0"}},
		{name: "archive CLI default value", args: []string{"--gcmode=archive", fmt.Sprintf("--history.transactions=%d", defaults)}, want: defaults},
		{name: "archive TOML default", config: "[Eth]\nNoPruning = true\n"},
		{name: "archive TOML bounded", config: "[Eth]\nNoPruning = true\nTransactionHistory = 250000\n", want: 250000},
		{name: "archive inline TOML bounded", config: "Eth = { NoPruning = true, TransactionHistory = 250000 }\n", want: 250000},
		{name: "archive TOML unlimited", config: "[Eth]\nNoPruning = true\nTransactionHistory = 0\n"},
		{name: "archive TOML default value", config: fmt.Sprintf("[Eth]\nNoPruning = true\nTransactionHistory = %d\n", defaults), want: defaults},
		{name: "CLI archive TOML bounded", config: "[Eth]\nTransactionHistory = 250000\n", args: []string{"--gcmode=archive"}, want: 250000},
		{name: "CLI bounded overrides TOML", config: "[Eth]\nNoPruning = true\nTransactionHistory = 250000\n", args: []string{"--history.transactions=50000"}, want: 50000},
		{name: "CLI unlimited overrides TOML", config: "[Eth]\nNoPruning = true\nTransactionHistory = 250000\n", args: []string{"--history.transactions=0"}},
		{name: "CLI bounded overrides TOML unlimited", config: "[Eth]\nNoPruning = true\nTransactionHistory = 0\n", args: []string{"--history.transactions=50000"}, want: 50000},
		{name: "CLI full overrides TOML archive", config: "[Eth]\nNoPruning = true\n", args: []string{"--gcmode=full"}, want: defaults},
		{name: "transaction history shorter than state history", args: []string{"--gcmode=archive", "--history.transactions=50000", "--history.state=250000"}, want: 50000},
		{name: "full TOML bounded", config: "[Eth]\nTransactionHistory = 250000\n", want: 250000},
		{name: "legacy TOML bounded", config: "[Eth]\nTxLookupLimit = 50000\n", want: 50000},
		{name: "modern TOML overrides legacy", config: fmt.Sprintf("[Eth]\nTransactionHistory = %d\nTxLookupLimit = 50000\n", defaults), want: defaults},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"geth", "--datadir", t.TempDir(), "--state.scheme=path"}
			if tt.config != "" {
				path := filepath.Join(t.TempDir(), "config.toml")
				if err := os.WriteFile(path, []byte(tt.config), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--config", path)
			}
			args = append(args, tt.args...)
			app := &cli.App{
				Flags: slices.Concat(nodeFlags, rpcFlags),
				Action: func(ctx *cli.Context) error {
					stack, cfg := makeConfigNode(ctx)
					defer stack.Close()
					if cfg.Eth.TransactionHistory != tt.want {
						t.Errorf("TransactionHistory = %d, want %d", cfg.Eth.TransactionHistory, tt.want)
					}
					return nil
				},
			}
			if err := app.Run(args); err != nil {
				t.Fatal(err)
			}
		})
	}
}
