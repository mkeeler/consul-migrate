package migrate

import "github.com/hashicorp/consul/api"

type Data struct {
	Enterprise bool                     `json:"enterprise,omitempty"`
	Namespaces map[string]NamespaceData `json:"namespaces,omitempty"`
	ACLData
}

type NamespaceData struct {
	Definition api.Namespace `json:"definition"`
	ACLData
}

type ACLData struct {
	// ACLPolicies is a mapping of the policy id to the policy definition
	ACLPolicies map[string]api.ACLPolicy `json:"acl_policies,omitempty"`
	// ACLRoles is a mapping of the role id to the role definition
	ACLRoles map[string]api.ACLRole `json:"acl_roles,omitempty"`
	// ACLTokens is a map of accessor id to acl token definition
	ACLTokens map[string]api.ACLToken `json:"acl_tokens,omitempty"`
}
