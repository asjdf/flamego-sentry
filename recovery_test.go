package sentryflame

import (
	"fmt"
	"github.com/flamego/flamego"
	"github.com/getsentry/sentry-go"
	"net/http"
	"testing"
)

func TestRecovery(t *testing.T) {
	// To initialize Sentry's handler, you need to initialize Sentry itself beforehand
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: "your-public-dsn",
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	// Then create your app
	f := flamego.Classic()

	// Once it's done, you can attach the handler as one of your middleware
	f.Use(New(Options{}))

	// Set up routes
	f.Get("/", func() string {
		return "Hello world!"
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

	// And run it
	f.Run()
}
