package routes

import (
	"fmt"
	"net"
	"net/http"

	eUtil "github.com/BasedDevelopment/eve/pkg/util"
	"github.com/BasedDevelopment/eve/pkg/wsproxy"
)

func GetConsole(w http.ResponseWriter, r *http.Request) {
	domain, err := getDomain(r)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusNotFound, "Invalid domain ID or can't be found")
		return
	}

	port, err := HV.GetVMConsole(domain)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to get console port")
		return
	}

	wsUrl := fmt.Sprintf("ws://%s:%d", "localhost", port)

	peer, err := net.Dial("tcp", wsUrl)
	if err != nil {
		eUtil.WriteError(w, r, err, http.StatusInternalServerError, "Failed to dial TCP connection to ws")
		return
	}

	return wsproxy.WsProxy(w, r, peer)
}
