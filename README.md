# vtp

Vault Transpose (vtp) is a tool for interpolating vault secrets into any file
regardless of its type. It will loop through a file and replace any instances
of `((path:key))` with the value of the key defined at the path in vault.

vtp is meant for replacing key/value secrets and will not handle properly
indenting multi-line secrets (such as certs) into YAML files. It will simply
replace any tokens it finds with the values it fetches from vault and is
not file-type dependent.

## Token

vtp will try to replace anything within `(( ))` blocks within your code. The
proper syntax for vault secrets is: `(( /path/to/secret:key ))`. If no `key`
is provided then vtp will default to using `value` as the key.

## Usage

```
Usage: vtp [--inplace] [--debug] FILES [FILES ...]

Positional arguments:
  FILES                  List of files to run against

Options:
  --inplace, -i          Overwrite input files
  --quiet, -q            Suppress output (useful when writing files in place)
  --debug, -d            Enable debug logging [default: false]
  --help, -h             display this help and exit
```

The `VAULT_` params needed to make a call to the CLI must be set to be able
to interpolate values with vtp. At a minimum you will need to set 
`VAULT_ADDR` and `VAULT_TOKEN`.

By default vtp will output to the screen, but you can suppress the output with
the `--quiet` flag. This is most useful when paired with the `--inplace` flag
when using with fluxcd and kustomize.