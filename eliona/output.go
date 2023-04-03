package eliona

import (
	"fmt"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
)

// Generates a websocket connection to the database and listens for any updates
// on assets (only output attributes). Returns a channel with all changes.
func ListenForOutputChanges() (chan api.Data, error) {
	conn, err := http.NewWebSocketConnectionWithApiKey(common.Getenv("API_ENDPOINT", "")+"/data-listener?dataSubtype=output", "X-API-Key", common.Getenv("API_TOKEN", ""))
	if err != nil {
		return nil, fmt.Errorf("creating websocket: %v", err)
	}
	outputs := make(chan api.Data)
	go http.ListenWebSocket(conn, outputs)
	return outputs, nil
}

func ResetDoorlockState(assetId int32) error {
	if err := upsertData(
		api.SUBTYPE_OUTPUT,
		assetId,
		doorlockOutputDataPayload{
			Open: false,
		},
	); err != nil {
		return fmt.Errorf("resetting doorlock output data: %v", err)
	}
	return nil
}
