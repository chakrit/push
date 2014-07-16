# PUSH

Features:

* Push to APN and GCM through a single interface.
* Streaming architecture.

Stream rows of device tokens from your database and send one by one until you are done. No
need to buffer the whole list in advance or build your own bufferring solution. GCM
payloads are collected into batches automatically.

# INSTALL

```
go get github.com/chakrit/push
```

Internally, `push` use the following noteworthy packages:

* APN - [github.com/anachronistic/apns](github.com/anachronistic/apns)
* GCM - [github.com/alexjlockwood/gcm](github.com/alexjlockwood/gcm)

More providers (or alternative implementation) can be added by implenmenting the `Service`
interface (docs coming soon.)

# CONFIGURE

Creates a client:

```go
client = push.NewClient()
```

Adds APN push target:

```go
client.Add(&push.APN{
  Gateway:  "gateway.sandbox.push.apple.com:2195",
  KeyFile:  "your_apn_key_file.pem",
  CertFile: "your_apn_cert_file.pem",
})
```

Adds GCM push target:

```go
client.Add(&push.GCM{
  ApiKey:    "YOUR_GCM_API_KEY_HERE",
  BatchSize: 1000, // or your preferred batch processing size.
})
```

Refer to GCM docs for the maximum number of tokens that can be sent at once.

Starts all the services and feedback listeners.

```go
e := client.Start()
if e != nil {
  // log
}
```

**NOTE:** You need to listen on the feedback channel, otherwise it will block the whole
process. Refer to the feedback section below for more info / code samples.

# SEND / STREAM

Starts a new session:

```go
session := client.Send(&push.Payload{
  Title:       "Hello Mr. Nice Customer!"
  Description: "A new product awaits you."
  Data:        map[string]interface{}{"referenceId": 123}
})
defer session.Close()
```

Sends `push.Device{}` as you gets them from your database:

```
defer rows.Close()
for rows.Next() {

  // obtain one device (example is using sqlx).
  var device *DeviceInfo
  e := rows.StructScan(device)
  noError(e)

  // send to a single device
  switch device.Class {
  case "iphone":
    session.Devices <- &push.Device{device.Token, push.DeviceTypeIOS}
  case "android":
    session.Devices <- &push.Device{device.Token, push.DeviceTypeAndroid}
  }

}
```

# FEEDBACK

You need to also listen for `Feedback`s for all push notifications. If you do not think
that you need to implement this right away, you can simply spin an empty loop to empty the
`Feedback` queue. This is required for the engine to keep going.

```go
go func() {
  for result := range client.Feedbacks {
    if result.NewToken != "" {
      // GCM may return a new canonical device token.
      // You need to update your result.Token in yoru database to result.NewToken here.
      go MergeDeviceToken(result.Token, result.NewToken)
    }

    switch {
    case nil: // success

    case result.Error == push.ErrInvalidDevice:
      // Bad device token as returned by both service providers.
      // You should remove the device association from your database here.
      go DeleteDeviceToken(result.Token)

    default:
      // Usually transfer general IO error.
    }
  }
}
```

There maybe any number of feedbacks received depending on the nature of each push
providers. For example, GCM returns feedback for all devices sent to their endpoint but
APN will only return ones where the delivery failed.

The presence (or absence) of `Feedback` does not directly translate to push notification
success or failure as both service providers provides this on a "best effort" basis.
Refers to each provider's documentation for more information.

# LICENSE

BSD-2 (see LICENSE.md file)

# SUPPORT / CONTRIBUTE

PRs infinitely welcome. Feel free to open a
[new GitHub issue](https://github.com/chakrit/push/issues/new) if you find any problem,
would like help, or just want to ask questions.

