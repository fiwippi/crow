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

// Timestamp is a time.Time which unmarshalls from a UNIX timestamp
type Timestamp struct {
	time.Time
}

// UnmarshalJSON decodes an int64 timestamp into a time.Time object
func (p *Timestamp) UnmarshalJSON(bytes []byte) error {
	// Decode the bytes into an int64
	var raw int64
	err := json.Unmarshal(bytes, &raw)
	if err != nil {
		return fmt.Errorf("error decoding timestamp: %w", err)
	}

	// Parse the unix timestamp
	*&p.Time = time.Unix(raw, 0)
	return nil
}
