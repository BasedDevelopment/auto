package routes

import (
	"net/http"

	"github.com/BasedDevelopment/auto/internal/util"
	eUtil "github.com/BasedDevelopment/eve/pkg/util"
)

func CreateDomain(w http.ResponseWriter, r *http.Request) {
	req := new(util.DomainCreateRequest)
	if err := util.ParseRequest(r, req); err != nil {
		eUtil.WriteError(w, r, err, http.StatusBadRequest, "Failed to parse request")
		return
	}

	//TODO: make cloudinit iso
	//TODO: make disk

	//TODO: Pass actual params to be called by virt-install
	if err := HV.CreateDomain(); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to create domain")
		return
	}

	if err := eUtil.WriteResponse("", w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
		return
	}
}
