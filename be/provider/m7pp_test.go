package provider

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignM7PPBody(t *testing.T) {
	body := []byte(`{"traceId":"f8c3de3d-1fea-4d7c-a8b0-29f63c4c3455","username":"bob12345","gameId":1,"language":"zh","platform":"web","currency":"CNY"}`)
	want := "d6b093eef96f5f2557589bd188f49d030eed994b69612fbaf3690b2b8b897362"
	got := SignM7PPBody(body, "813e9cb10f35c37a059c2761465781275ad641d3cb85436cdd17f08b0a6b50bf")
	if got != want {
		t.Fatalf("signature mismatch: got %s want %s", got, want)
	}
}

func TestIsM7PPWalletNotInitializedError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"undocumented internal error on balance", M7PPError{Op: "cash/balance", Status: "SC_INTERNAL_ERROR"}, true},
		{"documented missing user", M7PPError{Op: "cash/balance", Status: "SC_USER_NOT_EXISTS"}, true},
		{"invalid signature remains fatal", M7PPError{Op: "cash/balance", Status: "SC_INVALID_SIGNATURE"}, false},
		{"internal error on transfer remains fatal", M7PPError{Op: "cash/withdraw", Status: "SC_INTERNAL_ERROR"}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsM7PPWalletNotInitializedError(test.err); got != test.want {
				t.Fatalf("got %v want %v", got, test.want)
			}
		})
	}
}

func TestFetchTransactionsSupportsSingularHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
          "status":"SC_OK",
          "data":{
            "header":{"betId":0,"roundId":1,"username":2,"gameCode":3,"betAmount":4,"winAmount":5,"winLoss":6,"effectiveTurnover":7,"status":8,"vendorBetTime":9,"updateTime":10},
            "transactions":[["bet-1","round-1","platu16","BOANZZPP_ppt_live",1.25,2.5,1.25,1.25,1,1785735463000,1785735485000]],
            "currentPage":1,"totalPages":1,"totalItems":1
          }
        }`))
	}))
	defer server.Close()

	client, err := NewM7PPClient(M7PPConfig{BaseURL: server.URL, APIKey: "key", APISecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	items, pages, err := client.FetchTransactions(1, 2, 1, 100)
	if err != nil {
		t.Fatal(err)
	}
	if pages != 1 || len(items) != 1 {
		t.Fatalf("got pages=%d items=%d", pages, len(items))
	}
	item := items[0]
	if item.Username != "platu16" || item.GameCode != "BOANZZPP_ppt_live" || item.BetAmount != 1.25 || item.Status != 1 || item.VendorBetTime != 1785735463000 || item.UpdateTime != 1785735485000 {
		t.Fatalf("unexpected transaction: %+v", item)
	}
}

func TestInt64AtSupportsScientificNotation(t *testing.T) {
	row := []any{"1.785735463e+12", float64(1785735485000)}
	if got, want := int64At(row, 0), int64(1785735463000); got != want {
		t.Fatalf("scientific timestamp got %d want %d", got, want)
	}
	if got, want := int64At(row, 1), int64(1785735485000); got != want {
		t.Fatalf("numeric timestamp got %d want %d", got, want)
	}
}
