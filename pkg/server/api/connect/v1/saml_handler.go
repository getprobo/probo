package connect_v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
)

type SAMLHandler struct {
	iam          *iam.Service
	cookieConfig securecookie.Config
	baseURL      *baseurl.BaseURL
	logger       *log.Logger
}

func NewSAMLHandler(iam *iam.Service, cookieConfig securecookie.Config, baseURL *baseurl.BaseURL, logger *log.Logger) *SAMLHandler {
	return &SAMLHandler{iam: iam, cookieConfig: cookieConfig, baseURL: baseURL, logger: logger}
}

func (h *SAMLHandler) renderInternalServerError(w http.ResponseWriter, r *http.Request) {
	httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))
}

func (h *SAMLHandler) MetadataHandler(w http.ResponseWriter, r *http.Request) {
	metadataXML, err := h.iam.SAMLService.GenerateSpMetadata()
	if err != nil {
		panic(fmt.Errorf("cannot generate metadata: %w", err))
	}

	w.Header().Set("Content-Type", "application/samlmetadata+xml")
	w.WriteHeader(http.StatusOK)
	w.Write(metadataXML)
}

func (h *SAMLHandler) ConsumeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := r.ParseForm()
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("cannot parse form"))
		return
	}

	samlResponse := r.FormValue("SAMLResponse")
	relayState := r.FormValue("RelayState")

	configID, err := gid.ParseGID(relayState)
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid relay state"))
		return
	}

	user, membership, err := h.iam.SAMLService.HandleAssertion(ctx, samlResponse, configID)
	if err != nil {
		httpserver.RenderError(w, http.StatusUnauthorized, err)
		return
	}

	rootSession := SessionFromContext(ctx)

	switch {
	case rootSession == nil:
		rootSession, err = h.iam.AuthService.OpenSessionWithSAML(ctx, user.ID, membership.OrganizationID)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot open root session", log.Error(err))
			h.renderInternalServerError(w, r)
			return
		}
	case rootSession.IdentityID != user.ID:
		err = h.iam.SessionService.CloseSession(ctx, rootSession.ID)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot close session", log.Error(err))
			h.renderInternalServerError(w, r)
			return
		}

		rootSession, err = h.iam.AuthService.OpenSessionWithSAML(ctx, user.ID, membership.OrganizationID)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot open root session", log.Error(err))
			h.renderInternalServerError(w, r)
			return
		}
	}

	_, _, err = h.iam.SessionService.AssumeOrganizationSession(ctx, rootSession.ID, membership.OrganizationID)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot assume organization session", log.Error(err))
		h.renderInternalServerError(w, r)
		return
	}

	securecookie.Set(w, h.cookieConfig, rootSession.ID.String())
	redirectURL := h.baseURL.WithPath("/organizations/" + membership.OrganizationID.String()).MustString()
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h *SAMLHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	samlConfigIDParam := chi.URLParam(r, "samlConfigID")
	if samlConfigIDParam == "" {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing SAML config ID"))
		return
	}

	samlConfigID, err := gid.ParseGID(samlConfigIDParam)
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid SAML config ID"))
		return
	}

	url, err := h.iam.SAMLService.InitiateLogin(ctx, samlConfigID)
	if err != nil {
		panic(fmt.Errorf("cannot initiate SAML login: %w", err))
	}

	http.Redirect(w, r, url.String(), http.StatusFound)
}
