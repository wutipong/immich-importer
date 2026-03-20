package archive

import (
	"fmt"
	"log/slog"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

var charsetMap = map[string]string{
	"GB-18030": "GB18030",
}

func DetectCharSet(str string) (encoding encoding.Encoding, charset string, err error) {
	detector := chardet.NewTextDetector()

	chardetResult, err := detector.DetectBest([]byte(str))
	if err != nil {
		err = fmt.Errorf("failed to detect charset: %w", err)
		slog.Warn(
			"failed to detect encoding.",
			slog.String("text", str),
			slog.String("error", err.Error()),
		)

		return
	}
	charset, ok := charsetMap[chardetResult.Charset]
	if !ok {
		charset = chardetResult.Charset
	}

	encoding, err = ianaindex.IANA.Encoding(charset)
	return
}
