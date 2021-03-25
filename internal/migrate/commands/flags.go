package commands

import (
	"flag"

	"github.com/hashicorp/consul/api"
)

type httpFlags struct {
	// client api flags
	address       stringValue
	token         stringValue
	tokenFile     stringValue
	caFile        stringValue
	caPath        stringValue
	certFile      stringValue
	keyFile       stringValue
	tlsServerName stringValue
}

func (f *httpFlags) flags() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Var(&f.address, "http-addr",
		"The `address` and port of the Consul HTTP agent. The value can be an IP "+
			"address or DNS address, but it must also include the port. This can "+
			"also be specified via the CONSUL_HTTP_ADDR environment variable. The "+
			"default value is http://127.0.0.1:8500. The scheme can also be set to "+
			"HTTPS by setting the environment variable CONSUL_HTTP_SSL=true.")
	fs.Var(&f.token, "token",
		"ACL token to use in the request. This can also be specified via the "+
			"CONSUL_HTTP_TOKEN environment variable. If unspecified, the query will "+
			"default to the token of the Consul agent at the HTTP address.")
	fs.Var(&f.tokenFile, "token-file",
		"File containing the ACL token to use in the request instead of one specified "+
			"via the -token argument or CONSUL_HTTP_TOKEN environment variable. "+
			"This can also be specified via the CONSUL_HTTP_TOKEN_FILE environment variable.")
	fs.Var(&f.caFile, "ca-file",
		"Path to a CA file to use for TLS when communicating with Consul. This "+
			"can also be specified via the CONSUL_CACERT environment variable.")
	fs.Var(&f.caPath, "ca-path",
		"Path to a directory of CA certificates to use for TLS when communicating "+
			"with Consul. This can also be specified via the CONSUL_CAPATH environment variable.")
	fs.Var(&f.certFile, "client-cert",
		"Path to a client cert file to use for TLS when 'verify_incoming' is enabled. This "+
			"can also be specified via the CONSUL_CLIENT_CERT environment variable.")
	fs.Var(&f.keyFile, "client-key",
		"Path to a client key file to use for TLS when 'verify_incoming' is enabled. This "+
			"can also be specified via the CONSUL_CLIENT_KEY environment variable.")
	fs.Var(&f.tlsServerName, "tls-server-name",
		"The server name to use as the SNI host when connecting via TLS. This "+
			"can also be specified via the CONSUL_TLS_SERVER_NAME environment variable.")
	return fs
}

func (f *httpFlags) apiClient() (*api.Client, error) {
	c := api.DefaultConfig()

	f.mergeOntoConfig(c)

	return api.NewClient(c)
}

func (f *httpFlags) mergeOntoConfig(c *api.Config) {
	f.address.Merge(&c.Address)
	f.token.Merge(&c.Token)
	f.tokenFile.Merge(&c.TokenFile)
	f.caFile.Merge(&c.TLSConfig.CAFile)
	f.caPath.Merge(&c.TLSConfig.CAPath)
	f.certFile.Merge(&c.TLSConfig.CertFile)
	f.keyFile.Merge(&c.TLSConfig.KeyFile)
	f.tlsServerName.Merge(&c.TLSConfig.Address)
}

func flagMerge(dst, src *flag.FlagSet) {
	if dst == nil {
		panic("dst cannot be nil")
	}
	if src == nil {
		return
	}
	src.VisitAll(func(f *flag.Flag) {
		dst.Var(f.Value, f.Name, f.Usage)
	})
}

// stringValue provides a flag value that's aware if it has been set.
type stringValue struct {
	v *string
}

// merge will overlay this value if it has been set.
func (s *stringValue) Merge(onto *string) {
	if s.v != nil {
		*onto = *(s.v)
	}
}

// Set implements the flag.Value interface.
func (s *stringValue) Set(v string) error {
	if s.v == nil {
		s.v = new(string)
	}
	*(s.v) = v
	return nil
}

// String implements the flag.Value interface.
func (s *stringValue) String() string {
	var current string
	if s.v != nil {
		current = *(s.v)
	}
	return current
}
