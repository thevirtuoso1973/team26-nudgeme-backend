package main

// NOTE: These structs are mostly for the expected JSON format, not necessarily the
// database schema.
// 'omitempty' indicates not to use default values if omitted, e.g. do not use
// 0 for a missing int

type WellbeingRecord struct {
	PostCode       string `json:"postCode"`
	WellbeingScore int16  `json:"wellbeingScore"`
	WeeklySteps    int    `json:"weeklySteps,omitempty"`
	ErrorRate      int    `json:"errorRate,omitempty"`
	SupportCode    string `json:"supportCode"`
	DateSent       string `json:"date_sent,omitempty"`
}

type User struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"` // sent unhashed
}

type NewMessageJSON struct {
	Identifier_from string      `json:"identifier_from"`
	Password        string      `json:"password"` // verifies identifier_from
	Identifier_to   string      `json:"identifier_to"`
	Data            interface{} `json:"data"`
}
