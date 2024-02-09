package combo

import (
	"sptlrx/lyrics"
	"sptlrx/player"
)

func New(providers []lyrics.Provider) (*ComboClient, error) {

	return &ComboClient{providers: providers}, nil
}

type ComboClient struct {
	providers []lyrics.Provider
}

func (c *ComboClient) Lyrics(state player.State) ([]lyrics.Line, error) {
	for _, p := range c.providers {
		provider := p
		newLines, err := provider.Lyrics(state)
		if err != nil || newLines == nil {
			continue
		} else {
			return newLines, nil
		}
	}
	return nil, nil
}

func (c *ComboClient) Name() string {
	return "COMBO"
}
