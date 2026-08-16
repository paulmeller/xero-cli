package sync

import "testing"

// Regression for "bank-transfers missing from StreamRegistry" - previously absent entirely, so
// `xero bank-transfers list`/`get` never hit the local sync cache (always live API) and `xero
// sync run` couldn't sync the stream at all.
func TestStreamRegistry_HasBankTransfers(t *testing.T) {
	meta, ok := StreamRegistry["bank_transfers"]
	if !ok {
		t.Fatal("StreamRegistry missing \"bank_transfers\" entry")
	}
	if meta.APIPath != "BankTransfers" {
		t.Errorf("APIPath = %q, want %q", meta.APIPath, "BankTransfers")
	}
	if meta.JSONKey != "BankTransfers" {
		t.Errorf("JSONKey = %q, want %q", meta.JSONKey, "BankTransfers")
	}
	if meta.PrimaryKey != "BankTransferID" {
		t.Errorf("PrimaryKey = %q, want %q", meta.PrimaryKey, "BankTransferID")
	}
}

func TestStreamPriority_HasBankTransfers(t *testing.T) {
	for _, name := range StreamPriority {
		if name == "bank_transfers" {
			return
		}
	}
	t.Error("StreamPriority missing \"bank_transfers\"")
}
