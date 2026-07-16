# Google Nest Device Access / SDM Endpoints

This document lists the [Smart Device Management (SDM) API](https://developers.google.com/nest/device-access) surface we need to integrate for the `ghome` CLI + MCP server, patterned after [go-garmin](https://github.com/shotah/go-garmin)’s `ENDPOINTS.md`.

Official docs:
- [Device Access home](https://developers.google.com/nest/device-access)
- [Use the API](https://developers.google.com/nest/device-access/api)
- [REST reference](https://developers.google.com/nest/device-access/reference/rest)
- [Traits](https://developers.google.com/nest/device-access/traits)
- [Authorization](https://developers.google.com/nest/device-access/api/authorization)
- [Events (Pub/Sub)](https://developers.google.com/nest/device-access/api/events)

## Implementation Status

- [x] Implemented
- [ ] Not implemented

Unlike Garmin Connect (hundreds of REST paths), SDM is a **small REST surface** plus a large set of **trait commands** all posted through one method: `devices.executeCommand`. Reads return trait fields on `GET` device/structure/room resources.

---

## Base URLs & auth

| What | Value |
|------|-------|
| SDM API | `https://smartdevicemanagement.googleapis.com/v1` |
| Partner OAuth (PCM) | `https://nestservices.google.com/partnerconnections/{project-id}/auth` |
| Token exchange / refresh | `https://oauth2.googleapis.com/token` (standard Google OAuth) |
| Token revoke | `https://oauth2.googleapis.com/revoke` |
| OAuth scope (SDM) | `https://www.googleapis.com/auth/sdm.service` |
| OAuth scope (Pub/Sub) | `https://www.googleapis.com/auth/pubsub` |
| Redirect (this project) | `http://127.0.0.1:8787/oauth/callback` |

All SDM calls use `Authorization: Bearer {access_token}`.

Path prefix for resources:

```text
enterprises/{project-id}/...
```

`{project-id}` is the **Device Access** project UUID from [console.nest.google.com/device-access](https://console.nest.google.com/device-access), not the Google Cloud project number.

---

## SDK vs thin HTTP client

Google publishes an official Go client:

| Option | Package | Notes |
|--------|---------|-------|
| **Official Go API client** | [`google.golang.org/api/smartdevicemanagement/v1`](https://pkg.go.dev/google.golang.org/api/smartdevicemanagement/v1) | Generated stubs for the 7 REST methods. Docs mark it **complete / maintenance mode** (critical fixes only). Works with `option.WithTokenSource` after our OAuth flow. |
| **Thin HTTP (current)** | `home/` package | Matches go-garmin: one `doJSON`, session on disk, auto refresh. Enough because the REST surface is tiny. |
| **Pub/Sub client** (events only) | [`cloud.google.com/go/pubsub`](https://pkg.go.dev/cloud.google.com/go/pubsub) | Separate from SDM; needs a GCP service account with Pub/Sub Subscriber, not the user SDM token. |

**Recommendation:** keep the thin HTTP client for SDM (already working). Optionally pull `google.golang.org/api/smartdevicemanagement/v1` later if we want typed request/response structs — it does **not** expand trait coverage; traits/commands are still stringly-typed JSON. Add `cloud.google.com/go/pubsub` only when we implement event streaming.

---

## OAuth / session

Partner Connections Manager (PCM) replaces the normal Google authorize URL for Nest consent.

| Status | Method | Endpoint | Description |
|--------|--------|----------|-------------|
| [x] | GET | `https://nestservices.google.com/partnerconnections/{project-id}/auth?...` | PCM authorize (offline + consent) |
| [x] | POST | `https://oauth2.googleapis.com/token` | Exchange code / refresh access token |
| [ ] | POST | `https://oauth2.googleapis.com/revoke` | Revoke access token (logout currently drops local session only) |

Query params for PCM: `client_id`, `redirect_uri`, `response_type=code`, `scope=https://www.googleapis.com/auth/sdm.service`, `access_type=offline`, `prompt=consent`, optional `state`.

---

## REST: devices

List/get devices and run trait commands. Core of CLI + MCP.

| Status | Method | Endpoint | Description |
|--------|--------|----------|-------------|
| [x] | GET | `/enterprises/{project-id}/devices` | List authorized devices |
| [x] | GET | `/enterprises/{project-id}/devices/{device-id}` | Get device + all traits |
| [x] | POST | `/enterprises/{project-id}/devices/{device-id}:executeCommand` | Execute a trait command |

`executeCommand` body:

```json
{
  "command": "sdm.devices.commands.ThermostatMode.SetMode",
  "params": { "mode": "HEAT" }
}
```

Response is usually `{}` or `{ "results": { ... } }` (streams / images return results).

---

## REST: structures & rooms

Homes and rooms. Structure/room GETs expose Info / RoomInfo traits.

| Status | Method | Endpoint | Description |
|--------|--------|----------|-------------|
| [x] | GET | `/enterprises/{project-id}/structures` | List structures (homes) |
| [x] | GET | `/enterprises/{project-id}/structures/{structure-id}` | Get one structure |
| [x] | GET | `/enterprises/{project-id}/structures/{structure-id}/rooms` | List rooms in a structure |
| [x] | GET | `/enterprises/{project-id}/structures/{structure-id}/rooms/{room-id}` | Get one room |

---

## Trait commands (`executeCommand`)

Every write goes through the same REST method. Status below is whether we wrap the command in `home/` + CLI/MCP.

### Thermostat

Primary MCP use case. Device guide: [Thermostat](https://developers.google.com/nest/device-access/api/thermostat).

| Status | Command | Params | Description |
|--------|---------|--------|-------------|
| [x] | `sdm.devices.commands.ThermostatMode.SetMode` | `mode`: `HEAT` \| `COOL` \| `HEATCOOL` \| `OFF` | Standard thermostat mode |
| [x] | `sdm.devices.commands.ThermostatEco.SetMode` | `mode`: `MANUAL_ECO` \| `OFF` | Eco mode |
| [x] | `sdm.devices.commands.ThermostatTemperatureSetpoint.SetHeat` | `heatCelsius` | Heat setpoint (°C) |
| [x] | `sdm.devices.commands.ThermostatTemperatureSetpoint.SetCool` | `coolCelsius` | Cool setpoint (°C) |
| [x] | `sdm.devices.commands.ThermostatTemperatureSetpoint.SetRange` | `heatCelsius`, `coolCelsius` | HEATCOOL range |
| [x] | `sdm.devices.commands.Fan.SetTimer` | `timerMode`: `ON` \| `OFF`, optional `duration` (e.g. `"3600s"`) | Fan timer |

Read-only thermostat traits (via `GET` device): `Connectivity`, `Humidity`, `Info`, `Settings`, `Temperature`, `ThermostatHvac`, plus eco/mode/setpoint state fields.

### Camera live stream

WebRTC and/or RTSP depending on device. Guide: [Camera](https://developers.google.com/nest/device-access/api/camera), trait: [CameraLiveStream](https://developers.google.com/nest/device-access/traits/device/camera-live-stream).

| Status | Command | Params | Description |
|--------|---------|--------|-------------|
| [ ] | `sdm.devices.commands.CameraLiveStream.GenerateWebRtcStream` | `offerSdp` | Start WebRTC session |
| [ ] | `sdm.devices.commands.CameraLiveStream.ExtendWebRtcStream` | `mediaSessionId` | Extend WebRTC (~5 min sessions) |
| [ ] | `sdm.devices.commands.CameraLiveStream.StopWebRtcStream` | `mediaSessionId` | Stop WebRTC |
| [ ] | `sdm.devices.commands.CameraLiveStream.GenerateRtspStream` | `{}` | Start RTSP (`rtsps://…`) |
| [ ] | `sdm.devices.commands.CameraLiveStream.ExtendRtspStream` | `streamExtensionToken` | Extend RTSP |
| [ ] | `sdm.devices.commands.CameraLiveStream.StopRtspStream` | `streamExtensionToken` | Stop RTSP |

### Camera event image

Requires a Pub/Sub (or recent) `eventId` from motion/person/sound/chime.

| Status | Command | Params | Description |
|--------|---------|--------|-------------|
| [ ] | `sdm.devices.commands.CameraEventImage.GenerateImage` | `eventId` | Returns short-lived `url` + `token` for snapshot download |

Download is a separate GET to the returned URL with `Authorization: Basic {token}` (not the SDM Bearer token). Optional `?width=` / `?height=`.

### Camera clip preview

No executeCommand — clip URL arrives on Pub/Sub `CameraClipPreview.ClipPreview` events; download with Bearer access token.

| Status | Method | Endpoint | Description |
|--------|--------|----------|-------------|
| [ ] | GET | `{previewUrl}` from event | Download 10-frame mp4 preview (`Authorization: Bearer`) |

---

## Events (Google Cloud Pub/Sub)

Not SDM REST. Enable in Device Access Console, grant `sdm-publisher@googlegroups.com` publisher on your topic, create a subscription. Docs: [Subscribe to Events](https://developers.google.com/nest/device-access/subscribe-to-events).

| Status | API | Description |
|--------|-----|-------------|
| [ ] | Pub/Sub pull / push | Trait field changes + device events (motion, person, sound, chime, clip preview, connectivity, HVAC, …) |

Typical event types (resourceUpdate):

| Event | Trait |
|-------|-------|
| Motion | `sdm.devices.events.CameraMotion.Motion` |
| Person | `sdm.devices.events.CameraPerson.Person` |
| Sound | `sdm.devices.events.CameraSound.Sound` |
| Chime | `sdm.devices.events.DoorbellChime.Chime` |
| ClipPreview | `sdm.devices.events.CameraClipPreview.ClipPreview` |
| Trait field updates | Any trait field change (e.g. `ThermostatHvac`, `Connectivity`) |

Kickstart publishing after setup with `GET …/devices` once.

---

## Read-only traits (via GET device / structure / room)

No dedicated endpoints — present or absent on the resource JSON.

### Structure / room

| Trait | Resource | Notes |
|-------|----------|-------|
| `sdm.structures.traits.Info` | structure | Structure custom name |
| `sdm.structures.traits.RoomInfo` | room | Room display name |

### Device (common)

| Trait | Notes |
|-------|-------|
| `sdm.devices.traits.Info` | `customName` |
| `sdm.devices.traits.Connectivity` | `ONLINE` / `OFFLINE` |
| `sdm.devices.traits.Humidity` | Ambient humidity % |
| `sdm.devices.traits.Temperature` | Ambient °C |
| `sdm.devices.traits.Settings` | e.g. `temperatureScale` (read-only via API) |
| `sdm.devices.traits.Fan` | Fan timer state |

### Thermostat-specific

| Trait | Notes |
|-------|-------|
| `sdm.devices.traits.ThermostatMode` | Current + available modes |
| `sdm.devices.traits.ThermostatEco` | Eco mode + eco setpoints (setpoints not writable via API) |
| `sdm.devices.traits.ThermostatHvac` | `HEATING` / `COOLING` / `OFF` / … |
| `sdm.devices.traits.ThermostatTemperatureSetpoint` | Current heat/cool setpoints |

### Camera / doorbell / display

| Trait | Notes |
|-------|-------|
| `sdm.devices.traits.CameraImage` | Max resolution |
| `sdm.devices.traits.CameraLiveStream` | Codecs, protocols (`WEB_RTC`, `RTSP`) |
| `sdm.devices.traits.CameraEventImage` | Capability marker |
| `sdm.devices.traits.CameraClipPreview` | Capability marker |
| `sdm.devices.traits.CameraMotion` | Capability marker |
| `sdm.devices.traits.CameraPerson` | Capability marker |
| `sdm.devices.traits.CameraSound` | Capability marker |
| `sdm.devices.traits.DoorbellChime` | Capability marker |

Device `type` (`sdm.devices.types.*`) is informational only — **use traits**, not type, to decide capabilities.

Supported device guides: [Thermostat](https://developers.google.com/nest/device-access/api/thermostat), [Camera](https://developers.google.com/nest/device-access/api/camera) (+ battery / wired / floodlight variants), [Doorbell](https://developers.google.com/nest/device-access/api/doorbell) (+ battery / wired), [Display (Hub Max)](https://developers.google.com/nest/device-access/api/display).

---

## MCP tools ↔ API mapping

| MCP tool | Status | Backing API |
|----------|--------|-------------|
| `list_devices` | [x] | `GET …/devices` (+ compact trait summary) |
| `get_device` | [x] | `GET …/devices/{id}` |
| `list_structures` | [x] | `GET …/structures` |
| `get_structure` | [x] | `GET …/structures/{id}` |
| `list_rooms` | [x] | `GET …/structures/{id}/rooms` |
| `get_room` | [x] | `GET …/structures/{id}/rooms/{room}` |
| `set_thermostat_mode` | [x] | `ThermostatMode.SetMode` |
| `set_thermostat_temperature` | [x] | `SetHeat` / `SetCool` / `SetRange` |
| `set_thermostat_eco` | [x] | `ThermostatEco.SetMode` |
| `set_fan_timer` | [x] | `Fan.SetTimer` |
| (future) camera / events | [ ] | Skipped for now |

---

## Priority for Implementation

### Done / keep maintaining
1. OAuth PCM login + token refresh + session persist
2. List/get devices, list/get structures, list/get rooms
3. Thermostat mode / setpoints / eco / fan timer (CLI + MCP)

### High priority (remaining)
_(none for personal thermostat use — camera/image intentionally skipped)_

### Medium priority
1. Pub/Sub event subscriber (service account) for motion/chime/HVAC
2. OAuth revoke on `ghome logout`

### Low priority / out of SDM scope
1. Camera live stream + event image / clip preview (deferred)
2. Floodlight on/off — **not exposed** by SDM (camera-with-floodlight is stream/events only)
3. Non-Nest Matter plugs/lights — not this API
4. Switching to `google.golang.org/api/smartdevicemanagement/v1` for typed stubs only

---

## Notes

1. Temperatures in commands are **always Celsius**, even if the thermostat UI is Fahrenheit.
2. Setpoint commands fail in `OFF` or `MANUAL_ECO`; mode must match (`SetHeat` needs `HEAT`, etc.).
3. Camera event images expire ~30s after the event — download immediately.
4. Live stream sessions expire ~5 minutes — call Extend* or regenerate.
5. Refresh the user access token at least every **6 months** if the integration is event-heavy and rarely calls SDM (Google revokes idle refresh tokens).
6. Do **not** sync SDM devices into Home Graph / Assistant SYNC for commercial Assistant integrations.

## Reference

| Resource | URL |
|----------|-----|
| Device Access | https://developers.google.com/nest/device-access |
| SDM REST | https://developers.google.com/nest/device-access/reference/rest |
| Trait index | https://developers.google.com/nest/device-access/traits |
| Go client (optional) | https://pkg.go.dev/google.golang.org/api/smartdevicemanagement/v1 |
| Discovery doc | https://smartdevicemanagement.googleapis.com/$discovery/rest?version=v1 |
| go-garmin pattern | https://github.com/shotah/go-garmin |
