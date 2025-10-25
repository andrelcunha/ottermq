# End-to-End Tests

This directory contains end-to-end tests that verify OtterMQ's AMQP protocol implementation by using real AMQP client connections.

## Prerequisites

Before running these tests, you need to have the OtterMQ broker running:

```bash
# Terminal 1: Start the broker
make run

# Or if you need to rebuild first
make build && make run
```

## Running the Tests

With the broker running in the background:

```bash
# Run all e2e tests
go test -v ./tests/e2e/...

# Run specific test
go test -v ./tests/e2e -run TestQoS_PerConsumer_PrefetchLimit

# Run with race detector
go test -v -race ./tests/e2e/...
```

## Test Coverage

### QoS Tests (`qos_test.go`)

These tests verify the `basic.qos` implementation:

1. **TestQoS_PerConsumer_PrefetchLimit**: Verifies per-consumer prefetch limits (global=false)
   - Sets prefetch count to 3
   - Publishes 10 messages
   - Verifies only 3 messages are delivered initially
   - Acks one message and verifies another is delivered

2. **TestQoS_Global_ChannelLimit**: Verifies channel-wide prefetch limits (global=true)
   - Sets global prefetch to 5
   - Creates 2 consumers on the same channel
   - Verifies total unacked messages across both consumers is limited to 5

3. **TestQoS_MultipleAck**: Verifies that `multiple=true` in ack releases multiple slots
   - Sets prefetch to 2
   - Receives 2 messages
   - Uses `Ack(multiple=true)` to ack both at once
   - Verifies 2 more messages are delivered

4. **TestQoS_Nack_WithRequeue**: Verifies nack with requeue respects QoS limits
   - Nacks a message with requeue=true
   - Verifies the redelivered flag is set correctly
   - Verifies QoS limits still apply to requeued messages

5. **TestQoS_ZeroPrefetch_Unlimited**: Verifies prefetch=0 means unlimited
   - Sets prefetch to 0
   - Publishes 50 messages
   - Verifies all 50 are delivered without throttling

6. **TestQoS_ChangeLimit_MidConsume**: Verifies changing QoS mid-consumption
   - Starts with prefetch=2
   - Changes to prefetch=5 while consuming
   - Verifies the new limit is applied immediately

## Test Expectations

- Broker must be running on `localhost:5672`
- Default credentials: `guest/guest`
- Tests create temporary queues with auto-delete enabled
- Each test is independent and cleans up after itself

## Troubleshooting

**Connection refused errors**: Make sure the broker is running on port 5672

**Test timeouts**: Check broker logs for errors or panics

**Flaky tests**: Some tests use timing assumptions; adjust timeouts if running on slow hardware

## Adding New Tests

When adding new e2e tests:

1. Use unique queue names to avoid conflicts (include test name in queue name)
2. Use auto-delete queues to avoid manual cleanup
3. Set reasonable timeouts (1-2 seconds for most operations)
4. Log intermediate steps with `t.Logf()` for debugging
5. Clean up connections with `defer conn.Close()`
