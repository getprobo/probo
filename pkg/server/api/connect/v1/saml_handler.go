package connect_v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
)

type SAMLHandler struct {
	iam          *iam.Service
	cookieConfig securecookie.Config
	baseURL      *baseurl.BaseURL
}

func NewSAMLHandler(iam *iam.Service, cookieConfig securecookie.Config, baseURL *baseurl.BaseURL) *SAMLHandler {
	return &SAMLHandler{iam: iam, cookieConfig: cookieConfig, baseURL: baseURL}
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

	session := SessionFromContext(ctx)
	if session == nil {
		h.iam.AuthService.OpenSessionWithoutPassword(ctx, user.ID, membership.OrganizationID)
	}

	// TODO open or update the organization session

	securecookie.Set(w, h.cookieConfig, session.ID.String())

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
