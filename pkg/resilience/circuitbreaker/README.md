# `pkg/resilience/circuitbreaker` — Circuit Breaker

Circuit breaker implementation for resilience against cascading failures, used
to protect chain connectors and external service calls.

## States

| State | Behavior |
|-------|----------|
| closed | Requests flow normally; failures are counted. |
| open | Requests fail fast with `ErrCircuitOpen`. |
| half-open | A limited number of trial requests probe recovery. |

## Usage

```go
cb := circuitbreaker.New(circuitbreaker.DefaultConfig())

err := cb.Execute(func() error {
    return chainConnector.SubmitTransaction(ctx, tx)
})
if err == circuitbreaker.ErrCircuitOpen {
    // fail fast: downstream is degraded
}
```

## Configuration

`DefaultConfig` trips open after 5 consecutive failures, stays open for 30s,
then allows 3 half-open trial requests; 2 successes close the circuit.

## Related

- `pkg/multichain/chainregistry` — chain connectors protected by breakers
