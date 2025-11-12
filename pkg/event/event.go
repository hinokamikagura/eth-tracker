// Package event provides event data structures and serialization for blockchain events.
package event

import "encoding/json"

// Event represents a blockchain event that has been processed and filtered.
type Event struct {
	Block           uint64         `json:"block"`           // Block number
	ContractAddress string         `json:"contractAddress"` // Contract address involved
	Success         bool           `json:"success"`         // Whether the transaction succeeded
	Timestamp       uint64         `json:"timestamp"`       // Block timestamp
	TxHash          string         `json:"transactionHash"` // Transaction hash
	TxType          string         `json:"transactionType"` // Transaction type identifier
	Payload         map[string]any `json:"payload"`         // Event-specific payload data
	Index           uint           `json:"-"`               // Internal index (not serialized)
}

// Serialize converts the event to JSON bytes.
func (e Event) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// Deserialize parses JSON bytes into an Event.
func Deserialize(jsonData []byte) (Event, error) {
	var event Event
	if err := json.Unmarshal(jsonData, &event); err != nil {
		return event, err
	}
	return event, nil
}
