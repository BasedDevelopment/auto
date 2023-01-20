package routes

import (
	"net/http"

	"github.com/BasedDevelopment/auto/internal/controllers"
	eUtil "github.com/BasedDevelopment/eve/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetDomains(w http.ResponseWriter, r *http.Request) {
	resp := controllers.Hypervisor.VMs
	if err := eUtil.WriteResponse(resp, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}

func GetDomainState(w http.ResponseWriter, r *http.Request) {
	domidStr := chi.URLParam(r, "domain")
	domid, err := uuid.Parse(domidStr)

	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	hv := controllers.Hypervisor
	state, err := controllers.Hypervisor.GetVMState(hv.VMs[domid])
	if err := eUtil.WriteResponse(state, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}
