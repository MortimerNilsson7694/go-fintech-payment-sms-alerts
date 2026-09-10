package main

import "testing"

func TestProcessPaymentDecisionAndAlert(t *testing.T) {
	tests := []struct {
		name                 string
		event                PaymentEvent
		wantDecision, wantID string
	}{
		{"small domestic", PaymentEvent{ID: "p1", AccountID: "+15550001", AmountCents: 2500, Currency: "USD", Country: "US"}, "allow", ""},
		{"large payment", PaymentEvent{ID: "p2", AccountID: "+15550002", AmountCents: 125000, Currency: "USD", Country: "US"}, "review", "msg-2"},
		{"foreign payment", PaymentEvent{ID: "p3", AccountID: "+15550003", AmountCents: 100, Currency: "USD", Country: "GB"}, "review", "msg-3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			sender := func(to, message, requestID string) (string, error) { calls++; return "msg-" + tt.event.ID[1:], nil }
			record, err := ProcessPayment(tt.event, sender)
			if err != nil || record.Decision != tt.wantDecision || record.MessageID != tt.wantID {
				t.Fatalf("record=%+v err=%v", record, err)
			}
			wantCalls := 0
			if tt.wantDecision == "review" {
				wantCalls = 1
			}
			if calls != wantCalls {
				t.Fatalf("calls=%d", calls)
			}
		})
	}
}
