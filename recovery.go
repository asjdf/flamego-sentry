package recovery

import (
	"errors"
	"fmt"
	"github.com/flamego/flamego"
	"github.com/getsentry/raven-go"
	"log"
	"net/http"
	"runtime/debug"
)

func Recovery(client *raven.Client) flamego.Handler {
	return func(c flamego.Context, log *log.Logger) {
		defer func() {
			flags := map[string]string{
				"endpoint": c.Request().RequestURI,
			}
			if err := recover(); err != nil {
				debug.PrintStack()
				errStr := fmt.Sprint(err)
				packet := raven.NewPacket(errStr, raven.NewException(errors.New(errStr), raven.NewStacktrace(2, 3, nil)))
				client.Capture(packet, flags)
				c.ResponseWriter().WriteHeader(http.StatusInternalServerError)
			}
		}()

		c.Next()
	}
}
