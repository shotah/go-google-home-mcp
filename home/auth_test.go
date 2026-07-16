package home

import (
	"bytes"
	"testing"
	"time"
)

func TestSessionSaveLoad(t *testing.T) {
	s := &Session{
		AccessToken:  "access",
		RefreshToken: "refresh",
		Expiry:       time.Now().UTC().Truncate(time.Second),
		ClientID:     "cid",
		ClientSecret: "sec",
		ProjectID:    "proj",
	}
	var buf bytes.Buffer
	if err := s.save(&buf); err != nil {
		t.Fatal(err)
	}
	var loaded Session
	if err := loaded.load(&buf); err != nil {
		t.Fatal(err)
	}
	if loaded.AccessToken != "access" || loaded.ProjectID != "proj" {
		t.Fatalf("mismatch: %+v", loaded)
	}
}

func TestSummarizeDevice(t *testing.T) {
	d := Device{
		Name: "enterprises/p/devices/abc",
		Type: "sdm.devices.types.THERMOSTAT",
		Traits: map[string]any{
			"sdm.devices.traits.Info": map[string]any{
				"customName": "Hallway",
			},
			"sdm.devices.traits.Connectivity": map[string]any{
				"status": "ONLINE",
			},
			"sdm.devices.traits.Temperature": map[string]any{
				"ambientTemperatureCelsius": 21.5,
			},
			"sdm.devices.traits.ThermostatMode": map[string]any{
				"mode": "HEAT",
			},
		},
		ParentRelations: []ParentRelation{{DisplayName: "Living Room"}},
	}
	sum := SummarizeDevice(d)
	if sum.ID != "abc" || sum.CustomName != "Hallway" || sum.Mode != "HEAT" || sum.Room != "Living Room" {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}
