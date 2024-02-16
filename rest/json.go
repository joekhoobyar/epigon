package rest

import (
	"encoding/json"
	"io"
	"net/http"
)

func unmarshall(rq *http.Request, target any) error {
	body, err := io.ReadAll(rq.Body)
	defer rq.Body.Close()
	if err == nil {
		err = json.Unmarshal(body, target)
	}
	return err
}
