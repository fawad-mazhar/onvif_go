# Configuration Package

Handles loading and parsing of the ONVIF server configuration.

## Entry Points

```go
cfg, err := config.Load("path/to/config.conf")   // .conf or .json (auto-detects by extension)
cfg, err := config.LoadConfig("path/to/config.conf") // flat key=value only
cfg, err := config.LoadConfigJSON("path/to/config.json") // JSON only
```

`Load` is the recommended entry point — it dispatches to the right parser based on the file extension.

## Flat key=value format (`.conf`)

No sections, no brackets. Keys are case-insensitive. Comments start with `#`.

### Device identity

| Key | Type | Description |
|---|---|---|
| `manufacturer` | string | Device manufacturer |
| `model` | string | Device model |
| `firmware_ver` | string | Firmware version |
| `serial_num` | string | Serial number |
| `hardware_id` | string | Hardware ID |
| `uuid` | string | Device UUID (used in WSD and scopes) |
| `ifs` | string | Network interface for IP/URL generation (e.g. `eth0`, `lo`) |
| `port` | int | HTTP listen port (default `8080`) |
| `wsd_port` | int | WS-Discovery UDP port (default `3702`) |
| `notification_port` | int | Notification server port (default `8081`) |
| `user` | string | WS-UsernameToken username (empty = auth disabled) |
| `password` | string | WS-UsernameToken password |

### Scopes

Repeated `scope=` keys append to the scope list:

```ini
scope=onvif://www.onvif.org/Profile/Streaming
scope=onvif://www.onvif.org/Profile/S
```

### Media profiles

`name=` starts a new profile; subsequent keys attach to the most recent one.

```ini
name=Profile_0
width=1920
height=1080
url=rtsp://%s/stream0      # %s replaced with device IP at runtime
snapurl=http://%s/snap.jpg
type=H264                  # H264, H265, JPEG, MPEG4
audio_encoder=AAC          # AAC, G711, G726, AudioNone
audio_decoder=G711
```

### PTZ

`ptz=1` enables PTZ and resets step defaults; step/command keys follow.

```ini
ptz=1
min_step_x=0
max_step_x=360
min_step_y=0
max_step_y=180
min_step_z=0
max_step_z=0
move_left=/usr/local/bin/ptz_move -m left -s %f
move_stop=/usr/local/bin/ptz_move -m stop -t %s
jump_to_abs=/usr/local/bin/ptz_move -j %f,%f,%f
jump_to_rel=/usr/local/bin/ptz_move -J %f,%f,%f
```

### Relay outputs

`idle_state=` starts a new relay output.

```ini
idle_state=open
close=/usr/local/bin/set_relay -n 0 -a close
open=/usr/local/bin/set_relay -n 0 -a open
```

### Events

`events=N` sets delivery mode; `topic=` starts a new event entry.

| `events` value | Mode |
|---|---|
| `1` | PullPoint |
| `2` | BaseSubscription |
| `3` | Both |

```ini
events=3
topic=tns1:VideoSource/MotionAlarm
source_name=Source
source_type=tt:ReferenceToken
source_value=VideoSourceToken
input_file=/tmp/onvif_notify_server/motion_alarm
```

## JSON format (`.json`)

The JSON format uses the same field names as the flat format. Pass any `.json` file to
`-config` and `Load` will route it to `LoadConfigJSON` automatically.
See `test/fixtures/config/server.json` for a complete example.
