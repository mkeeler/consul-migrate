package migrate

import (
	"fmt"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-hclog"
)

type importer struct {
	client    *api.Client
	logger    hclog.Logger
	opts      *api.WriteOptions
	qopts     *api.QueryOptions
	policyMap map[string]string
	roleMap   map[string]string
}

func Import(client *api.Client, data *Data) error {
	imp := importer{
		client:    client,
		logger:    hclog.Default(),
		policyMap: make(map[string]string),
		roleMap:   make(map[string]string),
	}

	if ent, err := isEnterprise(client); err != nil {
		return fmt.Errorf("error determining whether Consul is OSS or Enterprise: %w", err)
	} else if ent {
		return imp.importEnterprise(data)
	} else {
		return imp.importOSS(data)
	}
}

func (imp *importer) WithLoggerAndOpts(logger hclog.Logger, opts *api.WriteOptions, qopts *api.QueryOptions) *importer {
	newImp := imp
	if logger != nil {
		newImp.logger = logger
	}
	if opts != nil {
		newImp.opts = opts
	}

	if qopts != nil {
		newImp.qopts = qopts
	}

	return newImp
}

func (imp *importer) importEnterprise(data *Data) error {
	imp.logger.Debug("importing data to Consul Enterprise")

	if !data.Enterprise {
		// the data was from oss so we just allow the data to go into
		// the default namespace
		aclData := data.ACLData
		return imp.importACLData(&aclData)
	}

	for _, ns := range data.Namespaces {
		if err := imp.importNamespace(&ns); err != nil {
			return err
		}
	}

	return nil
}

func (imp *importer) importOSS(data *Data) error {
	imp.logger.Debug("importing data to Consul OSS")

	aclData := data.ACLData
	if data.Enterprise {
		// if we are importing data from enterprise to oss
		// then the only stuff we can import is in the default ns
		aclData = data.Namespaces["default"].ACLData
	}

	return imp.importACLData(&aclData)
}

func (imp *importer) importNamespace(nsData *NamespaceData) error {
	ns := imp.client.Namespaces()

	nsData.Definition.CreateIndex = 0
	nsData.Definition.ModifyIndex = 0

	newNS, _, err := ns.Create(&nsData.Definition, imp.opts)
	if err != nil {
		return fmt.Errorf("error creating namespace %s: %w", nsData.Definition.Name, err)
	}

	logger := imp.logger.With("ns", nsData.Definition.Name)
	logger.Info("created Namespace")

	newImp := imp.WithLoggerAndOpts(logger, &api.WriteOptions{Namespace: newNS.Name}, &api.QueryOptions{Namespace: newNS.Name})
	return newImp.importACLData(&nsData.ACLData)
}

func (imp *importer) importACLData(aclData *ACLData) error {
	if err := imp.importACLPolicies(aclData.ACLPolicies); err != nil {
		return fmt.Errorf("failed to import acl policies: %w", err)
	}

	if err := imp.importACLRoles(aclData.ACLRoles); err != nil {
		return fmt.Errorf("failed to import acl roles: %w", err)
	}

	if err := imp.importACLTokens(aclData.ACLTokens); err != nil {
		return fmt.Errorf("failed to import acl tokens: %w", err)
	}

	return nil
}

func (imp *importer) importACLPolicies(policies map[string]api.ACLPolicy) error {
	acls := imp.client.ACL()

	for policyID, policy := range policies {
		policy.CreateIndex = 0
		policy.ModifyIndex = 0
		policy.Hash = nil
		policy.ID = ""

		newPolicy, _, err := acls.PolicyCreate(&policy, imp.opts)
		if err != nil {
			return fmt.Errorf("failed to create policy: %w", err)
		}

		imp.logger.Info("created ACL Policy", "id", newPolicy.ID, "from", policyID)

		imp.policyMap[policyID] = newPolicy.ID
	}

	return nil
}

func (imp *importer) importACLRoles(roles map[string]api.ACLRole) error {
	acls := imp.client.ACL()

	for roleID, role := range roles {
		role.CreateIndex = 0
		role.ModifyIndex = 0
		role.Hash = nil
		role.ID = ""

		// map the old policy ids to the new ids
		for _, link := range role.Policies {
			link.ID = imp.policyMap[link.ID]
		}

		newRole, _, err := acls.RoleCreate(&role, imp.opts)
		if err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}

		imp.logger.Info("created ACL Role", "id", newRole.ID, "from", roleID)

		imp.roleMap[roleID] = newRole.ID
	}

	return nil
}

func (imp *importer) importACLTokens(tokens map[string]api.ACLToken) error {
	acls := imp.client.ACL()

	for _, token := range tokens {
		token.CreateIndex = 0
		token.ModifyIndex = 0
		token.Hash = nil

		// map the old policy ids to the new ids
		for _, link := range token.Policies {
			link.ID = imp.policyMap[link.ID]
		}

		// map the old role ids to the new ids
		for _, link := range token.Roles {
			link.ID = imp.roleMap[link.ID]
		}

		if token.AccessorID == "00000000-0000-0000-0000-000000000002" {
			_, _, err := acls.TokenUpdate(&token, imp.opts)
			if err != nil {
				return fmt.Errorf("failed to update anonymous token: %w", err)
			}
			imp.logger.Info("updated anonymous ACL Token", "accessor-id", token.AccessorID)
		} else {
			_, _, err := acls.TokenCreate(&token, imp.opts)
			if err != nil {
				return fmt.Errorf("failed to create token: %w", err)
			}
			imp.logger.Info("created ACL Token", "accessor-id", token.AccessorID)
		}

	}
	return nil
}
