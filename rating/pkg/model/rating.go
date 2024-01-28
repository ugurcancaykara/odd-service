package model

// RecordID defines a record id. Together with RecordType
// identifies unique records across all types.
type RecordID string

// RecordType defines a record type. Together with RecordID
// identifies unique records across all types.
type RecordType string

// Existing record types
const (
	RecordTypeMovie       = RecordType("movie")
	RatingEventTypePut    = "put"
	RatingEventTypeDelete = "delete"
)

// UserID defines a user id.
type UserID string

// RatingValue defines a value of a rating record.
type RatingValue int

// Rating defines an individual rating created by a user
// for some record.
type Rating struct {
	RecordID   string      `json:"recordId"`
	RecordType string      `json:"recordType"`
	UserID     UserID      `json:"userId"`
	Value      RatingValue `json:"value"`
}

// RatingEventType defines the type of a rating event
type RatingEventType string

// An example of the provided rating data would be as follow in json:
// [{"userId":"105","recordId":"1","recordType":1,"value":5,"providerId":"test-provier","eventType":"put"},
// {"userId":"105"},"recordId":"2","recordType":1,"value":4,"providerId":"test-provider","eventType":"put"]
// RatingEvent defines an event containing rating information. For provided data from Kafka
type RatingEvent struct {
	UserID     UserID          `json:"userId"`
	RecordID   RecordID        `json:"recordId"`
	RecordType RecordType      `json:"recordType"`
	Value      RatingValue     `json:"value"`
	ProviderID string          `json:"providerId"`
	EventType  RatingEventType `json:"eventType"`
}
