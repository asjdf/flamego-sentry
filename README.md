<p align="center">
  <a href="https://sentry.io/?utm_source=github&utm_medium=logo" target="_blank">
    <picture>
      <source srcset="https://sentry-brand.storage.googleapis.com/sentry-logo-white.png" media="(prefers-color-scheme: dark)" />
      <source srcset="https://sentry-brand.storage.googleapis.com/sentry-logo-black.png" media="(prefers-color-scheme: light), (prefers-color-scheme: no-preference)" />
      <img src="https://sentry-brand.storage.googleapis.com/sentry-logo-black.png" alt="Sentry" width="280">
    </picture>
  </a>
</p>

# flamego-sentry
Package flamego-sentry is a middleware that capture and handle the error with Sentry  for Flamego

## Install

```bash
go get github.com/asjdf/flamego-sentry
```

```go
import (
    "fmt"
    "github.com/getsentry/sentry-go"
    sentryflame "github.com/asjdf/flamego-sentry"
    "github.com/flamego/flamego"
)

// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
if err := sentry.Init(sentry.ClientOptions{
    Dsn: "your-public-dsn",
}); err != nil {
    fmt.Printf("Sentry initialization failed: %v\n", err)
}

// Then create your app
f := flamego.Classic()

// Once it's done, you can attach the handler as one of your middleware
f.Use(sentryflame.New(sentryflame.Options{}))

// Set up routes
f.Get("/", func() string {
    return "Hello world!"
})

// And run it
f.Run()
```

## Configuration
`sentryflame` accepts a struct of `Options` that allows you to configure how the handler will behave.

Currently it respects 3 options:

```go
// Whether Sentry should repanic after recovery, in most cases it should be set to true,
// as flamego.Classic includes its own Recovery middleware that handles http responses.
Repanic         bool
// Whether you want to block the request before moving forward with the response.
// Because Flamego's default `Recovery` handler doesn't restart the application,
// it's safe to either skip this option or set it to `false`.
WaitForDelivery bool
// Timeout for the event delivery requests.
Timeout         time.Duration
```

## Usage

`sentryflame` invoke an instance of `*sentry.Hub` (https://godoc.org/github.com/getsentry/sentry-go#Hub) to the `flamego.Context`, which makes it available throughout the rest of the request's lifetime.
You can access it by using the `sentrygin.GetHubFromContext()` method on the context itself in any of your proceeding middleware and routes.
And it should be used instead of the global `sentry.CaptureMessage`, `sentry.CaptureException`, or any other calls, as it keeps the separation of data between the requests.

**Keep in mind that `*sentry.Hub` won't be available in middleware attached before to `sentryflame`!**

```go
f := flamego.Classic()

f.Use(sentryflame.New(sentrygin.Options{
    Repanic: true,
}))

f.Use(func(ctx flamego.Context) {
    if hub := sentryflame.GetHubFromContext(ctx); hub != nil {
        hub.Scope().SetTag("someRandomTag", "maybeYouNeedIt")
    }
    ctx.Next()
})

f.Get("/1", func(ctx flamego.Context) {
    if hub := GetHubFromContext(ctx); hub != nil {
        hub.WithScope(func(scope *sentry.Scope) {
            scope.SetExtra("unwantedQuery", "someQueryDataMaybe")
            hub.CaptureMessage("User provided unwanted query string, but we recovered just fine")
        })
    }
    ctx.ResponseWriter().WriteHeader(http.StatusOK)
})

f.Get("/2", func(ctx flamego.Context, hub *sentry.Hub) {
    hub.WithScope(func(scope *sentry.Scope) {
        scope.SetExtra("unwantedQuery", "someQueryDataMaybe")
        hub.CaptureMessage("User provided unwanted query string, but we recovered just fine")
    })
    ctx.ResponseWriter().WriteHeader(http.StatusOK)
})

f.Get("/foo", func() {
    // sentrygin handler will catch it just fine. Also, because we attached "someRandomTag"
    // in the middleware before, it will be sent through as well
    panic("y tho")
})

app.Run()
```

### Accessing Request in `BeforeSend` callback

```go
sentry.Init(sentry.ClientOptions{
    Dsn: "your-public-dsn",
    BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
        if hint.Context != nil {
            if req, ok := hint.Context.Value(sentry.RequestContextKey).(*http.Request); ok {
                // You have access to the original Request here
            }
        }

        return event
    },
})
```
