# tfc

A CLI for HCP Terraform (formerly Terraform Cloud). Browse workspaces, trigger runs, manage variable sets, and provision no-code modules from the private registry.

## Install

```
brew tap justinclayton/tap
brew install tfc
```

Or build from source:

```
git clone https://github.com/justinclayton/tfc-cli.git
cd tfc-cli
make build
```

## Prerequisites

Authenticate with HCP Terraform using the standard Terraform CLI:

```
terraform login
```

This stores a token at `~/.terraform.d/credentials.tfrc.json`, which `tfc` reads automatically.

## Configuration

Set your default organization (required) and optionally a default project:

```
tfc config init
```

Or set values directly:

```
tfc config set default_org my-org
tfc config set default_project my-project
```

Config is stored at `~/.config/tfc/config.yaml`. All defaults can be overridden per-command with `--org`, `--project`, and `--hostname` flags.

## Usage

### Workspaces

```
tfc workspace list                              # list all workspaces
tfc workspace list --search prod                # filter by name
tfc workspace list --status errored             # filter by run status
tfc workspace list --tags networking,prod        # filter by tags
tfc workspace show my-workspace                 # workspace details
tfc workspace run my-workspace                  # trigger a run
tfc workspace run my-workspace -m "deploy v2"   # trigger with message
tfc workspace runs my-workspace                 # list recent runs
tfc workspace runs my-workspace -n 5            # last 5 runs
tfc workspace show-run my-workspace             # show latest run details + errors
tfc workspace show-run my-workspace run-abc123  # show specific run
tfc workspace destroy my-workspace              # queue a destroy run
tfc workspace delete my-workspace               # permanently delete
```

The `show-run` command automatically detects errored runs and displays parsed Terraform diagnostics with source locations and error details.

### Modules

```
tfc module list                                             # list private registry modules
tfc module show my-module --provider aws                    # module details + versions
tfc module provision my-module --provider aws               # no-code provision (interactive)
tfc module provision my-module --provider aws \
  --var region=us-east-1 --var instance_type=t3.micro \
  --name my-new-workspace                                   # non-interactive
```

### Variable Sets

```
tfc varset list                                     # list variable sets
tfc varset show my-varset                           # show variables in a set
tfc varset var create my-varset --key FOO --value bar --category env
tfc varset var update my-varset FOO --value newval
tfc varset var delete my-varset FOO
```

Variable sets can be referenced by name or ID (`varset-xxxxx`).

## Output Modes

`tfc` detects whether stdout is a terminal:

- **Interactive** (TTY): colored, aligned table output
- **Piped**: tab-separated values with no headers, designed for `cut`, `awk`, `grep`
- **`--json`**: full structured JSON output

```
tfc workspace list                          # pretty tables
tfc workspace list | grep prod              # plain TSV, pipe-friendly
tfc workspace list --json | jq '.[].NAME'   # JSON mode
```

## Global Flags

```
--org string        override default organization
--project string    override default project
--hostname string   override HCP Terraform hostname
--json              output as JSON
--no-color          disable colored output
```

## License

MIT
