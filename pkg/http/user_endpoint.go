package http

import (
	"net/http"

	"github.com/quilla-hq/quilla/pkg/auth"
)

func (s *TriggerServer) userHandler(resp http.ResponseWriter, req *http.Request) {
	user := auth.GetAccountFromCtx(req.Context())
	response(user, 200, nil, resp, req)
}
