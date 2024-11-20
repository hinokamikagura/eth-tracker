package handler

import "context"

func (hc *HandlerContainer) checkWithinNetwork(ctx context.Context, contractAddress string, from string, to string) (bool, error) {
	exists, err := hc.cache.ExistsNetwork(ctx, contractAddress, from, to)
	if err != nil {
		return false, err
	}

	return exists, nil
}
