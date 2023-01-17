package routes

import (
	"net/http"

	"github.com/BasedDevelopment/auto/internal/controllers"
	eUtil "github.com/BasedDevelopment/eve/pkg/util"
)

func GetHV(w http.ResponseWriter, r *http.Request) {
	hv := controllers.Hypervisor

	if err := eUtil.WriteResponse(hv, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}
