# dapr-component-generator

A typed model of a Dapr SecretStore component that renders to a canonical Dapr
Component `metadata.yaml` through a YAML emitter — never via raw string
formatting. A borealis-style PUBLIC codegen primitive: one carrier interface,
one render verb, `New(required, opts...)` + `FromConfig(cfg)`.

## What

`SecretStore` is a typed Go value with the Dapr Component shape — name,
namespace, backend, spec metadata, and scopes. `Render` (the one verb) marshals
it through `go.yaml.in/yaml/v3` into a valid `metadata.yaml`. The `Backend` type
carries named constants for the standard public Dapr secret-store backends
(HashiCorp Vault, AWS Secrets Manager / Parameter Store, Azure Key Vault, GCP
Secret Manager, Kubernetes, local file/env), and accepts any other Dapr-
registered backend string.

## Why

A Dapr Component is normally hand-written YAML, which drifts: a mistyped key, a
missing `version`, an unquoted value. Modeling the component as a typed value
and rendering through a YAML emitter makes the document always valid — the
emitter owns escaping, indentation, and key ordering, so a generator cannot
produce a malformed manifest. `Validate` rejects a structurally-incomplete
component before any bytes are emitted.

It follows the borealis laws every `*-go` primitive obeys: one behaviour carrier
(`Renderable`, Law 5), one render verb (`Render`), weight import-gated — the only
dependency is the YAML emitter the fleet already uses (Law 6).

WORLDS-SEPARATE: this is a generic, PUBLIC Dapr SecretStore emitter. It names no
private org surface; an org-specific backend (for example a custom or vendor
secret store) is supplied as a plain `Backend` string by the caller.

## Install

```sh
go get github.com/pleme-io/dapr-component-generator
```

Or via the substrate Nix flake (pull-model, tag-only release):

```nix
inputs.dapr-component-generator.url = "github:pleme-io/dapr-component-generator";
```

## Usage

Built on: `go.yaml.in/yaml/v3` (the canonical yaml.v3 fork the pleme-io Go fleet
uses) and the Go standard library.

The canonical constructor is `New(name, backend, opts ...Option)`; the
config-driven peer is `FromConfig(cfg)`.

```go
package main

import (
	"fmt"

	dapr "github.com/pleme-io/dapr-component-generator"
)

func main() {
	ss := dapr.New("vault", dapr.BackendHashicorpVault,
		dapr.WithNamespace("prod"),
		dapr.WithMetadata("vaultAddr", "https://vault:8200"),
		dapr.WithSecretRef("token", "vault-token", "token"),
		dapr.WithScope("orders"),
	)
	out, err := dapr.Render(ss)
	if err != nil {
		panic(err)
	}
	fmt.Print(string(out))
}
```

renders a valid `metadata.yaml`:

```yaml
apiVersion: dapr.io/v1alpha1
kind: Component
metadata:
    name: vault
    namespace: prod
spec:
    type: secretstores.hashicorp.vault
    version: v1
    metadata:
        - name: vaultAddr
          value: https://vault:8200
        - name: token
          secretKeyRef:
            name: vault-token
            key: token
scopes:
    - orders
```

## Configuration

`FromConfig(cfg Config)` builds the same `SecretStore` from a typed,
shikumi-shaped struct (plain Go struct with `yaml` tags) — so a generator that
loads a YAML manifest gets the identical surface as the imperative `New` builder.
A manifest:

```yaml
name: aws-store
backend: aws.secretmanager
namespace: default
metadata:
  - name: region
    value: us-east-1
  - name: accessKey
    secretName: aws-creds
    secretKey: access-key-id
scopes:
  - orders
```

loads into `Config` (via any yaml decoder) and renders the component:

```go
var cfg dapr.Config
_ = yaml.Unmarshal(manifestBytes, &cfg)
out, _ := dapr.Render(dapr.FromConfig(cfg))
```

## Release

Pull-model, tag-only (Go): a semver git tag is pushed and `proxy.golang.org`
fetches the module lazily — there is no registry upload step. Versions follow
SemVer and are recorded in [CHANGELOG.md](./CHANGELOG.md) in Keep-a-Changelog
format. The flake's `apps.{release,bump}` delegate to substrate's
language-generic `forge tool <verb> --language go`.
