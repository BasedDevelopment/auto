package util

import (
	"encoding/json"
	"net/http"

	"github.com/BasedDevelopment/eve/pkg/status"
	validation "github.com/go-ozzo/ozzo-validation/v4"
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
	Hostname  string        `json:"hostname"`
	CPU       int           `json:"cpu"`
	Memory    int           `json:"memory"`
	State     status.Status `json:"state"`
	Image     string        `json:"image"`
	Cloud     bool          `json:"cloud"`
	OS        string        `json:"os"`
	OSVariant string        `json:"os_variant"`
	userData  string        `json:"userData"`
	metaData  string        `json:"metaData"`
	Disk      []struct {
		Id   int    `json:"id"`
		Size int    `json:"size"`
		Disk string `json:"disk"`
	} `json:"disk"`
	Iface []struct {
		Bridge string `json:"bridge"`
		MAC    string `json:"mac"`
	} `json:"iface"`
}

func (r *DomainCreateRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Hostname, validation.Required),
		validation.Field(&r.CPU, validation.Required),
		validation.Field(&r.Memory, validation.Required),
		validation.Field(&r.State, validation.Required),
	)
}

func ParseRequest[R Request, T Validatable[R]](r *http.Request, rq T) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(rq); err != nil {
		return err
	}

	return rq.Validate()
}
