package errors

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Encoding int

const (
	EncodingUTF8 Encoding = iota
	EncodingUTF16LE
	EncodingUTF16BE
	EncodingCP1252
	EncodingLatin1
)

func (e Encoding) String() string {
	switch e {
	case EncodingUTF16LE:
		return "UTF-16LE"
	case EncodingUTF16BE:
		return "UTF-16BE"
	case EncodingCP1252:
		return "Windows-1252"
	case EncodingLatin1:
		return "Latin-1"
	default:
		return "UTF-8"
	}
}

// DetectEncoding sniffs the first bytes to determine encoding.
func DetectEncoding(data []byte) Encoding {
	if len(data) >= 2 {
		if data[0] == 0xFF && data[1] == 0xFE {
			return EncodingUTF16LE
		}
		if data[0] == 0xFE && data[1] == 0xFF {
			return EncodingUTF16BE
		}
	}
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return EncodingUTF8 // UTF-8 BOM
	}
	if utf8.Valid(data) {
		return EncodingUTF8
	}
	// Check for CP1252-specific bytes (0x80-0x9F range) before falling back to Latin-1
	if hasCP1252Bytes(data) {
		return EncodingCP1252
	}
	return EncodingLatin1
}

// hasCP1252Bytes checks for bytes in 0x80-0x9F that are valid CP1252 but undefined in Latin-1.
func hasCP1252Bytes(data []byte) bool {
	for _, b := range data {
		if b >= 0x80 && b <= 0x9F {
			return true
		}
	}
	return false
}

// TranscodeToUTF8 converts data from detected encoding to UTF-8.
func TranscodeToUTF8(r io.Reader) (io.Reader, Encoding, error) {
	buf := bufio.NewReader(r)
	peek, err := buf.Peek(4)
	if err != nil && err != io.EOF {
		return buf, EncodingUTF8, nil
	}

	enc := DetectEncoding(peek)
	switch enc {
	case EncodingUTF16LE:
		return newUTF16Reader(buf, binary.LittleEndian), enc, nil
	case EncodingUTF16BE:
		return newUTF16Reader(buf, binary.BigEndian), enc, nil
	case EncodingCP1252:
		return transform.NewReader(buf, charmap.Windows1252.NewDecoder()), enc, nil
	case EncodingLatin1:
		return transform.NewReader(buf, charmap.ISO8859_1.NewDecoder()), enc, nil
	default:
		return buf, enc, nil
	}
}

func newUTF16Reader(r io.Reader, order binary.ByteOrder) io.Reader {
	return &utf16Reader{r: r, order: order, skipBOM: true}
}

type utf16Reader struct {
	r       io.Reader
	order   binary.ByteOrder
	skipBOM bool
	buf     bytes.Buffer
}

func (u *utf16Reader) Read(p []byte) (int, error) {
	if u.buf.Len() > 0 {
		return u.buf.Read(p)
	}

	raw := make([]byte, 4096)
	n, err := u.r.Read(raw)
	if n == 0 {
		return 0, err
	}
	raw = raw[:n]

	if u.skipBOM && len(raw) >= 2 {
		u.skipBOM = false
		if (u.order == binary.LittleEndian && raw[0] == 0xFF && raw[1] == 0xFE) ||
			(u.order == binary.BigEndian && raw[0] == 0xFE && raw[1] == 0xFF) {
			raw = raw[2:]
		}
	}

	// Ensure even number of bytes
	if len(raw)%2 != 0 {
		raw = raw[:len(raw)-1]
	}

	u16 := make([]uint16, len(raw)/2)
	for i := 0; i < len(raw); i += 2 {
		if u.order == binary.LittleEndian {
			u16[i/2] = uint16(raw[i]) | uint16(raw[i+1])<<8
		} else {
			u16[i/2] = uint16(raw[i])<<8 | uint16(raw[i+1])
		}
	}

	runes := utf16.Decode(u16)
	for _, r := range runes {
		u.buf.WriteRune(r)
	}

	return u.buf.Read(p)
}
