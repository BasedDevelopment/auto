package util

import (
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Validatable[T any] interface {
	Validate() error
	*T
}

type Request interface {
	SetDomainStateRequest |
		DomainCreateRequest
}

type SetDomainStateRequest struct {
	State string `json:"state"`
}

func (r *SetDomainStateRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.State, validation.Required, validation.In("start", "reboot", "poweroff", "stop", "reset")),
	)
}

type DomainCreateRequest struct {
	Id         string `json:"id"`
	Hostname   string `json:"hostname"`
	CPU        int    `json:"cpu"`
	Memory     int    `json:"memory"`
	Image      string `json:"image"`
	Cloud      bool   `json:"cloud"`
	CloudImage string `json:"cloud_image"`
	OSVariant  string `json:"os_variant"`
	UserData   string `json:"user_data"`
	MetaData   string `json:"meta_data"`
	Disk       []struct {
		Id   int    `json:"id"`
		Size int    `json:"size"`
		Path string `json:"path"`
	} `json:"disk"`
	Iface []struct {
		Bridge string `json:"bridge"`
		MAC    string `json:"mac"`
	} `json:"iface"`
}

func (r *DomainCreateRequest) Validate() error {
	if r.Cloud {
		if err := validation.ValidateStruct(r,
			validation.Field(&r.UserData, validation.Required),
			validation.Field(&r.MetaData, validation.Required),
		); err != nil {
			return err
		}
	}
	return validation.ValidateStruct(r,
		validation.Field(&r.Hostname, validation.Required, is.Domain),
		validation.Field(&r.CPU, validation.Required, validation.Min(1)),
		validation.Field(&r.Memory, validation.Required, validation.Min(1)),
		// Validation of disk and image path is not here due to import cycle
	)
}

func ParseRequest[R Request, T Validatable[R]](r *http.Request, rq T) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(rq); err != nil {
		return err
	}

	return rq.Validate()
}
