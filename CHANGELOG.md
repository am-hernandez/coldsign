# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - 2026-01-11

### Added

- **sign** command: review and sign `ETH_SEND` transaction intents (default: review-only; no signing unless explicitly requested).
- **--sign** flag: required to authorize signing; without it, coldsign prints the review and exits.
- TTY-based hidden mnemonic and passphrase entry (no env vars); input is not echoed and memory is cleared after signing.
- Intent from stdin: `--intent-stdin` for JSON or `coldintent:v1:<base64url>` envelope (e.g. for QR/pipe workflows).
- Terminal QR output: `--qr` to print signed raw tx or address as a terminal QR for air-gap transfer.
- **addr** command: derive and display Ethereum addresses by BIP-44 index, with optional `--qr`.
- **help** and **version** commands; backward-compatible `coldsign <intent.json>` (no subcommand) for review.
- Strict policy defaults: chainId 1 (mainnet) and fee/value bounds enforced.
- Early intent validation: invalid `to` / `fromAddress` rejected before use.
- Interactive signing confirmation: re-type a displayed fragment (first and last 4 hex chars) of the destination address before signing (unless `--yes`).
- Validation of `maxFeePerGasWei` and address format to avoid panics and unintended numeric coercion.

[Unreleased]: https://github.com/am-hernandez/coldsign/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/am-hernandez/coldsign/releases/tag/v1.0.0
