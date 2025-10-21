---
title: AMQP 0.9.1 Support Status
---

## AMQP 0.9.1 Support Status

This page tracks OtterMQ's support for AMQP 0.9.1 classes and methods. It is intended to help users understand current capabilities and to guide contributors.

Status levels:

- Implemented: Feature is available and tested
- Partial: Some behavior is missing or differs from spec
- Planned: Not yet implemented but on the roadmap
- Not Supported: Out of scope or no plans yet

## Summary by Class

| Class | Status | Notes |
|------:|:------:|-------|
| connection | Partial | Handshake and basic lifecycle supported; see details below |
| channel | Partial | Basic open/close; see details |
| exchange | Partial | direct/fanout/topic basics implemented |
| queue | Partial | declare/bind basics implemented |
| basic | Partial | push-based consume/deliver implemented |

## connection

| Method | Status | Notes |
|--------|:------:|------|
| connection.start / start-ok | Implemented | |
| connection.tune / tune-ok | Implemented | |
| connection.open / open-ok | Implemented | |
| connection.close / close-ok | Implemented | |
| connection.blocked / unblocked | Planned | |

## channel

| Method | Status | Notes |
|--------|:------:|------|
| channel.open / open-ok | Implemented | |
| channel.flow / flow-ok | Planned | |
| channel.close / close-ok | Implemented | |

## exchange

| Method | Status | Notes |
|--------|:------:|------|
| exchange.declare / declare-ok | Partial | direct/fanout/topic |
| exchange.delete / delete-ok | Planned | |

## queue

| Method | Status | Notes |
|--------|:------:|------|
| queue.declare / declare-ok | Implemented | |
| queue.bind / bind-ok | Implemented | |
| queue.unbind / unbind-ok | Planned | |
| queue.purge / purge-ok | Planned | |
| queue.delete / delete-ok | Planned | |

## basic

| Method | Status | Notes |
|--------|:------:|------|
| basic.publish | Implemented | |
| basic.consume / deliver | Implemented | Push-based delivery |
| basic.get / get-ok | Planned | |
| basic.ack / nack / reject | Planned | |
| basic.recover / recover-ok | Planned | |
| basic.qos / qos-ok | Planned | |

Notes:

- Keep this table in sync with the implementation in `internal/core/amqp/*` and `internal/core/broker/*`.
- When adding or changing behavior, update the status and add notes on limitations or differences from RabbitMQ behavior.
