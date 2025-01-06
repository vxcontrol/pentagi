package graph

import (
	"pentagi/pkg/controller"
	"pentagi/pkg/database"
	"pentagi/pkg/graph/subscriptions"
	"pentagi/pkg/providers"
	"pentagi/pkg/templates"

	"github.com/sirupsen/logrus"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB              database.Querier
	Logger          *logrus.Entry
	DefaultPrompter templates.Prompter
	ProvidersCtrl   providers.ProviderController
	Controller      controller.FlowController
	Subscriptions   subscriptions.SubscriptionsController
}
