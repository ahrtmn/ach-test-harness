package instantverify

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
)

type InstantVerifyController interface {
	AppendRoutes(router *mux.Router) *mux.Router
}

func NewInstantVerifyController(logger log.Logger, service InstantVerifyService) InstantVerifyController {
	return &instantVerifyController{
		logger:  logger,
		service: service,
	}
}

type instantVerifyController struct {
	logger  log.Logger
	service InstantVerifyService
}

func (c *instantVerifyController) AppendRoutes(router *mux.Router) *mux.Router {
	router.
		Name("InstantVerify.get").
		Methods("GET").
		Path("/instantverify/{traceNumber}").
		HandlerFunc(c.getInstantVerification)

	return router
}

func (c *instantVerifyController) getInstantVerification(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	traceNumber := vars["traceNumber"]

	if traceNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		err := errors.New("traceNumber required for getInstantVerification")
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	verification, err := c.service.Search(SearchOptions{TraceNumber: traceNumber})
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(verification)
}
