# git-synchronizer

[![build](https://github.com/insightsengineering/git-synchronizer/actions/workflows/test.yml/badge.svg)](https://github.com/insightsengineering/git-synchronizer/actions/workflows/test.yml)

## Installing

Simply download the project for your distribution from the [releases](https://github.com/insightsengineering/git-synchronizer/releases) page. `git-synchronizer` is distributed as a single binary file and does not require any additional system requirements.

## Usage

`git-synchronizer` is a command line utility, so after installing the binary in your `PATH`, simply run the following command to view its capabilities:

```bash
git-synchronizer --help
```

## Configuration file

If you'd like to set the options in a configuration file, by default `git-synchronizer` checks `~/.git-synchronizer`, `~/.git-synchronizer.yaml` and `~/.git-synchronizer.yml` files.
If any of these files exist, `git-synchronizer` uses options defined there, unless they are overridden by command line flags.

You can also specify custom path to configuration file with `--config <your-configuration-file>.yml` command line flag.

Example contents of configuration file:

```yaml
# Default authentication methods to use for source and destination repositories (optional).
defaults:
  source:
    auth:
      method: token
      # Name of environment variable storing the token.
      token_name: GITHUB_TOKEN
  destination:
    auth:
      method: token
      # Name of environment variable storing the token.
      token_name: GITLAB_TOKEN

# List of repository pairs to be synchronized.
repositories:

  # Repositories using default tokens.
  - source:
      repo: https://github.example.com/org-1/repo-1
    destination:
      repo: https://gitlab.example.com/org-5/repo-1

  - source:
      repo: https://github.example.com/org-1/repo-2
      # Overriding token for source repository.
      auth:
        method: token
        token_name: GITHUB_TOKEN_EXTRA
    destination:
      repo: https://gitlab.example.com/org-5/repo-2

  - source:
      repo: https://github.example.com/org-1/repo-3
    destination:
      repo: https://gitlab.example.com/org-5/repo-3
      # Overriding token for destination repository.
      auth:
        method: token
        token_name: GITLAB_TOKEN_EXTRA
```

## Environment variables

`git-synchronizer` reads environment variables with `GITSYNCHRONIZER_` prefix and tries to match them with CLI flags.
For example, setting the following variables will override the respective values from configuration file:
`GITSYNCHRONIZER_LOGLEVEL` etc.

The order of precedence is:

CLI flag → environment variable → configuration file → default value.

To check the available names of environment variables, please run `git-synchronizer --help`.

Please note that providing the list of repositories to be synchronized with a CLI flag is not supported.

## Development

This project is built with the [Go programming language](https://go.dev/).

### Development Environment

It is recommended to use Go 1.21+ for developing this project. This project uses a pre-commit configuration and it is recommended to [install and use pre-commit](https://pre-commit.com/#install) when you are developing this project.

### Common Commands

Run `make help` to list all related targets that will aid local development.

## License

`git-synchronizer` is licensed under the Apache 2.0 license. See [LICENSE](LICENSE) for details.
