<!--generated-from:bf1398c42b7dfe269385837d6c210cea156131bacff2e0012ec54535d76bc415 DO NOT REMOVE, DO UPDATE -->
moov-io/ach-test-harness
===

[![GoDoc](https://godoc.org/github.com/moov-io/ach-test-harness?status.svg)](https://godoc.org/github.com/moov-io/ach-test-harness)
[![Build Status](https://github.com/moov-io/ach-test-harness/workflows/Go/badge.svg)](https://github.com/moov-io/ach-test-harness/actions)
[![Coverage Status](https://codecov.io/gh/moov-io/ach-test-harness/branch/master/graph/badge.svg)](https://codecov.io/gh/moov-io/ach-test-harness)
[![Go Report Card](https://goreportcard.com/badge/github.com/moov-io/ach-test-harness)](https://goreportcard.com/report/github.com/moov-io/ach-test-harness)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/moov-io/ach-test-harness/master/LICENSE)

A configurable FTP/SFTP server and Go library to interactively test ACH scenarios to replicate real world originations, returns, changes, prenotes, and transfers.


Docs: [docs](https://moov-io.github.io/ach-test-harness/) | [open api specification](api/api.yml)

## Project Status

This project is currently under development and could introduce breaking changes to reach a stable status. We are looking for community feedback so please try out our code or give us feedback!

## Getting Started

Read through the [project docs](docs/README.md) over here to get an understanding of the purpose of this project and how to run it.

## Getting Help

 channel | info
 ------- | -------
 [Project Documentation](docs/README.md) | Our project documentation available online.
Twitter [@moov_io](https://twitter.com/moov_io)	| You can follow Moov.IO's Twitter feed to get updates on our project(s). You can also tweet us questions or just share blogs or stories.
[GitHub Issue](https://github.com/moov-io/ach-test-harness/issues) | If you are able to reproduce a problem please open a GitHub Issue under the specific project that caused the error.
[moov-io slack](https://slack.moov.io/) | Join our slack channel (`#ach-test-harness`) to have an interactive discussion about the development of the project.

## Supported and Tested Platforms

- 64-bit Linux (Ubuntu, Debian), macOS, and Windows

## Contributing

Yes please! Please review our [Contributing guide](CONTRIBUTING.md) and [Code of Conduct](https://github.com/moov-io/ach/blob/master/CODE_OF_CONDUCT.md) to get started! Checkout our [issues for first time contributors](https://github.com/moov-io/ach-test-harness/contribute) for something to help out with.

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) and uses Go 1.14 or higher. See [Golang's install instructions](https://golang.org/doc/install) for help setting up Go. You can download the source code and we offer [tagged and released versions](https://github.com/moov-io/ach-test-harness/releases/latest) as well. We highly recommend you use a tagged release for production.

### Test Coverage

Improving test coverage is a good candidate for new contributors while also allowing the project to move more quickly by reducing regressions issues that might not be caught before a release is pushed out to our users. One great way to improve coverage is by adding edge cases and different inputs to functions (or [contributing and running fuzzers](https://github.com/dvyukov/go-fuzz)).

Tests can run processes (like sqlite databases), but should only do so locally.

## License

Apache License 2.0 See [LICENSE](LICENSE) for details.
