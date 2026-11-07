package probe

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"
)

type fakeLicenseClient struct {
	response []byte
	err      error
}

func (f *fakeLicenseClient) Get(_ string, obj any) error {
	if f.err != nil {
		return f.err
	}
	return json.Unmarshal(f.response, obj)
}

func TestGetLicenseProbe_WithExpirationDate(t *testing.T) {
	payload := []byte(`{
		"kind": "tm:sys:license:licensestats",
		"entries": {
			"https://localhost/mgmt/tm/sys/license/0": {
				"nestedStats": {
					"entries": {
						"applianceSerialNumber": {"description": "f5-xxxxxxxxxxx"},
						"dailyRenewNotifPeriod": {"description": "5"},
						"licenseEndDate": {"description": "2026/12/03"},
						"licenseStartDate": {"description": "2026/10/17"},
						"licensedOnDate": {"description": "2026/10/18"},
						"licensedVersion": {"description": "17.1.3"},
						"platformId": {"description": "C115"},
						"registrationKey": {"description": "xxxxxxxxxxxxx"},
						"serviceCheckDate": {"description": "2024/09/05"}
					}
				}
			}
		}
	}`)

	c := &fakeLicenseClient{response: payload}
	metrics, ok := GetLicenseProbe(c, "target1")
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if len(metrics) == 0 {
		t.Fatalf("expected metrics to be returned")
	}

	var sawAvailable1, sawExpired bool
	for _, m := range metrics {
		var pb dto.Metric
		if err := m.Write(&pb); err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}
		desc := m.Desc().String()
		switch {
		case strings.Contains(desc, "bigip_license_available") && pb.GetGauge().GetValue() == 1:
			sawAvailable1 = true
		case strings.Contains(desc, "bigip_license_expired"):
			sawExpired = true
		}
	}

	if !sawAvailable1 {
		t.Errorf("expected bigip_license_available=1 to be present")
	}
	if !sawExpired {
		t.Errorf("expected bigip_license_expired metric to be present")
	}
}

func TestGetLicenseProbe_WithoutExpirationDate(t *testing.T) {
	payload := []byte(`{
		"kind": "tm:sys:license:licensestdemo",
		"entries": {
			"https://localhost/mgmt/tm/sys/license/0": {
				"nestedStats": {
					"entries": {
						"registrationKey": {"description": "xxxxxxxxxxxxx"}
					}
				}
			}
		}
	}`)

	c := &fakeLicenseClient{response: payload}
	metrics, ok := GetLicenseProbe(c, "target1")
	if !ok {
		t.Fatalf("expected ok=true (probe should not fail when expiration is unavailable), got false")
	}

	var sawAvailable0 bool
	for _, m := range metrics {
		var pb dto.Metric
		if err := m.Write(&pb); err != nil {
			t.Fatalf("failed to write metric: %v", err)
		}
		desc := m.Desc().String()
		if strings.Contains(desc, "bigip_license_available") && pb.GetGauge().GetValue() == 0 {
			sawAvailable0 = true
		}
		if strings.Contains(desc, "bigip_license_expiration") {
			t.Errorf("did not expect expiration metrics when licenseEndDate is absent")
		}
	}

	if !sawAvailable0 {
		t.Errorf("expected bigip_license_available=0 to be present")
	}
}

func TestGetLicenseProbe_HTTPError(t *testing.T) {
	c := &fakeLicenseClient{err: errors.New("connection refused")}
	metrics, ok := GetLicenseProbe(c, "target1")
	if !ok {
		t.Fatalf("expected ok=true (HTTP errors shouldn't fail the whole probe), got false")
	}
	if len(metrics) != 1 {
		t.Fatalf("expected exactly 1 metric (bigip_license_available=0), got %d", len(metrics))
	}
}

func TestParseLicenseDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantOK  bool
		wantUTC time.Time
	}{
		{
			name:    "valid date",
			input:   "2026/12/03",
			wantOK:  true,
			wantUTC: time.Date(2026, time.December, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "empty string",
			input:  "",
			wantOK: false,
		},
		{
			name:   "garbage value",
			input:  "not-a-date",
			wantOK: false,
		},
		{
			name:   "wrong format",
			input:  "03-12-2026",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, ok := parseLicenseDate(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("parseLicenseDate(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && ts != tt.wantUTC.Unix() {
				t.Fatalf("parseLicenseDate(%q) = %d, want %d", tt.input, ts, tt.wantUTC.Unix())
			}
		})
	}
}
