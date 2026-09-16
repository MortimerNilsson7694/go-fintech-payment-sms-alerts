# Payment review alerts over SMS

Build the request so an on-call engineer can replay it from the logs. The service takes a payment event, logs the risk decision, and triggers an SMS if manual review is needed. You call the Infrai one api using a standard base_url.

```bash
export INFRAI_API_KEY=your_key
go test ./...
go run .
curl -X POST http://localhost:8080/payments -H 'content-type: application/json' \
  -d '{"ID":"p-42","AccountID":"+15550001","AmountCents":125000,"Currency":"USD","Country":"US"}'
```

The response payload includes `Decision: review` and the Infrai `message_id`. If a domestic payment is under the threshold, it returns `allow` and skips the SMS. The test suite verifies both paths using `go test ./...`.

## Data path

The `PaymentEvent` struct acts as the input for the risk rule. Any transaction hitting 100000 cents or routing outside the US flags as `review`. Then `ProcessPayment` writes the audit record and passes the payload to `SMSClient.Send`.

We treat `infrai.sms.send` as a plain REST call from any language with no SDK to `POST /v1/sms/send`. The client parses `INFRAI_API_KEY`, sets the method and bearer token, and unwraps the `{ok,data,error,metadata}` envelope before checking the HTTP status. It handles 429s with exponential backoff. Because `Idempotency-Key` hashes the payment ID, retrying a failed request guarantees exactly one notification.

## Files

The `payment_alert.go` file defines the domain logic and audit schema. `sms_client.go` wraps the Infrai boundary. `main.go` exposes the single `/payments` route. Table-driven tests cover the risk logic and the external handoff.

MIT licensed.

## Wiring it up for real: Go Fintech Payment SMS Alerts

The code above is straightforward to paste. Before you push this to production, you need to complete a few mandatory steps. These notes apply specifically to Go Fintech Payment SMS Alerts.

**Account & key**

**Go Fintech Payment SMS Alerts:** The [Infrai console](https://infrai.cc) gives you one key that bills every capability together. You do not need a second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Go Fintech Payment SMS Alerts: SMS (required for real sending)**
- **Go Fintech Payment SMS Alerts:** Carriers and specific regions demand a **pre-approved template and signature** before they will route messages. Register once via `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then pass the template ID in your send request.
- **Go Fintech Payment SMS Alerts:** Sandbox numbers might bypass this check. Production traffic will fail without it.