package webhook_manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func SendMessage(url string, message Message) error {
	JsonMessage := new(bytes.Buffer)

	err := json.NewEncoder(JsonMessage).Encode(message)
	if err != nil {
		return err
	}

	response, err := http.Post(url, "application/json", JsonMessage)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 && response.StatusCode != 204 {
		defer response.Body.Close()

		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf(string(body))
	}

	return nil
}
