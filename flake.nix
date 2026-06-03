# flake.nix — dapr-component-generator (GSDS Biblioteca) via substrate's
# goLibraryFlakeBuilder. vendorHash is OMITTED → spec-sourced (__from-spec__):
# the single dep (go.yaml.in/yaml/v3) is pinned on the proxy and the vendorHash
# is computed at release time. Pre-publish proof is `go test` (green) —
# `GOTOOLCHAIN=local go build ./... && go test ./...`.
{
  description = "dapr-component-generator — typed Dapr SecretStore component model rendered to metadata.yaml via a YAML emitter";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    substrate = {
      # Published repo uses: url = "github:pleme-io/substrate";
      url = "git+file:///Users/drzzln/code/github/pleme-io/substrate";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = inputs @ { self, nixpkgs, substrate, ... }:
    (import substrate.goLibraryFlakeBuilder { inherit nixpkgs; }) {
      name = "dapr-component-generator";
      version = "0.1.0";
      src = self;
      repo = "pleme-io/dapr-component-generator";
    };
}
