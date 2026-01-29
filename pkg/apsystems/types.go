package apsystems

import (
	"encoding/json"
	"strconv"
	"time"
)

// StringInt is a custom type that handles JSON unmarshaling of integer values
// that may be represented as either strings ("123") or numbers (123) in API responses.
type StringInt int

// UnmarshalJSON implements json.Unmarshaler for StringInt.
// It attempts to unmarshal the value first as a string, then as an integer.
// This allows the type to handle both "123" and 123 from the same JSON field.
func (si *StringInt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		val, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*si = StringInt(val)
		return nil
	}

	// Fall back to unmarshaling as integer
	var i int
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	*si = StringInt(i)
	return nil
}

type DeviceInfo struct {
	Data struct {
		DeviceID string    `json:"deviceId"`
		DeviceSN string    `json:"devSn"`
		SSIDName string    `json:"ssid"`
		IPAddr   string    `json:"ipAddr"`
		MinPower StringInt `json:"minPower"`
		MaxPower StringInt `json:"maxPower"`
		Firmware string    `json:"devVer"`
		Model    string    `json:"model"`
		CurPower StringInt `json:"curPower"`
	} `json:"data"`
}

type AlarmInfo struct {
	Data struct {
		Og    StringInt `json:"og"`    // Grid fault
		Isce1 StringInt `json:"isce1"` // PV1 short circuit
		Isce2 StringInt `json:"isce2"` // PV2 short circuit
		Oe    StringInt `json:"oe"`    // Output error
	} `json:"data"`
}

type OutputData struct {
	Data struct {
		P1  int     `json:"p1"`  // Power input 1 in Watts
		E1  float64 `json:"e1"`  // Energy input 1 today in kWh
		Te1 float64 `json:"te1"` // Total lifetime energy input 1 in kWh
		P2  int     `json:"p2"`  // Power input 2 in Watts
		E2  float64 `json:"e2"`  // Energy input 2 today in kWh
		Te2 float64 `json:"te2"` // Total lifetime energy input 2 in kWh
	} `json:"data"`
}

type PowerStatus struct {
	Data struct {
		Status StringInt `json:"status"` // 0 = normal, 1 = off, 2 = sleep
	} `json:"data"`
}

type PowerLimit struct {
	Data struct {
		MaxPower StringInt `json:"maxPower"`
	} `json:"data"`
}

type Statistics struct {
	// TODO: add individual inputs
	TotalPower          int
	TotalEnergyToday    float64
	TotalEnergyLifetime float64
	LastUpdate          time.Time
}
