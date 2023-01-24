package util

import (
	"encoding/json"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Validatable[T any] interface {
	Validate() error
	*T
}

type Request interface {
	SetDomainStateRequest
}

type SetDomainStateRequest struct {
	State string `json:"state"`
}

func (r *SetDomainStateRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.State, validation.Required, validation.In("start", "reboot", "poweroff", "stop", "reset")),
	)
}

func ParseRequest[R Request, T Validatable[R]](r *http.Request, rq T) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(rq); err != nil {
		return err
	}

	return rq.Validate()
}
