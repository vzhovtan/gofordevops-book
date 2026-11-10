package marshalling

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const myDateFormat = "2006-01-02"

// MyData represents user information requiring custom JSON processing.
type MyData struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	BirthDate time.Time `json:"birth_date"`
	Token     string    `json:"token"` // Sensitive field to redact on Marshal
}

// MarshalJSON implements the json.Marshaler interface.
// It redacts the Token field and formats BirthDate as YYYY-MM-DD.
func (d *MyData) MarshalJSON() ([]byte, error) {
	if d == nil {
		return []byte("null"), nil
	}

	// Define an alias type to prevent infinite recursion during json.Marshal.
	type Alias MyData

	// Create an auxiliary anonymous struct to customize output.
	return json.Marshal(&struct {
		*Alias
		BirthDate string `json:"birth_date"`
		Token     string `json:"token"`
	}{
		Alias:     (*Alias)(d),
		BirthDate: d.BirthDate.Format(myDateFormat),
		Token:     "[REDACTED]", // Redact sensitive information
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It parses the custom date format and validates fields.
func (d *MyData) UnmarshalJSON(data []byte) error {
	type Alias MyData
	// The outer struct fields take precedence over embedded alias fields.
	aux := &struct {
		*Alias
		BirthDate string `json:"birth_date"`
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// 1. Custom Validation
	if d.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if !strings.Contains(d.Email, "@") {
		return fmt.Errorf("invalid email address")
	}
	// 2. Custom Date Formatting / Parsing
	if aux.BirthDate != "" {
		parsedDate, err := time.Parse(myDateFormat, aux.BirthDate)
		if err != nil {
			return fmt.Errorf("invalid birth_date format, expected %s: %w", myDateFormat, err)
		}
		d.BirthDate = parsedDate
	}
	return nil
}

func CustomMarshalling() {
	// 1. Demonstrate Custom Marshaling (Encoding)
	fmt.Println("--- 1. Custom Marshaling ---")
	dataToEncode := &MyData{
		Username:  "gopher_dev",
		Email:     "gopher@example.com",
		BirthDate: time.Date(1995, 11, 10, 0, 0, 0, 0, time.UTC),
		Token:     "secret_api_key_12345_super_secure",
	}

	// Using MarshalIndent to pretty-print the JSON output
	encodedJSON, err := json.MarshalIndent(dataToEncode, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling: %v\n", err)
		return
	}

	fmt.Println("Resulting JSON (Note the redacted token and YYYY-MM-DD date):")
	fmt.Println(string(encodedJSON))

	// 2. Demonstrate Custom Unmarshaling (Decoding & Date Parsing)
	fmt.Println("\n--- 2. Custom Unmarshaling ---")
	validJSON := []byte(`{
  "username": "custom_gopher",
  "email": "custom@example.com",
  "birth_date": "2001-05-23",
  "token": "incoming_token_abc"
 }`)
	var decodedData MyData
	if err := json.Unmarshal(validJSON, &decodedData); err != nil {
		fmt.Printf("Error unmarshaling valid JSON: %v\n", err)
	} else {
		fmt.Println("Successfully decoded struct:")
		fmt.Printf("  Username:   %s\n", decodedData.Username)
		fmt.Printf("  Email:      %s\n", decodedData.Email)
		fmt.Printf("  BirthDate:  %s\n", decodedData.BirthDate.Format(time.RFC3339))
		fmt.Printf("  Raw Struct: %+v\n", decodedData)
	}
	// 3. Demonstrate Validation Failures during Custom Unmarshaling
	fmt.Println("\n--- 3. Unmarshaling Validation Failures ---")

	// Case A: Missing Username
	badJSONNoUsername := []byte(`{"email": "<REDACTED_PII>", "birth_date": "2001-05-23"}`)
	var failedDataA MyData
	if err := json.Unmarshal(badJSONNoUsername, &failedDataA); err != nil {
		fmt.Printf("Caught Expected Error A: %v\n", err)
	}

	// Case B: Invalid Email
	badJSONInvalidEmail := []byte(`{"username": "user1", "email": "invalid-email-pattern", "birth_date": "2001-05-23"}`)
	var failedDataB MyData
	if err := json.Unmarshal(badJSONInvalidEmail, &failedDataB); err != nil {
		fmt.Printf("Caught Expected Error B: %v\n", err)
	}

	// Case C: Invalid Date Format
	badJSONInvalidDate := []byte(`{"username": "user1", "email": "<REDACTED_PII>", "birth_date": "05/23/2001"}`)
	var failedDataC MyData
	if err := json.Unmarshal(badJSONInvalidDate, &failedDataC); err != nil {
		fmt.Printf("Caught Expected Error C: %v\n", err)
	}
}
