package routes

import (
	"net/http"

	"github.com/BasedDevelopment/auto/internal/util"
	eUtil "github.com/BasedDevelopment/eve/pkg/util"
)

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
