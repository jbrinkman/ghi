# GitHub Info CLI

`github-info` is a command-line tool to retrieve information from GitHub repositories.

## Installation

To install the `github-info` CLI, use the following command:

```sh
go install github.com/jbrinkman/ghi@latest
```

## Usage

### Retrieve Pull Requests

The `pr` command retrieves a list of pull requests from a specified GitHub repository.

#### Options

- `--repo` or `-r`: The name of the GitHub repository in the format `owner/repo`. This option is required.
- `--author` or `-A`: Filter pull requests by author. Multiple `--author` options can be used to provide a list of author filters. This option is optional.
- `--state` or `-s`: Filter pull requests by state. Valid values are `ALL`, `OPEN`, and `CLOSED`. The default value is `ALL`.
- `--reviewer` or `-R`: Filter pull requests by reviewer. Multiple `--reviewer` options can be used to provide a list of reviewer filters. This option is optional.
- `--config` or `-c`: Path to the configuration file in YAML format. This option is optional.

#### Example

Retrieve all pull requests from the `octocat/Hello-World` repository:

```sh
ghi pr --repo octocat/Hello-World
```

Retrieve pull requests from the `octocat/Hello-World` repository filtered by author `octocat`:

```sh
ghi pr --repo octocat/Hello-World --author octocat
```

Retrieve pull requests from the `octocat/Hello-World` repository filtered by multiple authors:

```sh
ghi pr --repo octocat/Hello-World --author octocat --author anotheruser
```

Retrieve pull requests from the `octocat/Hello-World` repository filtered by reviewer `octocat`:

```sh
ghi pr --repo octocat/Hello-World --reviewer octocat
```

Retrieve pull requests using a configuration file:

```sh
ghi pr --config path/to/config.yaml
```

### Configuration File

You can use a YAML configuration file to specify the options for the `pr` command. Here is an example configuration file:

```yaml
repo: "valkey-io/valkey-glide"
author:
  - "jbrinkman"
  - "Yury-Fridlyand"
  - "acarbonetto"
  - "jamesx-improving"
  - "jonathanl-bq"
  - "tjzhang-BQ"
  - "prateek-kumar-improving"
  - "cyip10"
  - "yipin-chen"
  - "edlng"
state: "open"
reviewer:
  - "reviewer1"
  - "reviewer2"
```
