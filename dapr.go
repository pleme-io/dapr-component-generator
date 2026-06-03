// Package dapr is a typed model of a Dapr SecretStore component that renders to
// a canonical Dapr Component `metadata.yaml` via a YAML emitter — never via raw
// string formatting.
//
// A Dapr Component is normally authored as hand-written YAML, which drifts: a
// mistyped key, a missing `version`, an unquoted secret value. This library
// makes the component a typed Go value ([SecretStore]) with `yaml`-tagged
// fields, and renders it through [SecretStore.Render], which marshals via
// go.yaml.in/yaml/v3 (the canonical yaml.v3 fork the fleet already uses). The
// emitter owns escaping, indentation, and key ordering, so the rendered
// document is always a valid Dapr Component manifest.
//
// # The one render verb (borealis Law 5 — behaviour carrier)
//
// [SecretStore] implements the single carrier interface [Renderable]:
//
//	type Renderable interface {
//		RenderYAML() ([]byte, error)
//	}
//
// and the package exposes ONE free render verb, [Render], that dispatches over
// the carrier. There is no per-type Marshal / ToYAML variant.
//
// # Construction (GSDS §3.5 — New(required, opts...))
//
// The canonical constructor is New(name, backend, opts ...Option) — the
// required pieces a SecretStore cannot omit (its name and backend), then
// functional options:
//
//	ss := dapr.New("vault", dapr.BackendHashicorpVault,
//		dapr.WithMetadata("vaultAddr", "https://vault:8200"),
//		dapr.WithMetadata("vaultTokenMountPath", "/run/secrets/token"),
//		dapr.WithScope("orders"),
//	)
//	out, _ := dapr.Render(ss)   // []byte of metadata.yaml
//
// FromConfig builds the same [SecretStore] from a typed [Config] (a shikumi-
// shaped struct with yaml tags), so a generator driven by a manifest gets the
// identical surface as the imperative builder.
//
// # WORLDS-SEPARATE
//
// This is a PUBLIC, generic Dapr SecretStore emitter. [Backend] is an open typed
// string with named constants for the standard public Dapr secret-store
// backends (HashiCorp Vault, AWS Secrets Manager, Azure Key Vault, GCP Secret
// Manager, local file/env). It names no private org surface; any org-specific
// backend is supplied as a plain [Backend] string by the caller.
package dapr

// Renderable is any typed value that can render itself to a Dapr manifest. It is
// the single carrier interface behind the one render verb [Render] (borealis
// Law 5).
type Renderable interface {
	RenderYAML() ([]byte, error)
}

// Render is THE one render verb. It dispatches over the [Renderable] carrier and
// returns the rendered manifest bytes. There is no other YAML emitter in this
// package — every value is serialized through here.
func Render(r Renderable) ([]byte, error) { return r.RenderYAML() }

// Backend identifies a Dapr secret-store backend. It is an open typed string:
// the named constants cover the standard public Dapr backends, and a caller may
// pass any other Dapr-registered backend string directly.
type Backend string

// Standard public Dapr secret-store backends (the `secretstores.<backend>` type
// suffix). See https://docs.dapr.io/reference/components-reference/supported-secret-stores/.
const (
	// BackendLocalFile is the `local.file` JSON-file secret store.
	BackendLocalFile Backend = "local.file"
	// BackendLocalEnv is the `local.env` environment-variable secret store.
	BackendLocalEnv Backend = "local.env"
	// BackendKubernetes is the built-in `kubernetes` secret store.
	BackendKubernetes Backend = "kubernetes"
	// BackendHashicorpVault is the `hashicorp.vault` secret store.
	BackendHashicorpVault Backend = "hashicorp.vault"
	// BackendAWSSecretManager is the `aws.secretmanager` secret store.
	BackendAWSSecretManager Backend = "aws.secretmanager"
	// BackendAWSParameterStore is the `aws.parameterstore` secret store.
	BackendAWSParameterStore Backend = "aws.parameterstore"
	// BackendAzureKeyVault is the `azure.keyvault` secret store.
	BackendAzureKeyVault Backend = "azure.keyvault"
	// BackendGCPSecretManager is the `gcp.secretmanager` secret store.
	BackendGCPSecretManager Backend = "gcp.secretmanager"
)

// componentType returns the Dapr `spec.type` for a SecretStore backend:
// `secretstores.<backend>`.
func (b Backend) componentType() string { return "secretstores." + string(b) }
