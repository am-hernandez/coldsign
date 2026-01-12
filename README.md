<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->
<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <h3 align="center">coldsign</h3>

  <p align="center">
    An experimental, air‑gapped Ethereum transaction signer
    <br />
    <a href="#about-the-project"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/am-hernandez/coldsign/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <a href="https://github.com/am-hernandez/coldsign/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>



<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#what-coldsign-does">What coldsign does</a></li>
        <li><a href="#what-coldsign-does-not-do">What coldsign does NOT do</a></li>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li>
      <a href="#usage">Usage</a>
      <ul>
        <li><a href="#offline-machine">Offline Machine</a></li>
        <li><a href="#online-machine">Online Machine</a></li>
        <li><a href="#verification-tooling">Verification Tooling</a></li>
      </ul>
    </li>
    <li><a href="#repository-structure">Repository Structure</a></li>
    <li><a href="#security-model">Security Model</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#status">Status</a></li>
  </ol>
</details>



## About The Project

**coldsign** is an experimental, air-gapped Ethereum transaction signer.

It parses explicit transaction intents, enforces strict policy and identity checks, builds and signs EIP-1559 ETH transfers **offline**, and exports the signed transaction via terminal output as text or QR for broadcasting on an online machine.

The design is intentionally minimal, auditable, and refusal-first.

---

## What coldsign does (v1)

- Parses explicit `ETH_SEND` transaction intents
- Enforces local, refusal-first policy (chain, fees, bounds)
- Derives an Ethereum account from a BIP-39 mnemonic (BIP-44 index)
- Verifies the derived address matches the intent (`fromAddress`)
- Builds an unsigned EIP-1559 ETH transfer
- Requires **explicit user confirmation** before signing
- Signs the transaction **offline**
- Outputs:
  - human-readable transaction review
  - signed transaction hash
  - raw signed transaction hex
  - optional terminal QR for air-gap transfer

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## What coldsign does NOT do

- No networking
- No RPC calls
- No broadcasting
- No ERC-20, NFT, or contract calls (ETH transfers only)
- No key storage or persistence
- No GUI
- No intent construction (signing only)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Getting Started

This section describes how to build coldsign on an **online build machine** and run it on an **offline signing machine**.

### Prerequisites

- Go 1.25.5 or later
- One online machine (build / broadcast)
- One offline machine (air-gapped signer)
- A BIP-39 mnemonic phrase

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Installation

### Build (online machine)

```sh
git clone https://github.com/am-hernandez/coldsign.git
cd coldsign
go build -o coldsign
```

Transfer the resulting `coldsign` binary to the offline machine using removable media.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Usage (offline signing machine)

coldsign uses a subcommand-based interface. Running `coldsign help` shows:

```
░█▀▀░█▀█░█░░░█▀▄░█▀▀░▀█▀░█▀▀░█▀█
░█░░░█░█░█░░░█░█░▀▀█░░█░░█░█░█░█
░▀▀▀░▀▀▀░▀▀▀░▀▀░░▀▀▀░▀▀▀░▀▀▀░▀░▀
air-gapped Ethereum transaction signer

Usage:
  coldsign sign [flags] <intent.json>
  coldsign addr --index N [--qr]
  coldsign help
  coldsign version

Commands:
  sign     Review and sign transaction intents
  addr     Derive and display Ethereum addresses
  help     Show this help message
  version  Show version information
```

### Commands

- `coldsign sign` - Review and sign transaction intents
- `coldsign addr` - Derive and display Ethereum addresses
- `coldsign help` - Show help message
- `coldsign version` - Show version information

### Review-only mode (default)

```sh
./coldsign sign sample_intent.json
```

Or use backward-compatible syntax:

```sh
./coldsign sample_intent.json
```

This prints a full transaction review and exits without signing.

### Sign with explicit authorization

```sh
./coldsign sign --sign sample_intent.json
```

You will be shown a detailed review and asked to confirm the destination address before signing.

### Read intent from stdin (QR / pipe workflows)

```sh
echo "coldintent:v1:..." | ./coldsign sign --intent-stdin --sign
```

This mode is designed for camera / QR pipelines and reads a **single-line** intent from stdin.

### Render QR for air-gap transfer

```sh
./coldsign sign --sign --qr sample_intent.json
```

This prints the signed raw transaction as a terminal QR (to stderr).

### Derive and display addresses

```sh
./coldsign addr --index 0
```

Derives an Ethereum address from a BIP-39 mnemonic at the specified BIP-44 index. Optionally output as QR:

```sh
./coldsign addr --index 0 --qr
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

## Security Model

- Keys never leave the offline machine
- Signing is review-only by default
- Intent explicitly binds identity and transaction
- Address mismatch causes refusal
- User must explicitly confirm destination before signing
- All signed bytes are inspectable before broadcast
- QR acts as a one-way data diode

<p align="right">(<a href="#readme-top">back to top</a>)</p>

---

<!-- STATUS -->
## Status

Experimental / educational.

Do not use with funds you cannot afford to lose.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/am-hernandez/coldsign.svg?style=for-the-badge
[contributors-url]: https://github.com/am-hernandez/coldsign/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/am-hernandez/coldsign.svg?style=for-the-badge
[forks-url]: https://github.com/am-hernandez/coldsign/network/members
[stars-shield]: https://img.shields.io/github/stars/am-hernandez/coldsign.svg?style=for-the-badge
[stars-url]: https://github.com/am-hernandez/coldsign/stargazers
[issues-shield]: https://img.shields.io/github/issues/am-hernandez/coldsign.svg?style=for-the-badge
[issues-url]: https://github.com/am-hernandez/coldsign/issues
[license-shield]: https://img.shields.io/github/license/am-hernandez/coldsign.svg?style=for-the-badge
[license-url]: https://github.com/am-hernandez/coldsign/blob/main/LICENSE
[Go]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://go.dev/
[Ethereum]: https://img.shields.io/badge/Ethereum-3C3C3D?style=for-the-badge&logo=Ethereum&logoColor=white
[Ethereum-url]: https://ethereum.org/
