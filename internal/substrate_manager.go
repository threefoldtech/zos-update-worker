package internal

import (
	client "github.com/threefoldtech/substrate-client"
)

type substrateClient struct {
	client *client.Substrate
}

// new instance of substrateClient
func newSubstrateClient(url ...string) (substrateClient, error) {
	manager := client.NewManager(url...)
	substrate, err := manager.Substrate()
	return substrateClient{substrate}, err
}

// get the latest version for the provided substrate url (main, test, qa)
func (s *substrateClient) checkVersion() (string, error) {

	version, err := s.client.GetZosVersion()
	if err != nil {
		return "", err
	}

	return *version, nil
}
