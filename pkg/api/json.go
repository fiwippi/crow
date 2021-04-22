package api

import (
	"encoding/json"
	"fmt"
	"time"
)

// Bool allows 0/1 to also become boolean.
type Bool bool

func (bit *Bool) UnmarshalJSON(b []byte) error {
	var n int
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}
	*bit = n == 1

	return nil
}

// Unix Timestamp
type Timestamp struct {
	time.Time
}

// UnmarshalJSON decodes an int64 timestamp into a time.Time object
func (p *Timestamp) UnmarshalJSON(bytes []byte) error {
	// 1. Decode the bytes into an int64
	var raw int64
	err := json.Unmarshal(bytes, &raw)

	if err != nil {
		fmt.Printf("error decoding timestamp: %s\n", err)
		return err
	}

	// 2 - Parse the unix timestamp
	*&p.Time = time.Unix(raw, 0)
	return nil
}
