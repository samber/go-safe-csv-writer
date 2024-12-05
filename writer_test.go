package csv

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func buildTest(forceDoubleQuotes bool, escapeCharEqual bool, escapeCharPlus bool, escapeCharMinus bool, escapeCharAt bool, escapeCharTab bool, escapeCharCR bool) string {
	var buff strings.Builder

	writer := NewSafeWriter(
		&buff,
		SafetyOpts{
			ForceDoubleQuotes: forceDoubleQuotes,
			EscapeCharEqual:   escapeCharEqual,
			EscapeCharPlus:    escapeCharPlus,
			EscapeCharMinus:   escapeCharMinus,
			EscapeCharAt:      escapeCharAt,
			EscapeCharTab:     escapeCharTab,
			EscapeCharCR:      escapeCharCR,
		},
	)
	must(writer.Write([]string{"userId", "secret", "comment"}))
	must(writer.Write([]string{"-21+63", "=A1", "foo, bar"}))
	must(writer.Write([]string{"+42", "\tsecret", "\nplop"}))
	must(writer.Write([]string{"123", "blablabla", "@foobar"}))
	writer.Flush()
	must(writer.Error())

	return buff.String()
}

func TestNewSafeWriter(t *testing.T) {
	is := assert.New(t)

	// base case
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
+42,"	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, false, false, false, false, false),
	)

	// double quotes
	is.Equal(
		`"userId","secret","comment"
"-21+63","=A1","foo, bar"
"+42","	secret","
plop"
"123","blablabla","@foobar"
`,
		buildTest(true, false, false, false, false, false, false),
	)

	// escape "="
	is.Equal(
		`userId,secret,comment
-21+63," =A1","foo, bar"
+42,"	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, true, false, false, false, false, false),
	)

	// escape "+"
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
" +42","	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, true, false, false, false, false),
	)

	// escape "-"
	is.Equal(
		`userId,secret,comment
" -21+63",=A1,"foo, bar"
+42,"	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, false, true, false, false, false),
	)

	// escape "@"
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
+42,"	secret","
plop"
123,blablabla," @foobar"
`,
		buildTest(false, false, false, false, true, false, false),
	)

	// escape "\t"
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
+42," 	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, false, false, false, true, false),
	)

	// escape "\n"
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
+42,"	secret"," 
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, false, false, false, false, true),
	)

	// escape everything
	is.Equal(
		`userId,secret,comment
" -21+63"," =A1","foo, bar"
" +42"," 	secret"," 
plop"
123,blablabla," @foobar"
`,
		buildTest(false, true, true, true, true, true, true),
	)

	// escape everything + force double quotes
	is.Equal(
		`"userId","secret","comment"
" -21+63"," =A1","foo, bar"
" +42"," 	secret"," 
plop"
"123","blablabla"," @foobar"
`,
		buildTest(true, true, true, true, true, true, true),
	)
}

func TestNewSafeWriterNoOpts(t *testing.T) {
	is := assert.New(t)

	var buff strings.Builder

	w := NewSafeWriter(&buff, SafetyOpts{})
	is.Empty(w.opts)
}

func TestFullSafety(t *testing.T) {
	is := assert.New(t)

	is.NotEmpty(FullSafety)
	is.True(FullSafety.ForceDoubleQuotes)
	is.True(FullSafety.EscapeCharEqual)
	is.True(EscapeAll.EscapeCharPlus)
	is.True(EscapeAll.EscapeCharMinus)
	is.True(EscapeAll.EscapeCharAt)
	is.True(EscapeAll.EscapeCharTab)
	is.True(EscapeAll.EscapeCharCR)
}

func TestEscapeAll(t *testing.T) {
	is := assert.New(t)

	is.NotEmpty(EscapeAll)
	is.False(EscapeAll.ForceDoubleQuotes)
	is.True(EscapeAll.EscapeCharEqual)
	is.True(EscapeAll.EscapeCharPlus)
	is.True(EscapeAll.EscapeCharMinus)
	is.True(EscapeAll.EscapeCharAt)
	is.True(EscapeAll.EscapeCharTab)
	is.True(EscapeAll.EscapeCharCR)
}
