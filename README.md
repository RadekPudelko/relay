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
DB=Name of the sqlite database - ex:relay.db3
```
