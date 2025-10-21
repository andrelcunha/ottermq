---
title: AMQP 0.9.1 Support Status
---

## AMQP 0.9.1 Support Status

This page tracks OtterMQ's support for AMQP 0.9.1 classes and methods. It is intended to help users understand current capabilities and to guide contributors.

Status levels:

- **Implemented** ✅: Feature is available and tested
- **Partial** ⏳: Some behavior is missing or differs from spec
- **Planned** 📕: Not yet implemented but on the roadmap
- **Not Supported** ❌: Out of scope or no plans yet

## Summary by Class

| Class | Status | Notes |
|------:|:------:|-------|
| connection | 100% | Handshake and basic lifecycle supported |
| channel | 67% | Basic open/close implemented; flow control not yet implemented |
| exchange | 80% | direct/fanout declare implemented; missing topic |
| queue | 55% | declare/bind implemented; unbind/purge/delete planned |
| basic | 72% | Most methods implemented; nack and qos missing |
| tx | 0% | |

## connection

| Method | Status | Notes |
|--------|:------:|------|
| connection.start | ✅ | |
| connection.start-ok | ✅ | |
| connection.tune | ✅ | |
| connection.tune-ok | ✅ | |
| connection.open | ✅ | |
| connection.open-ok | ✅ | |
| connection.close | ✅ | |
| connection.close-ok | ✅ | |

## channel

| Method | Status | Notes |
|--------|:------:|------|
| channel.open | ✅ | |
| channel.open-ok | ✅ | |
| channel.flow | 📕 | Flow control not yet implemented |
| channel.flow-ok | 📕 | |
| channel.close | ✅ | |
| channel.close-ok | ✅ | |

## exchange

| Method | Status | Notes |
|--------|:------:|------|
| exchange.declare | ⏳ | Supports `direct`/`fanout`; missing `topic` |
| exchange.declare-ok | ✅ | |
| exchange.delete | ✅ | |
| exchange.delete-ok | ✅ | |

## queue

| Method | Status | Notes |
|--------|:------:|------|
| queue.declare | ✅ | |
| queue.declare-ok | ✅ | |
| queue.bind | ✅ | |
| queue.bind-ok | ✅ | |
| queue.unbind | 📕 | |
| queue.unbind-ok | 📕 | |
| queue.purge | 📕 | |
| queue.purge-ok | 📕 | |
| queue.delete | ⏳ | missing `if-unused`/`if-empty`|
| queue.delete-ok | ✅ | |

## basic

| Method | Status | Notes |
|--------|:------:|------|
| basic.qos | 📕 | Not yet implemented |
| basic.qos-ok | 📕 | |
| basic.consume | ⏳ | noLocal not supported (same as RabbitMQ) |
| basic.consume-ok | ✅ | |
| basic.cancel | ✅ | |
| basic.cancel-ok | ✅ | |
| basic.publish | ✅ | |
| basic.return | 📕 | Mandatory/immediate flags not fully handled |
| basic.deliver | ✅ | |
| basic.get | ✅ | Pull-based message retrieval |
| basic.get-ok | ✅ | |
| basic.get-empty | ✅ | |
| basic.ack | ✅ | Supports multiple flag |
| basic.reject | ⏳ | Requeue works; dead-lettering TODO |
| basic.recover-async | ✅ | |
| basic.recover | ✅ | |
| basic.recover-ok | ✅ | |
| basic.nack | 📕 | Not yet implemented |

## tx (Transactions)

| Method | Status | Notes |
|--------|:------:|------|
| tx.select | 📕 | Transaction support not yet implemented |
| tx.select-ok | 📕 | |
| tx.commit | 📕 | |
| tx.commit-ok | 📕 | |
| tx.rollback | 📕 | |
| tx.rollback-ok | 📕 | |

---

**Notes:**

- Keep this table in sync with the implementation in `internal/core/amqp/*` and `internal/core/broker/*`.
- When adding or changing behavior, update the status and add notes on limitations or differences from RabbitMQ behavior.
- "Partial" (⏳) means one or more optional behaviors/properties are not yet implemented.
