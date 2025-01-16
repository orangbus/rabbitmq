package rabbitmq

import (
	"github.com/goravel/framework/contracts/foundation"
)

const Binding = "rabbitmq"

var App foundation.Application

type ServiceProvider struct {
}

func (receiver *ServiceProvider) Register(app foundation.Application) {
	App = app
	app.Singleton(Binding, func(app foundation.Application) (any, error) {
		return NewRabbitmq()
	})
}

func (receiver *ServiceProvider) Boot(app foundation.Application) {
	app.Publishes("github.com/orangbus/rabbitmq", map[string]string{
		"config/rabbitmq.go": app.ConfigPath("rabbitmq.go"),
	})
}
