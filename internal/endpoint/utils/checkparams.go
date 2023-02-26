package utils

import (
	"net/http"
)

func CheckParameters(r *http.Request, params []string) bool {
	// make sure the parameters are all present
	for _, p := range params {
		v := r.FormValue(p)
		if v == "" {
			return false
		}
	}

	return true
}
