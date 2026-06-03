# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-03

### Added

- Typed `SecretStore` model of a Dapr SecretStore component (name, namespace,
  backend, spec metadata, scopes) with literal-value and `secretKeyRef` metadata
  entries.
- `Backend` typed string with named constants for the standard public Dapr
  secret-store backends (local file/env, kubernetes, hashicorp.vault,
  aws.secretmanager, aws.parameterstore, azure.keyvault, gcp.secretmanager).
- One carrier interface `Renderable` (`RenderYAML() ([]byte, error)`) and one
  free render verb `Render` — the borealis Law 5 shape. Rendering marshals
  through `go.yaml.in/yaml/v3` (never raw string formatting).
- Canonical constructor `New(name, backend, opts ...Option)` with `With*`
  options, and the config-driven peer `FromConfig(cfg Config)` (shikumi-shaped
  struct with `yaml` tags).
- `Validate` / `MustRender`: structural validation before emission.
- Table-driven tests including a YAML round-trip check and a manifest-to-component
  end-to-end test.

[Unreleased]: https://github.com/pleme-io/dapr-component-generator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/pleme-io/dapr-component-generator/releases/tag/v0.1.0
