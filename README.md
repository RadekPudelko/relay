# relay

An app to interface with Particle IO's devices, providing a means to trigger cloud functions, even if the devices are offline.

An api interface provides a means to create "relays", the cloud function with arguments and desired returns code for a specific device.

As long as the app is running, it will periodically check to see if a device is online and try to run the relay.

Currently configured to run localhost:8080

```
GET "/api/relays/{id}" - get information about a relay by its id
```

```
POST "/api/relays/" - create a relay providing:
{
    "device_id": string,
    "cloud_function": string
    "argument": optional string
    "desired_return_code": optional int
    "scheduled_time": optional datetime
}
Returns the id of a successfully created relay
```

```
DELETE "/api/relays/{id}" - cancel a relay by id
```

Requires a .env file in the format
```
PARTICLE_TOKEN=Particle IO token
```

It is possible to configure the app via a config.toml file. Lowering ping_retry_seconds and cf_retry_seconds will result in a higher chance of reaching a device when it comes online, however, I recommend sticking to the defaults.
```
# config.toml
[server]
host = "127.0.0.1"]
port = 8080

[database]
filename = "relay.db3"

[settings]
relay_limit = 100          # Number of relay requests loaded into memory at once
max_routines = 4           # Number of routines that can process relays at the same time
ping_retry_seconds = 180   # Seconds to wait before pinging a device again
cf_retry_seconds = 120     # Seconds to wait before retrying a cloud function if there was no answer
max_retries =  3           # Max number cloud function attempts given no answer in previous attempts
```

