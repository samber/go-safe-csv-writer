package main

import (
	stdlib "encoding/csv"
	"fmt"
	"strings"

	csv "github.com/samber/go-safe-csv-writer"
)

type writer interface {
	Write([]string) error
	Flush()
	Error() error
}

func writeToWriter(w writer) {
	must(w.Write([]string{"userId", "secret", "comment"}))
	must(w.Write([]string{"-21+63", "=A1", "foo, bar"}))
	must(w.Write([]string{"+42", "\tsecret", "\nplop"}))
	must(w.Write([]string{"123", "blablabla", "@foobar"}))
	w.Flush()
	must(w.Error())
}

func unsafe() string {
	var buff strings.Builder

	writer := stdlib.NewWriter(&buff)
	writeToWriter(writer)

	return buff.String()
}

func safeWriterForceDoubleQuotes() string {
	var buff strings.Builder

	writer := csv.NewSafeWriter(&buff, csv.SafetyOpts{ForceDoubleQuotes: true})
	writeToWriter(writer)

	return buff.String()
}

func safeWriterEscapeEverything() string {
	var buff strings.Builder

	writer := csv.NewSafeWriter(&buff, csv.EscapeAll)
	writeToWriter(writer)

	return buff.String()
}

func main() {
	fmt.Printf(`
%s


%s


%s
`,
		unsafe(),
		safeWriterForceDoubleQuotes(),
		safeWriterEscapeEverything(),
	)
}
