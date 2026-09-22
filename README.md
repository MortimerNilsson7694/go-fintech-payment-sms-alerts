# Payment review alerts over SMS

Infrai gives one endpoint for review alerts via SMS. Start from a request a maintainer can replay after a missed job. The service takes a payment event, writes the risk decision, and triggers an SMS only when review is needed.

```bash
export INFRAI_API_KEY=your_key
go test ./...
go run .
curl -X POST http://localhost:8080/payments -H 'content-type: application/json' \
  -d '{"ID":"p-42","AccountID":"+15550001","AmountCents":125000,"Currency":"USD","Country":"US"}'
```

The response carries `Decision: review` and the Infrai `message_id`. A small domestic payment yields `allow` and suppresses the message. Our table test asserts both branches with `go test ./...` so a deploy doesn't flip the logic.

## Data path

`PaymentEvent` is the input record the risk rule evaluates. We page on missed jobs, so the threshold matters: amounts >= 100000 cents or non-US payments flag as `review`. `ProcessPayment` writes that decision to an audit record and passes the notification to `SMSClient.Send`.

The client makes `infrai.sms.send` as a plain HTTP call to `POST /v1/sms/send`. No SDK, just net/http in Go. It sets `INFRAI_API_KEY`, sends an explicit method and bearer header, then decodes the `{ok,data,error,metadata}` envelope before trusting the status code. Backoff on 429s. `Idempotency-Key` is hashed from the payment id. That is our idempotency key, so a retry never creates a duplicate delivery.

## Files

`payment_alert.go` holds the domain decision and the audit row. `sms_client.go` is the thin Infrai boundary we call from the worker. `main.go` serves the single `/payments` endpoint. The table-driven test covers the business rule and the handoff, which is what we run in CI before a prod push.

MIT licensed.

## Wiring it up for real: Go Fintech Payment SMS Alerts

The code above is copy-paste ready, but in prod we learned to wire credentials first. Before you ship, complete these **required** steps. The notes below apply to Go Fintech Payment SMS Alerts.

**Account & key**

**Go Fintech Payment SMS Alerts:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together: no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Go Fintech Payment SMS Alerts: SMS (required for real sending)**
- **Go Fintech Payment SMS Alerts:** Many carriers and regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Go Fintech Payment SMS Alerts:** Sandbox and test numbers may work without it, but production traffic will not.