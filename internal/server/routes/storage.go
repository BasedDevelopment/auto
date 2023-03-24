package routes

import (
	"net/http"

	"github.com/BasedDevelopment/auto/internal/config"
	"github.com/BasedDevelopment/auto/internal/controllers"
	eUtil "github.com/BasedDevelopment/eve/pkg/util"
)

func GetStorages(w http.ResponseWriter, r *http.Request) {
	resp := config.Config.Storage
	if err := eUtil.WriteResponse(resp, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}

func GetImages(w http.ResponseWriter, r *http.Request) {
	resp := controllers.Images
	if err := eUtil.WriteResponse(resp, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}

func GetDisks(w http.ResponseWriter, r *http.Request) {
	resp := controllers.Disks
	if err := eUtil.WriteResponse(resp, w, http.StatusOK); err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to marshall/send response")
	}
}
