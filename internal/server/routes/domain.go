package routes

import (
	"fmt"
	"net/http"

	"github.com/BasedDevelopment/auto/internal/controllers"
	"github.com/BasedDevelopment/auto/internal/util"
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

func GetDomainState(w http.ResponseWriter, r *http.Request) {
	domain, err := getDomain(r)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusNotFound, "Invalid domain ID or can't be found")
		return
	}

	state, err := HV.GetVMState(domain)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to get domain state")
		return
	}

	if err := eUtil.WriteResponse(state, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}

func SetDomainState(w http.ResponseWriter, r *http.Request) {
	domain, err := getDomain(r)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusNotFound, "Invalid domain ID or can't be found")
		return
	}

	req := new(util.SetDomainStateRequest)
	if err := util.ParseRequest(r, req); err != nil {
		eUtil.WriteError(w, r, err, http.StatusBadRequest, "Failed to parse request")
		return
	}

	switch req.State {
	case "start":
		if err := HV.Libvirt.VMStart(domain.Domain); err != nil {
			eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to start VM")
			return
		}
	case "reboot":
		if err := HV.Libvirt.VMReboot(domain.Domain); err != nil {
			eUtil.WriteError(w, r, err, http.StatusInternalServerError, "failed to reboot virtual machine")
			return
		}
	case "poweroff":
		if err := HV.Libvirt.VMPowerOff(domain.Domain); err != nil {
			eUtil.WriteError(w, r, err, http.StatusInternalServerError, "failed to power off virtual machine")
			return
		}
	case "stop":
		if err := HV.Libvirt.VMStop(domain.Domain); err != nil {
			eUtil.WriteError(w, r, err, http.StatusInternalServerError, "failed to stop virtual machine")
			return
		}
	case "reset":
		if err := HV.Libvirt.VMReset(domain.Domain); err != nil {
			eUtil.WriteError(w, r, err, http.StatusInternalServerError, "failed to reset virtual machine")
			return
		}
	}

	state, err := HV.GetVMState(domain)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to get domain state")
		return
	}

	if err := eUtil.WriteResponse(state, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}
