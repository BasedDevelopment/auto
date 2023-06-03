package routes

import (
	"fmt"
	"net/http"

	"github.com/BasedDevelopment/auto/internal/controllers"
	"github.com/BasedDevelopment/auto/pkg/models"
	eUtil "github.com/BasedDevelopment/eve/pkg/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var HV = controllers.Hypervisor

func GetDomains(w http.ResponseWriter, r *http.Request) {
	var resp []models.VM
	for _, vm := range controllers.Hypervisor.VMs {
		resp = append(resp, *vm)
	}
	if err := eUtil.WriteResponse(resp, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}

func getDomain(r *http.Request) (*models.VM, error) {
	domidStr := chi.URLParam(r, "domain")
	domid, err := uuid.Parse(domidStr)

	if err != nil {
		return nil, err
	}

	vm, ok := HV.VMs[domid]
	if !ok {
		return nil, fmt.Errorf("Domain not found")
	}

	return vm, nil
}

func GetDomain(w http.ResponseWriter, r *http.Request) {
	domain, err := getDomain(r)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusNotFound, "Invalid domain ID or can't be found")
		return
	}

	if err := eUtil.WriteResponse(domain, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}

func DeleteDomain(w http.ResponseWriter, r *http.Request) {
	domain, err := getDomain(r)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusNotFound, "Invalid domain ID or can't be found")
		return
	}

	for _, disk := range domain.Storages {
		if err := HV.DeleteDiskFile(disk.Path); err != nil {
			eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to delete disk")
			return
		}
	}

	if err := HV.DestroyVM(domain); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to destroy domain")
		return
	}

	if err := HV.UndefineVM(domain); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to undefine domain")
		return
	}

	if err := eUtil.WriteResponse(domain, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}
