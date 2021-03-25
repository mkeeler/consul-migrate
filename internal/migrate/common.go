package migrate

import (
	"fmt"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-hclog"
)

func isEnterprise(client *api.Client) (bool, error) {
	hclog.L().Debug("retrieving agent info to determine if this is enterprise or oss")
	info, err := client.Agent().Self()
	if err != nil {
		return false, fmt.Errorf("error retrieving Consul info: %w", err)
	}

	vers, ok := info["Config"]["Version"].(string)
	if !ok {
		return false, fmt.Errorf("consul info version field is not a string")
	}
	return strings.Contains(vers, "+ent"), nil
}
