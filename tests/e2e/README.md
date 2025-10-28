# End-to-End Tests

This directory contains end-to-end tests that verify OtterMQ's AMQP protocol implementation by using real AMQP client connections.

## Prerequisites

Before running these tests, you need to have the OtterMQ broker running:

```bash
# Terminal 1: Start the broker
## Prerequisites

You usually do NOT need to start a separate OtterMQ broker manually. The e2e test suite includes a `TestMain` in `tests/e2e/setup_test.go` that will programmatically create and start a broker instance for the duration of the tests. The tests use a temporary data directory and clean up after themselves.

If you prefer to run tests against an already-running broker (manual mode), start the broker first as shown below and then run the tests. This is optional:

```bash
# Terminal 1: Start the broker (optional)
make run

# Or if you need to rebuild first
make build && make run
```
make run

# Or if you need to rebuild first
make build && make run
```

## Running the Tests

With the broker running in the background:

```bash
# Run all e2e tests
Run the tests. By default the suite will start and stop a test broker automatically.

```bash
# Run all e2e tests (automatic broker startup)
go test -v ./tests/e2e/...

# Run a specific test
go test -v ./tests/e2e -run TestQoS_PerConsumer_PrefetchLimit

# Run with race detector
go test -v -race ./tests/e2e/...
```
go test -v ./tests/e2e/...

# Run specific test

go test -v -race ./tests/e2e/...




   - Sets prefetch count to 3
   - Publishes 10 messages
   - Verifies only 3 messages are delivered initially
When all tests pass, you should see similar output with PASS results for the QoS tests.
   - Acks one message and verifies another is delivered

2. **TestQoS_Global_ChannelLimit**: Verifies channel-wide prefetch limits (global=true)
**Connection refused errors**: By default tests start an embedded broker. If you are running tests against a manually started broker, ensure it is listening on port 5672 and credentials match `guest/guest` (or adjust the test setup accordingly).

**Test timeouts**: Check broker logs for errors or panics. If your machine is slow or overloaded, increase timeouts in the tests.

**Flaky tests**: Some tests use timing assumptions; adjust timeouts if running on slow hardware.
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
