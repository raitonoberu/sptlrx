package mopidy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sptlrx/player"
)

func New(address string) *Client {
	return &Client{address: address}
}

// Client implements player.Player
type Client struct {
	address string
}

func (c *Client) get(method string, out interface{}) error {
	body := requestBody{
		JsonRPC: "2.0",
		ID:      1,
		Method:  method,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/mopidy/rpc", c.address)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) State() (*player.State, error) {
	var state stateResponse
	err := c.get("core.playback.get_state", &state)
	if err != nil {
		return nil, err
	}

	var current currentResponse
	err = c.get("core.playback.get_current_track", &current)
	if err != nil {
		return nil, err
	}

	var position positionResponse
	err = c.get("core.playback.get_time_position", &position)
	if err != nil {
		return nil, err
	}

	var artist string
	for i, a := range current.Result.Artists {
		if i != 0 {
			artist += " "
		}
		artist += a.Name
	}

	query := artist + " " + current.Result.Name

	return &player.State{
		Track:    player.TrackMetadata{
			ID:    current.Result.URI,
			Query: query,
		},
		Position: position.Result,
		Playing:  state.Result == "playing",
	}, err
}

type requestBody struct {
	JsonRPC string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Method  string `json:"method"`
}

type currentResponse struct {
	Result struct {
		URI     string `json:"uri"`
		Name    string `json:"name"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
	} `json:"result"`
}

type stateResponse struct {
	Result string `json:"result"`
}

type positionResponse struct {
	Result int `json:"result"`
}
