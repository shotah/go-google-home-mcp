package home

// Device is an SDM device resource.
type Device struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Traits          map[string]any `json:"traits"`
	ParentRelations []ParentRelation `json:"parentRelations,omitempty"`
}

// ParentRelation links a device to a structure/room.
type ParentRelation struct {
	Parent      string `json:"parent"`
	DisplayName string `json:"displayName"`
}

// Structure is a home/building.
type Structure struct {
	Name   string         `json:"name"`
	Traits map[string]any `json:"traits"`
}

// Room is a room within a structure.
type Room struct {
	Name   string         `json:"name"`
	Traits map[string]any `json:"traits"`
}

// listDevicesResponse is the SDM ListDevices payload.
type listDevicesResponse struct {
	Devices []Device `json:"devices"`
}

type listStructuresResponse struct {
	Structures []Structure `json:"structures"`
}

type listRoomsResponse struct {
	Rooms []Room `json:"rooms"`
}

// ExecuteCommandRequest is the SDM executeCommand body.
type ExecuteCommandRequest struct {
	Command string         `json:"command"`
	Params  map[string]any `json:"params,omitempty"`
}

// DeviceSummary is a compact view for MCP/CLI listing.
type DeviceSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	CustomName  string `json:"customName,omitempty"`
	Room        string `json:"room,omitempty"`
	Online      string `json:"online,omitempty"`
	Temperature float64 `json:"temperatureC,omitempty"`
	Mode        string `json:"mode,omitempty"`
}
