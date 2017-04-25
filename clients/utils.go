package clients

import "net/http"

func GetCredential(request *http.Request) string {
	return request.Header.Get("X-Vtex-Credential")
}
