package sentryflame

import (
	"context"
	"github.com/flamego/flamego"
	"github.com/flamego/flamego/inject"
	"github.com/getsentry/sentry-go"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

//const valuesKey = "sentry"

type handler struct {
	repanic         bool
	waitForDelivery bool
	timeout         time.Duration
}

type Options struct {
	// Repanic configures whether Sentry should repanic after recovery, in most cases it should be set to true,
	// as flamego.Classic includes it's own Recovery middleware what handles http responses.
	Repanic bool
	// WaitForDelivery configures whether you want to block the request before moving forward with the response.
	// Because Flamego's default Recovery handler doesn't restart the application,
	// it's safe to either skip this option or set it to false.
	WaitForDelivery bool
	// Timeout for the event delivery requests.
	Timeout time.Duration
}

// New returns a function that satisfies flamego.Handler interface
// It can be used with Use() methods.
func New(options Options) flamego.Handler {
	timeout := options.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	return (&handler{
		repanic:         options.Repanic,
		timeout:         timeout,
		waitForDelivery: options.WaitForDelivery,
	}).handle
}

func (h *handler) handle(ctx flamego.Context) {
	hub := GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	hub.Scope().SetRequest(ctx.Request().Request)
	ctx.Map(hub)
	defer h.recoverWithSentry(hub, ctx.Request().Request)
	ctx.Next()
}

func (h *handler) recoverWithSentry(hub *sentry.Hub, r *http.Request) {
	if err := recover(); err != nil {
		if !isBrokenPipeError(err) {
			eventID := hub.RecoverWithContext(
				context.WithValue(r.Context(), sentry.RequestContextKey, r),
				err,
			)
			if eventID != nil && h.waitForDelivery {
				hub.Flush(h.timeout)
			}
		}
		if h.repanic {
			panic(err)
		}
	}
}

// Check for a broken connection, as this is what Flamego does already.
func isBrokenPipeError(err interface{}) bool {
	if netErr, ok := err.(*net.OpError); ok {
		if sysErr, ok := netErr.Err.(*os.SyscallError); ok {
			if strings.Contains(strings.ToLower(sysErr.Error()), "broken pipe") ||
				strings.Contains(strings.ToLower(sysErr.Error()), "connection reset by peer") {
				return true
			}
		}
	}
	return false
}

// GetHubFromContext retrieves attached *sentry.Hub instance from flamego.Context.
func GetHubFromContext(ctx flamego.Context) *sentry.Hub {
	if ctx.Value(inject.InterfaceOf((*sentry.Hub)(nil))).IsValid() {
		return ctx.Value(inject.InterfaceOf((*sentry.Hub)(nil))).Interface().(*sentry.Hub)
	}
	return nil
}
