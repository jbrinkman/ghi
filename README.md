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

### View Pull Request Details

The `view` subcommand retrieves and displays details of a specific pull request from a specified GitHub repository.

#### Options

- `--repo` or `-r`: The name of the GitHub repository in the format `owner/repo`. This option is required.
- `--number` or `-n`: The number of the pull request. This option is required.
- `--web` or `-w`: Open the pull request in the default web browser. This option is optional.
- `--config` or `-c`: Path to the configuration file in YAML format. This option is optional.
- `--log` or `-l`: Log that you're reviewing this pull request. This stores the review in your local database. This option is optional.

#### Example

View details of pull request #2856 from the `octocat/Hello-World` repository:

```sh
ghi pr view --repo octocat/Hello-World --number 2856
```

View details and log your review of pull request #2856:

```sh
ghi pr view --repo octocat/Hello-World --number 2856 --log
```

View details of pull request #2856 from the `octocat/Hello-World` repository in the default web browser:

```sh
ghi pr view --repo octocat/Hello-World --number 2856 --web
```

### Authentication and Database Settings

The `auth` command allows you to configure settings for the review tracking database.

#### Subcommands

##### Set Credentials

```sh
ghi auth set [flags]
```

Flags:
- `--db-url`: The Turso/LibSQL database URL.
- `--auth-token`: Authentication token for the database.
- `--username`: Your username for review tracking.

Example:
```sh
ghi auth set --db-url "libsql://your-database.turso.io" --auth-token "your-token" --username "your-github-username"
```

##### View Current Settings

```sh
ghi auth info
```

This displays your current database connection settings.

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
  - "jbrinkman"
```

## Global Flags

### Version Information
To check the version of the CLI tool:
```sh
ghi --version
```
This displays the current version, build date, and commit hash of your installation.

## Database Setup

GitHub Info CLI uses Turso/LibSQL to track your code reviews locally. This enables you to maintain a history of pull requests you've reviewed.

### Setting Up Your Turso Database

1. Create a Turso account at [turso.tech](https://turso.tech).

2. Create a database:
   ```sh
   turso db create github-info
   ```

3. Create an authentication token:
   ```sh
   turso db tokens create github-info
   ```

4. Configure GitHub Info CLI with your database details:
   ```sh
   ghi auth set --db-url "libsql://github-info-[your-username].turso.io" --auth-token "[your-token]" --username "[your-github-username]"
   ```

### Database Schema

The database automatically creates the following table:

```sql
CREATE TABLE reviews (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repo TEXT NOT NULL,
    pr_number INTEGER NOT NULL,
    reviewer TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(repo, pr_number, reviewer, timestamp)
);
```

This schema tracks:
- Repository name
- Pull request number
- Reviewer (your username)
- Timestamp of the review
