package routes

import (
	"net/http"

	"github.com/BasedDevelopment/auto/internal/util"
	eUtil "github.com/BasedDevelopment/eve/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func CreateDomain(w http.ResponseWriter, r *http.Request) {
	domID, err := uuid.Parse(chi.URLParam(r, "domain"))
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusBadRequest, "invalid domain id")
		return
	}

	req := new(util.DomainCreateRequest)
	if err := util.ParseRequest(r, req); err != nil {
		eUtil.WriteError(w, r, err, http.StatusBadRequest, "Failed to parse request")
		return
	}

	if err := HV.CreateDomain(domID, req); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to create domain")
		return
	}

	if err := eUtil.WriteResponse("", w, http.StatusCreated); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
		return
	}
}
