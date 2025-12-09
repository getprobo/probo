package slack_v1

import (
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/slack"
	"go.probo.inc/probo/pkg/trust"
)

func NewMux(
	logger *log.Logger,
	slackSvc *slack.Service,
	trustSvc *trust.Service,
) *chi.Mux {
	r := chi.NewMux()
	logger.Info("Registering Slack interactive endpoint")

	r.Post("/interactive", SlackHandler(
		slackSvc,
		slackSvc.GetSlackSigningSecret(),
		logger,
		trustSvc,
	))

	return r
}
