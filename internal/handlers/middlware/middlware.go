package middlware

import (
	"net/http"
	"strings"
)

func CkeckIfAllCorrect(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(res http.ResponseWriter, req *http.Request) {

			if req.Method == http.MethodPost {

				if !strings.HasPrefix(req.URL.RequestURI(), "/update/") {
					http.Error(res, "Incorrect type for Get request", http.StatusBadRequest)
				}

				pathWoSlash, _ := strings.CutSuffix(req.URL.Path, "/")
				pathWoSlash, _ = strings.CutPrefix(pathWoSlash, "/")
				reqData := strings.Split(pathWoSlash, "/")

				if len(reqData) < 3 {
					http.Error(res, "Not enough args (No name)", http.StatusNotFound)
				}

				if len(reqData) < 4 {
					http.Error(res, "Not enough args (No type)", http.StatusBadRequest)
				}

			} else if req.Method == http.MethodGet {

				if !strings.HasPrefix(req.URL.RequestURI(), "/value/") && !(req.URL.RequestURI() == "/") {
					http.Error(res, "Incorrect command for Post request", http.StatusBadRequest)
				}
			}

			h.ServeHTTP(res, req)

		})
}
