package migrate

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-hclog"
)

func Export(client *api.Client) (*Data, error) {
	if ent, err := isEnterprise(client); err != nil {
		return nil, fmt.Errorf("error determining whether Consul is OSS or Enterprise: %w", err)
	} else if ent {
		hclog.L().Debug("exporting data from Consul Enterprise")
		return exportEnterprise(client)
	} else {
		hclog.L().Debug("exporting data from Consul OSS")
		return exportOSS(client)
	}
}

func exportEnterprise(client *api.Client) (*Data, error) {
	ns := client.Namespaces()

	hclog.L().Debug("gathering namespace list")
	nsList, _, err := ns.List(nil)
	if err != nil {
		return nil, fmt.Errorf("error listing namespaces: %w", err)
	}

	data := &Data{
		Enterprise: true,
		Namespaces: make(map[string]NamespaceData),
	}

	for _, ns := range nsList {
		// ignore deleted namespaces
		if ns.DeletedAt != nil && !ns.DeletedAt.IsZero() {
			hclog.L().Debug("ignoring deleted namespace", "ns", ns.Name)
			continue
		}

		opts := api.QueryOptions{
			Namespace: ns.Name,
		}

		hclog.L().Debug("exporting ACL data for namespace", "ns", ns.Name)
		aclData, err := exportACLData(client, &opts)
		if err != nil {
			return nil, fmt.Errorf("error exporting acl data for namespace %s: %w", ns.Name, err)
		}

		data.Namespaces[ns.Name] = NamespaceData{
			Definition: *ns,
			ACLData:    *aclData,
		}
	}

	return data, nil
}

func exportOSS(client *api.Client) (*Data, error) {
	hclog.L().Debug("exporting ACL data")
	aclData, err := exportACLData(client, nil)
	if err != nil {
		return nil, err
	}

	return &Data{ACLData: *aclData}, nil
}

func exportACLData(client *api.Client, opts *api.QueryOptions) (*ACLData, error) {
	policies, err := exportACLPolicies(client, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to export acl policies: %w", err)
	}

	roles, err := exportACLRoles(client, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to export acl roles: %w", err)
	}

	tokens, err := exportACLTokens(client, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to export acl tokens: %w", err)
	}

	return &ACLData{ACLPolicies: policies, ACLRoles: roles, ACLTokens: tokens}, nil
}

func exportACLPolicies(client *api.Client, opts *api.QueryOptions) (map[string]api.ACLPolicy, error) {
	acls := client.ACL()

	policyList, _, err := acls.PolicyList(opts)
	if err != nil {
		return nil, fmt.Errorf("error listing policies: %w", err)
	}

	policies := make(map[string]api.ACLPolicy)
	for _, policyStub := range policyList {
		if policyStub.ID == "00000000-0000-0000-0000-000000000001" {
			// no need to save off the global-management policy
			continue
		}

		policy, _, err := acls.PolicyRead(policyStub.ID, opts)
		if err != nil {
			return nil, fmt.Errorf("error reading policy: %w", err)
		}

		policies[policy.ID] = *policy
	}

	return policies, nil
}

func exportACLRoles(client *api.Client, opts *api.QueryOptions) (map[string]api.ACLRole, error) {
	acls := client.ACL()

	roleList, _, err := acls.RoleList(opts)
	if err != nil {
		return nil, fmt.Errorf("error listing roles: %w", err)
	}

	roles := make(map[string]api.ACLRole)
	for _, roleStub := range roleList {
		role, _, err := acls.RoleRead(roleStub.ID, opts)
		if err != nil {
			return nil, fmt.Errorf("error reading role: %w", err)
		}

		roles[role.ID] = *role
	}

	return roles, nil
}

func exportACLTokens(client *api.Client, opts *api.QueryOptions) (map[string]api.ACLToken, error) {
	acls := client.ACL()

	tokenList, _, err := acls.TokenList(opts)
	if err != nil {
		return nil, fmt.Errorf("error listing tokens: %w", err)
	}

	tokens := make(map[string]api.ACLToken)
	for _, tokenStub := range tokenList {
		token, _, err := acls.TokenRead(tokenStub.AccessorID, opts)
		if err != nil {
			return nil, fmt.Errorf("error reading token: %w", err)
		}

		tokens[token.AccessorID] = *token
	}

	return tokens, nil
}
