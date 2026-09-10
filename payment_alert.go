package main

import "fmt"

type PaymentEvent struct {
	ID          string
	AccountID   string
	AmountCents int64
	Currency    string
	Country     string
}

type AuditRecord struct {
	PaymentID string
	Decision  string
	MessageID string
}

func EvaluatePayment(p PaymentEvent) string {
	if p.AmountCents >= 100000 || p.Country != "US" {
		return "review"
	}
	return "allow"
}

func ProcessPayment(p PaymentEvent, sender func(string, string, string) (string, error)) (AuditRecord, error) {
	decision := EvaluatePayment(p)
	record := AuditRecord{PaymentID: p.ID, Decision: decision}
	if decision != "review" {
		return record, nil
	}
	message := fmt.Sprintf("Payment %s needs review: %s %d.%02d", p.ID, p.Currency, p.AmountCents/100, p.AmountCents%100)
	id, err := sender(p.AccountID, message, "payment-"+p.ID)
	if err != nil {
		return record, err
	}
	record.MessageID = id
	return record, nil
}
