package router

import (
	ctrl "github.com/MHSarmadi/Umbra/Server/controllers"
	"github.com/gogearbox/gearbox"
)

// SetupRoutes registers application routes on the provided gearbox App.
func SetupRoutes(app gearbox.Gearbox, c *ctrl.Controller) {
	// hello-world group: supports GET and POST on /hello-world/
	helloWorldRoutes := []*gearbox.Route{
		app.Get("/", c.HelloWorld),
		app.Post("/", c.HelloWorld),
	}

	app.Group("/hello-world", helloWorldRoutes)

	// /demo/captcha - GET
	app.Get("/demo/captcha", c.DemoCaptcha)

	// /session/init - POST
	app.Post("/session/init", c.SessionInit)
}
