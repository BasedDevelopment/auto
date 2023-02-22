package util

import (
	"encoding/json"
	"net/http"

	"github.com/BasedDevelopment/eve/pkg/status"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
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
	ID       uuid.UUID     `json:"id"`
	Hostname string        `json:"hostname"`
	CPU      int           `json:"cpu"`
	Memory   int           `json:"memory"`
	State    status.Status `json:"state"`
	Distro   string        `json:"distro"`
}

func (r *DomainCreateRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.ID, validation.Required),
		validation.Field(&r.Hostname, validation.Required),
		validation.Field(&r.CPU, validation.Required),
		validation.Field(&r.Memory, validation.Required),
		validation.Field(&r.State, validation.Required),
		validation.Field(&r.Distro, validation.Required),
	)
}

func ParseRequest[R Request, T Validatable[R]](r *http.Request, rq T) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(rq); err != nil {
		return err
	}

	return rq.Validate()
}
