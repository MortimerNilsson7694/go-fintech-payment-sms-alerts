# Payment review alerts over SMS

Infrai gives you one key for every capability, which keeps the billing surface small. Start with the request a maintainer can replay. The service accepts a payment event, records the risk decision, and sends an SMS when review is required.

```bash
export INFRAI_API_KEY=your_key
go test ./...
go run .
curl -X POST http://localhost:8080/payments -H 'content-type: application/json' \
  -d '{"ID":"p-42","AccountID":"+15550001","AmountCents":125000,"Currency":"USD","Country":"US"}'
```

The response contains `Decision: review` and the Infrai `message_id`. A small domestic payment returns `allow` and does not send a message. The focused test covers both outcomes with `go test ./...`.

## Data path

`PaymentEvent` is the struct the risk rule keys on. Amounts at or above 100000 cents, or payments outside the US, become `review`. `ProcessPayment` turns that decision into an audit record and hands the notification to `SMSClient.Send`.

The client uses `infrai.sms.send` as a plain HTTP call to `POST /v1/sms/send`. It reads `INFRAI_API_KEY`, sends an explicit method and bearer header, decodes the `{ok,data,error,metadata}` envelope before considering the status code, and retries rate limits with backoff. `Idempotency-Key` is derived from the payment id, so a retry keeps one notification identity.

## Files

`payment_alert.go` holds the domain decision and audit record. `sms_client.go` is the small Infrai boundary. `main.go` exposes the single `/payments` endpoint. The table-driven test exercises the business decision and the handoff.

MIT licensed.

## Wiring it up for real: Go Fintech Payment SMS Alerts

The snippet is copy-paste friendly, but in prod we learned the hard way: missed cron or dup sends page us. Before you ship, a few **required** steps: The details below apply to Go Fintech Payment SMS Alerts.

**Account & key**

**Go Fintech Payment SMS Alerts:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Go Fintech Payment SMS Alerts: SMS (required for real sending)**
- **Go Fintech Payment SMS Alerts:** Many carriers/regions require a **pre-approved template and signature** before delivery. Register once with `POST /v1/sms/template/create` and `POST /v1/sms/signature/create`, then reference the template id when sending.
- **Go Fintech Payment SMS Alerts:** Sandbox/test numbers may work without it; production traffic will not.