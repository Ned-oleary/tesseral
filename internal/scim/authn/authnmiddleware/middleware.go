package authnmiddleware

import (
	"net/http"
	"strings"

	"github.com/tesseral-labs/tesseral/internal/common/projectid"
	"github.com/tesseral-labs/tesseral/internal/scim/authn"
	"github.com/tesseral-labs/tesseral/internal/scim/store"
	"github.com/tesseral-labs/tesseral/internal/store/idformat"
)

func New(s *store.Store, p *projectid.Sniffer, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		projectID, err := p.GetProjectID(r.Header.Get("X-Tesseral-Host"))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		bearerToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if bearerToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		scimAPIKey, err := s.GetSCIMAPIKeyByToken(ctx, *projectID, bearerToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx = authn.NewContext(ctx, &authn.SCIMAPIKey{
			ID:             scimAPIKey.ID,
			OrganizationID: scimAPIKey.OrganizationID,
		}, idformat.Project.Format(*projectID))
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
