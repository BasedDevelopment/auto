package routes

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	eUtil "github.com/BasedDevelopment/eve/pkg/util"
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

	// Since libvirt's vnc doesn't accept any path, we will rewrite it to empty since proxyer will add it
	r.URL.Path = ""
	wsUrl := &url.URL{Scheme: "http", Host: "localhost:" + port}

	proxy := httputil.NewSingleHostReverseProxy(wsUrl)
	proxy.ServeHTTP(w, r)
}
