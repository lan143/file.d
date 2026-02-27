package decoder

import (
	"fmt"
	insaneJSON "github.com/ozontech/insane-json"
	"regexp"
	"time"
)

type (
	esp32Decoder struct{}

	ESP32Row struct {
		Time             []byte
		TimeAfterRestart []byte
		Level            []byte
		File             []byte
		Function         []byte
		Tag              []byte
		Message          []byte
	}

	esp32ParseState int
)

const (
	esp32ParseTimeState esp32ParseState = iota
)

var esp32logRegexp = regexp.MustCompile(
	`^\[\s*(\d+)\s*\]\[([A-Z])\]\[([^\]]+)\]\s+([a-zA-Z0-9_]+\(\)):\s+\[([^\]]+)\]\s+(.*)$`,
)

func NewESP32Decoder(params Params) (Decoder, error) {
	return &esp32Decoder{}, nil
}

func (d *esp32Decoder) Type() Type {
	return ESP32
}

func (d *esp32Decoder) DecodeToJson(root *insaneJSON.Root, data []byte) error {
	rowRaw, err := d.Decode(data)
	if err != nil {
		return err
	}
	row := rowRaw.(ESP32Row)

	root.AddFieldNoAlloc(root, "timestamp").MutateToBytesCopy(root, row.Time)
	root.AddFieldNoAlloc(root, "timeAfterRestart").MutateToBytesCopy(root, row.TimeAfterRestart)
	root.AddFieldNoAlloc(root, "level").MutateToBytesCopy(root, row.Level)
	root.AddFieldNoAlloc(root, "caller").MutateToBytesCopy(root, row.File)
	root.AddFieldNoAlloc(root, "function").MutateToBytesCopy(root, row.Function)
	root.AddFieldNoAlloc(root, "logger").MutateToBytesCopy(root, row.Tag)
	root.AddFieldNoAlloc(root, "message").MutateToBytesCopy(root, row.Message)

	return nil
}

func (d *esp32Decoder) Decode(data []byte, args ...any) (any, error) {
	now := time.Now()
	matches := esp32logRegexp.FindStringSubmatch(string(data))
	if matches == nil {
		return nil, fmt.Errorf("log format mismatch")
	}

	var level string
	switch matches[2] {
	case "E":
		level = "ERROR"
	case "W":
		level = "WARN"
	case "I":
		level = "INFO"
	case "D":
		level = "DEBUG"
	case "V":
		level = "VERBOSE"
	default:
		level = "UNKNOWN"
	}

	return ESP32Row{
		Time:             []byte(now.Format(time.RFC3339)),
		TimeAfterRestart: []byte(matches[1]),
		Level:            []byte(level),
		File:             []byte(matches[3]),
		Function:         []byte(matches[4]),
		Tag:              []byte(matches[5]),
		Message:          []byte(matches[6]),
	}, nil
}
