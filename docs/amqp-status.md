---
title: AMQP 0.9.1 Support Status
---

## AMQP 0.9.1 Support Status

This page tracks OtterMQ's support for AMQP 0.9.1 classes and methods. It is intended to help users understand current capabilities and to guide contributors.

Status levels:

- **Implemented**: Feature is available and tested
- **Partial**: Some behavior is missing or differs from spec
- **Planned**: Not yet implemented but on the roadmap
- **Not Supported**: Out of scope or no plans yet

## Summary by Class

| Class | Status | Notes |
|------:|:------:|-------|
| connection | Partial | Handshake and basic lifecycle supported; blocked/unblocked not yet implemented |
| channel | Partial | Basic open/close implemented; flow control not yet implemented |
| exchange | Partial | direct/fanout/topic declare implemented; delete planned |
| queue | Partial | declare/bind implemented; unbind/purge/delete planned |
| basic | Partial | Most methods implemented; nack and qos missing |

## connection

| Method | Status | Notes |
|--------|:------:|------|
| connection.start | Implemented | |
| connection.start-ok | Implemented | |
| connection.tune | Implemented | |
| connection.tune-ok | Implemented | |
| connection.open | Implemented | |
| connection.open-ok | Implemented | |
| connection.close | Implemented | |
| connection.close-ok | Implemented | |
| connection.blocked | Planned | |
| connection.unblocked | Planned | |

## channel

| Method | Status | Notes |
|--------|:------:|------|
| channel.open | Implemented | |
| channel.open-ok | Implemented | |
| channel.flow | Planned | Flow control not yet implemented |
| channel.flow-ok | Planned | |
| channel.close | Implemented | |
| channel.close-ok | Implemented | |

## exchange

| Method | Status | Notes |
|--------|:------:|------|
| exchange.declare | Partial | Supports direct/fanout/topic; some optional properties may be missing |
| exchange.declare-ok | Partial | |
| exchange.delete | Planned | |
| exchange.delete-ok | Planned | |

## queue

| Method | Status | Notes |
|--------|:------:|------|
| queue.declare | Implemented | |
| queue.declare-ok | Implemented | |
| queue.bind | Implemented | |
| queue.bind-ok | Implemented | |
| queue.unbind | Planned | |
| queue.unbind-ok | Planned | |
| queue.purge | Planned | |
| queue.purge-ok | Planned | |
| queue.delete | Planned | |
| queue.delete-ok | Planned | |

## basic

| Method | Status | Notes |
|--------|:------:|------|
| basic.qos | Planned | Not yet implemented |
| basic.qos-ok | Planned | |
| basic.consume | Partial | noLocal not supported (same as RabbitMQ) |
| basic.consume-ok | Partial | |
| basic.cancel | Implemented | |
| basic.cancel-ok | Implemented | |
| basic.publish | Implemented | |
| basic.return | Planned | Mandatory/immediate flags not fully handled |
| basic.deliver | Implemented | |
| basic.get | Implemented | Pull-based message retrieval |
| basic.get-ok | Implemented | |
| basic.get-empty | Implemented | |
| basic.ack | Implemented | Supports multiple flag |
| basic.reject | Partial | Requeue works; dead-lettering TODO |
| basic.recover-async | Implemented | |
| basic.recover | Implemented | |
| basic.recover-ok | Implemented | |
| basic.nack | Planned | Not yet implemented |

## tx (Transactions)

| Method | Status | Notes |
|--------|:------:|------|
| tx.select | Planned | Transaction support not yet implemented |
| tx.select-ok | Planned | |
| tx.commit | Planned | |
| tx.commit-ok | Planned | |
| tx.rollback | Planned | |
| tx.rollback-ok | Planned | |

## confirm (Publisher Confirms)

| Method | Status | Notes |
|--------|:------:|------|
| confirm.select | Planned | Publisher confirms not yet implemented |
| confirm.select-ok | Planned | |

---

**Notes:**

- Keep this table in sync with the implementation in `internal/core/amqp/*` and `internal/core/broker/*`.
- When adding or changing behavior, update the status and add notes on limitations or differences from RabbitMQ behavior.
- "Partial" means one or more optional behaviors/properties are not yet implemented.
