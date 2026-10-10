package config

import "github.com/paroki/domus/api/internal/core/controller"

func initCore(state State) {
	app := state.FiberApp
	hc := controller.HealthController{}
	hc.InitRoutes(app)
}
