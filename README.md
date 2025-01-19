# GitHub Info CLI

`github-info` is a command-line tool to retrieve information from GitHub repositories.

## Installation

To install the `github-info` CLI, use the following command:

```sh
go install github.com/jbrinkman/github-info@latest
```

## Usage

### Retrieve Pull Requests

The `pullrequest` command retrieves a list of pull requests from a specified GitHub repository.

#### Options

- `--repo` or `-r`: The name of the GitHub repository in the format

owner/repo

. This option is required.

- `--author` or `-a`: Filter pull requests by author. Multiple `--author` options can be used to provide a list of author filters. This option is optional.

#### Example

Retrieve all pull requests from the `octocat/Hello-World` repository:

```sh
github-info pullrequest --repo octocat/Hello-World
```

Retrieve pull requests from the `octocat/Hello-World` repository filtered by author `octocat`:

```sh
github-info pullrequest --repo octocat/Hello-World --author octocat
```

Retrieve pull requests from the `octocat/Hello-World` repository filtered by multiple authors:

```sh
github-info pullrequest --repo octocat/Hello-World --author octocat --author anotheruser
```

### Output

The `pullrequest` command outputs the following fields for each pull request:

- Number
- Title
- Author
- State
- URL

Example output:

```
Pull requests for octocat/Hello-World
=====================================
#1 - Fix all the bugs
Author: octocat
State: open
URL: https://github.com/octocat/Hello-World/pull/1
=====================================
#2 - Add new feature
Author: anotheruser
State: closed
URL: https://github.com/octocat/Hello-World/pull/2
=====================================
```

```
