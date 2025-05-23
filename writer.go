package csv

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type SafetyOpts struct {
	ForceDoubleQuotes bool
	EscapeCharEqual   bool
	EscapeCharPlus    bool
	EscapeCharMinus   bool
	EscapeCharAt      bool
	EscapeCharTab     bool
	EscapeCharCR      bool
}

var FullSafety = SafetyOpts{
	ForceDoubleQuotes: true,
	EscapeCharEqual:   true,
	EscapeCharPlus:    true,
	EscapeCharMinus:   true,
	EscapeCharAt:      true,
	EscapeCharTab:     true,
	EscapeCharCR:      true,
}

var EscapeAll = SafetyOpts{
	ForceDoubleQuotes: false,
	EscapeCharEqual:   true,
	EscapeCharPlus:    true,
	EscapeCharMinus:   true,
	EscapeCharAt:      true,
	EscapeCharTab:     true,
	EscapeCharCR:      true,
}

// A SafeWriter writes records using CSV encoding.
//
// As returned by [NewSafeWriter], a SafeWriter writes records terminated by a
// newline and uses ',' as the field delimiter. The exported fields can be
// changed to customize the details before
// the first call to [SafeWriter.Write] or [SafeWriter.WriteAll].
//
// [SafeWriter.Comma] is the field delimiter.
//
// If [SafeWriter.UseCRLF] is true,
// the SafeWriter ends each output line with \r\n instead of \n.
//
// The writes of individual records are buffered.
// After all data has been written, the client should call the
// [SafeWriter.Flush] method to guarantee all data has been forwarded to
// the underlying [io.Writer].  Any errors that occurred should
// be checked by calling the [SafeWriter.Error] method.
type SafeWriter struct {
	Comma   rune // Field delimiter (set to ',' by NewSafeWriter)
	UseCRLF bool // True to use \r\n as the line terminator
	w       *bufio.Writer
	opts    SafetyOpts
}

// NewSafeWriter returns a new SafeWriter that writes to w.
func NewSafeWriter(w io.Writer, opts SafetyOpts) *SafeWriter {
	return &SafeWriter{
		Comma: ',',
		w:     bufio.NewWriter(w),
		opts:  opts,
	}
}

// Write writes a single CSV record to w along with any necessary quoting.
// A record is a slice of strings with each string being one field.
// Writes are buffered, so [SafeWriter.Flush] must eventually be called to ensure
// that the record is written to the underlying [io.Writer].
func (w *SafeWriter) Write(record []string) error {
	if !validDelim(w.Comma) {
		return errInvalidDelim
	}

	for n, field := range record {
		if n > 0 {
			if _, err := w.w.WriteRune(w.Comma); err != nil {
				return err
			}
		}

		if len(field) > 0 {
			// ADDED BY @samber ON 2024-12-05
			switch {
			case w.opts.EscapeCharEqual && field[0] == '=':
				field = " " + field
			case w.opts.EscapeCharPlus && field[0] == '+':
				field = " " + field
			case w.opts.EscapeCharMinus && field[0] == '-':
				field = " " + field
			case w.opts.EscapeCharAt && field[0] == '@':
				field = " " + field
			case w.opts.EscapeCharTab && field[0] == '\t':
				field = " " + field
			case w.opts.EscapeCharCR && field[0] == '\n':
				field = " " + field
			}
		}

		// If we don't have to have a quoted field then just
		// write out the field and continue to the next field.
		if !w.fieldNeedsQuotes(field) {
			if _, err := w.w.WriteString(field); err != nil {
				return err
			}
			continue
		}

		if err := w.w.WriteByte('"'); err != nil {
			return err
		}
		for len(field) > 0 {
			// Search for special characters.
			i := strings.IndexAny(field, "\"\r\n")
			if i < 0 {
				i = len(field)
			}

			// Copy verbatim everything before the special character.
			if _, err := w.w.WriteString(field[:i]); err != nil {
				return err
			}
			field = field[i:]

			// Encode the special character.
			if len(field) > 0 {
				var err error
				switch field[0] {
				case '"':
					_, err = w.w.WriteString(`""`)
				case '\r':
					if !w.UseCRLF {
						err = w.w.WriteByte('\r')
					}
				case '\n':
					if w.UseCRLF {
						_, err = w.w.WriteString("\r\n")
					} else {
						err = w.w.WriteByte('\n')
					}
				}
				field = field[1:]
				if err != nil {
					return err
				}
			}
		}
		if err := w.w.WriteByte('"'); err != nil {
			return err
		}
	}
	var err error
	if w.UseCRLF {
		_, err = w.w.WriteString("\r\n")
	} else {
		err = w.w.WriteByte('\n')
	}
	return err
}

// Flush writes any buffered data to the underlying [io.Writer].
// To check if an error occurred during Flush, call [SafeWriter.Error].
func (w *SafeWriter) Flush() {
	w.w.Flush()
}

// Error reports any error that has occurred during
// a previous [SafeWriter.Write] or [SafeWriter.Flush].
func (w *SafeWriter) Error() error {
	_, err := w.w.Write(nil)
	return err
}

// WriteAll writes multiple CSV records to w using [SafeWriter.Write] and
// then calls [SafeWriter.Flush], returning any error from the Flush.
func (w *SafeWriter) WriteAll(records [][]string) error {
	for _, record := range records {
		err := w.Write(record)
		if err != nil {
			return err
		}
	}
	return w.w.Flush()
}

// fieldNeedsQuotes reports whether our field must be enclosed in quotes.
// Fields with a Comma, fields with a quote or newline, and
// fields which start with a space must be enclosed in quotes.
// We used to quote empty strings, but we do not anymore (as of Go 1.4).
// The two representations should be equivalent, but Postgres distinguishes
// quoted vs non-quoted empty string during database imports, and it has
// an option to force the quoted behavior for non-quoted CSV but it has
// no option to force the non-quoted behavior for quoted CSV, making
// CSV with quoted empty strings strictly less useful.
// Not quoting the empty string also makes this package match the behavior
// of Microsoft Excel and Google Drive.
// For Postgres, quote the data terminating string `\.`.
func (w *SafeWriter) fieldNeedsQuotes(field string) bool {
	if field == "" {
		return false
	}

	if field == `\.` {
		return true
	}

	// ADDED BY @samber ON 2024-12-05
	if w.opts.ForceDoubleQuotes {
		return true
	}

	if w.Comma < utf8.RuneSelf {
		for i := 0; i < len(field); i++ {
			c := field[i]
			if c == '\n' || c == '\r' || c == '"' || c == byte(w.Comma) {
				return true
			}
		}
	} else {
		if strings.ContainsRune(field, w.Comma) || strings.ContainsAny(field, "\"\r\n") {
			return true
		}
	}

	r1, _ := utf8.DecodeRuneInString(field)
	return unicode.IsSpace(r1)
}

func validDelim(r rune) bool {
	return r != 0 && r != '"' && r != '\r' && r != '\n' && utf8.ValidRune(r) && r != utf8.RuneError
}

var errInvalidDelim = errors.New("csv: invalid field or comment delimiter")
