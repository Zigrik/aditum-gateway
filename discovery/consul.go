package discovery

import (
	"fmt"

	consul "github.com/hashicorp/consul/api"
)

type ConsulClient struct {
	client *consul.Client
}

func NewConsulClient(addr string) (*ConsulClient, error) {
	config := consul.DefaultConfig()
	config.Address = addr
	client, err := consul.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}
	return &ConsulClient{client: client}, nil
}

// GetHealthyServices возвращает адреса здоровых инстансов сервиса (host:port)
func (c *ConsulClient) GetHealthyServices(serviceName string) ([]string, error) {
	entries, _, err := c.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return []string{}, nil
	}

	var addresses []string
	for _, entry := range entries {
		service := entry.Service
		addr := fmt.Sprintf("%s:%d", service.Address, service.Port)
		addresses = append(addresses, addr)
	}
	return addresses, nil
}
