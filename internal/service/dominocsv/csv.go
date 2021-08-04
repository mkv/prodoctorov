package dominocsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
)

const (
	prefix    = `&\#`
	prefixLen = 2

	suffix    = `;`
	suffixLen = 1

	minUnquotedLen = 4
)

var re *regexp.Regexp

func init() {
	re = regexp.MustCompile(fmt.Sprintf(`%s\d+%s`, prefix, suffix))
}

// DecodeValue - unquotes symbols in a string value
func DecodeValue(text string) string {
	return re.ReplaceAllStringFunc(text, func(s string) string {
		n := len(s)
		if n < minUnquotedLen {
			return s
		}

		i, err := strconv.Atoi(s[prefixLen : n-suffixLen])
		if err != nil {
			return s
		}

		return string(rune(i))
	})
}

type Reader struct {
	reader *csv.Reader
}

func NewReader(in io.Reader) (*Reader, error) {
	r := &Reader{
		reader: csv.NewReader(in),
	}

	return r, nil
}

func (r *Reader) Read() ([]string, error) {
	rawRecord, err := r.reader.Read()
	if err != nil {
		return nil, err
	}

	record := make([]string, len(rawRecord))

	for i, cell := range rawRecord {
		record[i] = DecodeValue(cell)
	}

	return record, nil
}
