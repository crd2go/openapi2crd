# OpenAPI2CRD

Generates Kubernetes `CustomResourceDefinition` manifests from an OpenAPI 3.0 specification.

A configuration file maps OpenAPI schemas to CRD kinds and controls how fields are split between `spec` (writable) and `status` (read-only), which properties are sensitive or skipped, and how cross-resource references are represented.

It is specially suited for writing Kubernetes operators that manage resources behind well defined OpenAPI specs.

## Install

```shell
go install github.com/crd2go/openapi2crd@latest
```

Or build from source:

```shell
make build          # produces bin/openapi2crd
```

## Usage

```
openapi2crd -c config.yaml -o crds.yaml [flags]

Flags:
  -c, --config string    Path to the config file (required)
  -o, --output string    Path to output file or directory (required)
  -f, --force            Overwrite output if it already exists
      --multi-file       Write each CRD to a separate file in the output directory
      --crds string      Comma-separated Kind names to generate; defaults to "all"
```

### Examples

```shell
# Single output file
openapi2crd -c config.yaml -o crds.yaml

# One file per CRD
openapi2crd -c config.yaml -o crds/ --multi-file

# Generate a subset of kinds
openapi2crd -c config.yaml -o crds.yaml --crds Group,Cluster

# Force overwrite
openapi2crd -c config.yaml -o crds.yaml --force
```

## Configuration

```yaml
apiVersion: atlas2crd.mongodb.com/v1alpha1
kind: Config

spec:
  # Named OpenAPI specs that CRDs can reference.
  openapi:
    - name: v20250312
      package: go.mongodb.org/atlas-sdk/v20250312006/admin
      path: path/to/openapi.yaml

  # One or more named plugin sets. Mark one as default: true.
  # Plugin sets can inherit from another set via inheritFrom.
  pluginSets:
    - name: atlas
      default: true
      plugins:
        - base
        - major_version
        - parameters
        - connection_secret
        - mutual_exclusive_group
        - entry
        - status
        - sensitive_properties
        - skipped_properties
        - read_only_properties
        - read_write_properties
        - references
        - reference_extensions
        - mutual_exclusive_major_versions
        - atlas_sdk_version
        - print_conditions

  # CRDs to generate.
  crd:
    - gvk:
        group: example.io
        version: v1
        kind: MyResource
      categories: [mygroup]
      shortNames: [mr]
      # pluginSet: atlas   # optional; uses the default set if omitted
      mappings:
        - majorVersion: v20250312
          openAPIRef:
            name: v20250312
          parameters:
            path:
              name: '/api/v2/groups/{groupId}/resources'
              verb: post
            references:
              - name: groupRef
                property: $.groupId
                target:
                  type:
                    kind: Group
                    resource: groups
                    group: example.io
                    version: v1
                  properties:
                    - $.status.v20250312.id
          entry:
            schema: MyResourceSpec     # OpenAPI schema name for spec fields
            filters:
              readWriteOnly: true       # include only read-write properties
              skipProperties:
                - $.links
              sensitiveProperties:
                - $.password
          status:
            schema: MyResourceStatus
            filters:
              readOnly: true
              skipProperties:
                - $.links
```

### Mappings

Each CRD can have multiple `mappings`, one per API major version. A mapping ties together:

| Field | Description |
|---|---|
| `majorVersion` | Logical version label used to namespace status fields |
| `openAPIRef.name` | Which OpenAPI spec to read from |
| `parameters.path` | API path and verb to extract path parameters (e.g. `groupId`) |
| `parameters.references` | Cross-resource references that resolve a parameter from another CRD's status |
| `entry` | Schema and filters for the CRD `spec` (writable fields) |
| `status` | Schema and filters for the CRD `status` (read-only fields) |

### Filters

Applied independently to `entry` and `status` blocks:

| Filter | Description |
|---|---|
| `skipProperties` | JSONPath expressions for properties to remove |
| `readOnly: true` | Keep only properties marked `readOnly` in the OpenAPI spec |
| `readWriteOnly: true` | Keep only properties **not** marked `readOnly` |
| `sensitiveProperties` | JSONPath list — these fields are annotated as sensitive (e.g. for secret injection) |

### References

A reference replaces a scalar parameter field with a typed Kubernetes object reference. The operator can then resolve the actual value from the referenced resource's status at runtime.

```yaml
references:
  - name: groupRef
    property: $.groupId          # field to replace in the generated CRD spec
    target:
      type:
        kind: Group
        resource: groups
        group: example.io
        version: v1
      properties:
        - $.status.v20250312.id  # path on the target resource that provides the value
```

## Plugins

Plugins are run in the order listed in the plugin set. They fall into four categories:

**CRD-level**

| Plugin | Description |
|---|---|
| `base` | Scaffolds the CRD skeleton (metadata, names, scope) |
| `mutual_exclusive_major_versions` | Adds `oneOf` validation so only one major-version status block is set at a time |

**Mapping-level** (run once per version mapping)

| Plugin | Description |
|---|---|
| `major_version` | Creates the versioned sub-object in status (e.g. `status.v20250312`) |
| `parameters` | Injects path parameters as spec fields |
| `entry` | Populates `spec` from the chosen OpenAPI schema |
| `status` | Populates `status` from the chosen OpenAPI schema |
| `references` | Replaces scalar parameter fields with typed cross-resource references |
| `connection_secret` | Adds a `connectionSecretRef` field for injecting credentials |
| `mutual_exclusive_group` | Adds `oneOf` validation for mutually exclusive field groups (e.g. `anyOf` in OpenAPI) |
| `print_conditions` | Configures `additionalPrinterColumns` for status conditions |

**Property-level** (run per schema property during entry/status population)

| Plugin | Description |
|---|---|
| `read_only_properties` | Marks properties read-only based on OpenAPI spec |
| `read_write_properties` | Marks properties as read-write |
| `sensitive_properties` | Annotates sensitive fields; used by downstream tooling to wire Kubernetes Secrets |
| `skipped_properties` | Removes properties matched by the `skipProperties` filter |

**Extension-level**

| Plugin | Description |
|---|---|
| `atlas_sdk_version` | Embeds the Atlas SDK version annotation on the CRD |
| `reference_extensions` | Adds CRD-level annotations describing reference relationships |

## Development

```shell
make unit-test    # run tests with race detection
make ci           # fmt + tests + build (same as CI)
make gen-mock     # regenerate mocks (requires mockery)
```

## Acknowledgements

Inspired by:
- https://github.com/fybrik/openapi2crd
- https://github.com/ant31/crd-validation
- https://github.com/kubeflow/crd-validation
