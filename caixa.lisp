(defcaixa dapr-component-generator
  :kind :Biblioteca
  :ecosystem :go
  :description "Typed Dapr SecretStore component model that renders to a canonical Dapr Component metadata.yaml via a YAML emitter (go.yaml.in/yaml/v3), never raw string formatting. One Renderable carrier, one Render verb, New(name, backend, opts...) + FromConfig(cfg). Generic/public — names no private backend (WORLDS-SEPARATE)."
  :module "github.com/pleme-io/dapr-component-generator")
