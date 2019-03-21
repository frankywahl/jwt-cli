# JWT-CLI

`jwt-cli` is a little command line utility for dealing with `JWT` web tokens.

This utility allows you to quickly encode/decode tokens from the command line or as part of a bash script.

## Usage examples

```bash
echo '{"Hello":"world"}' | jwt encode --secret secret # eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJIZWxsbyI6IndvcmxkIiwiZXhwIjoxNTUzNzI1NTIwfQ.ghG6wlutmLvifu29pGQRFJPe9-GkPvU3Rw3EDaeSzNU

```

```bash
echo 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJIZWxsbyI6IndvcmxkIiwiZXhwIjoxNTUzNzI1NTIwfQ.ghG6wlutmLvifu29pGQRFJPe9-GkPvU3Rw3EDaeSzNU' | jwt decode
```

## Installation

### Prerequisites

* golang (if installing from source)

### Procedure

```
make install
```

### Homebrew

```bash
brew tap frankywahl/brew git@github.com:frankywahl/homebrew-brew.git
brew install jwt-cli
```
