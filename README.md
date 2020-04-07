# JWT-CLI

`jwt-cli` is a little command line utility for dealing with `JWT` web tokens.

This utility allows you to quickly encode/decode tokens from the command line or as part of a bash script.

## Getting Started

### Using Homebrew

```bash
brew install frankywahl/tap/jwt
```
### Using a Binary

1. Go grab the latest binary from the [Releases](https://github.com/frankywahl/jwt-cli/releases) page for your platform / operating system.
1. Extract the archive.
1. Run `./jwt encode -d '{"hello":"world"}'`

### Using Docker

```bash
docker pull frankywahl/jwt
docker run frankywahl/jwt encode -d '{"hello":"world"}'
```

## Usage examples

```bash
echo '{"Hello":"world"}' | jwt encode --secret secret # eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJIZWxsbyI6IndvcmxkIiwiZXhwIjoxNTUzNzI1NTIwfQ.ghG6wlutmLvifu29pGQRFJPe9-GkPvU3Rw3EDaeSzNU

```

```bash
echo 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJIZWxsbyI6IndvcmxkIiwiZXhwIjoxNTUzNzI1NTIwfQ.ghG6wlutmLvifu29pGQRFJPe9-GkPvU3Rw3EDaeSzNU' | jwt decode
```

## Development

### Prerequisites

* golang (if installing from source)

### Procedure

```bash
make install
```
