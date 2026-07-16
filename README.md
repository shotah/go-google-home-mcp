# go-google-home-mcp

Go CLI + **MCP server** for [Google Nest / Google Home](https://developers.google.com/nest/device-access) via the **Smart Device Management (SDM)** API.

Scaffolded from the same patterns as [shotah/go-garmin](https://github.com/shotah/go-garmin): static binary, stdio MCP, OAuth session on disk with **auto token refresh + persist**.

```bash
ghome login          # browser OAuth → ~/.config/ghome/session.json
ghome devices        # list Nest devices
ghome mcp            # MCP server for ZeroClaw / Claude / Cursor
```

## MCP tools

| Tool | Description |
|---|---|
| `list_devices` | Compact device list (id, type, room, temp, mode) |
| `get_device` | Full SDM traits for one device |
| `list_structures` | Homes / structures |
| `set_thermostat_mode` | `HEAT` / `COOL` / `HEATCOOL` / `OFF` |
| `set_thermostat_temperature` | Heat and/or cool setpoints (°C) |
| `set_thermostat_eco` | `MANUAL_ECO` / `OFF` |

## Setup (one-time)

1. **Google Cloud** project → enable [Smart Device Management API](https://console.cloud.google.com/apis/library/smartdevicemanagement.googleapis.com).
2. **OAuth client** (Desktop or Web) with redirect URI:
   `http://127.0.0.1:8787/oauth/callback`
3. **Device Access Console** → [console.nest.google.com/device-access](https://console.nest.google.com/device-access)  
   Create a project (**$5** one-time), link the OAuth client, note the **Project ID**.
4. Login:

```bash
export GHOME_CLIENT_ID=...
export GHOME_CLIENT_SECRET=...
export GHOME_PROJECT_ID=...   # Device Access project id (UUID)

make build
./ghome login
# open the printed URL, approve Nest access
```

Session path:

```text
$XDG_CONFIG_HOME/ghome/session.json
# typically ~/.config/ghome/session.json
# Docker/ZeroClaw: HOME=/zeroclaw-data → /zeroclaw-data/.config/ghome/session.json
```

Access tokens refresh automatically; rotated tokens are written back to `session.json`.

## CLI

```bash
ghome devices
ghome structures
ghome thermostat mode --device DEVICE_ID --mode COOL
ghome thermostat temp --device DEVICE_ID --cool 22
ghome thermostat eco --device DEVICE_ID --mode MANUAL_ECO
ghome logout
```

## MCP (ZeroClaw / Tim)

```toml
[[mcp.servers]]
name = "google-home"
transport = "stdio"
command = "ghome"
args = ["mcp"]

[mcp_bundles.google-home]
servers = ["google-home"]
```

Mount `secrets/ghome` → `/zeroclaw-data/.config/ghome` and set `HOME=/zeroclaw-data`.

## Build / release

```bash
make build
make test
# tag v0.1.0 → GitHub Actions goreleaser publishes linux/darwin/windows binaries
```

## Limits / not yet

- Camera WebRTC/RTSP stream helpers (trait exists; not wired yet)
- Pub/Sub event streaming
- Non-thermostat write commands (plugs / lights via Matter bridges vary)

## License

Apache-2.0 (same family as go-garmin). Confirm before publishing if you need a different license.
