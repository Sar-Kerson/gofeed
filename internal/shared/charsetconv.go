package shared

import (
	"bufio"
	"io"
	"mime"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
)

func IsContentTypeUTF8(contentType string) bool {
	if _, params, err := mime.ParseMediaType(contentType); err == nil {
		if c, ok := params["charset"]; ok {
			cs := strings.ToLower(c)
			switch cs {
			case "unicode-1-1-utf-8",
				"utf-8",
				"utf8":
				return true
			}
		}
	}

	return false
}

func NewReaderLabel(label string, input io.Reader) (io.Reader, error) {
	conv, err := charset.NewReaderLabel(label, input)

	if err != nil {
		return nil, err
	}

	// Wrap the charset decoder reader with a XML sanitizer
	//clean := NewXMLSanitizerReader(conv)
	return conv, nil
}

type ValidUTF8Reader struct {
	buffer *bufio.Reader
}

func NewValidUTF8Reader(rd io.Reader) *ValidUTF8Reader {
	return &ValidUTF8Reader{bufio.NewReader(rd)}
}

func (rd ValidUTF8Reader) Read(b []byte) (n int, err error) {
	for {
		var r rune
		var size int
		r, size, err = rd.buffer.ReadRune()
		if err != nil {
			return
		}
		if r == unicode.ReplacementChar && size == 1 {
			// 遇到 invalid 的 rune，丢弃
			continue
		} else if !unicode.IsPrint(r) && size == 1 {
			// 遇到 unprintable 的 rune，丢弃
			continue
		} else if n+size < len(b) {
			// 边缘条件，将 rune 写入 buffer
			utf8.EncodeRune(b[n:], r)
			n += size
		} else {
			rd.buffer.UnreadRune()
			break
		}
	}
	return
}
