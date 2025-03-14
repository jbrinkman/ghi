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
- `--debug` or `-d`: Enable debug logging to a file. Logs will be saved in `~/.ghi/logs/` directory with date-based rotation. This option is optional.

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

Enable debug logging while retrieving pull requests:

```sh
ghi pr --repo octocat/Hello-World --debug
```

### View Pull Request Details

The `view` subcommand retrieves and displays details of a specific pull request from a specified GitHub repository.

#### Options

- `--repo` or `-r`: The name of the GitHub repository in the format `owner/repo`. This option is required.
- `--number` or `-n`: The number of the pull request. This option is required.
- `--web` or `-w`: Open the pull request in the default web browser. This option is optional.
- `--config` or `-c`: Path to the configuration file in YAML format. This option is optional.
- `--log` or `-l`: Log that you're reviewing this pull request. This stores the review in your local database. This option is optional.
- `--debug` or `-d`: Enable debug logging to a file. Logs will be saved in `~/.ghi/logs/` directory with date-based rotation. This option is optional.

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

Enable debug logging while viewing pull request details:

```sh
ghi pr view --repo octocat/Hello-World --number 2856 --debug
```

### Review History

The `review` subcommand displays a list of pull requests you've reviewed within a specified date range. This data is pulled from your local database where reviews are logged when using the `--log` flag with the view command.

#### Options

- `--repo` or `-r`: Filter reviews by repository in the format `owner/repo`. This option is optional.
- `--start-date` or `-s`: The start date for the review search in YYYY-MM-DD format. If not provided, defaults to 30 days ago.
- `--end-date` or `-e`: The end date for the review search in YYYY-MM-DD format. If not provided, defaults to today.
- `--debug` or `-d`: Enable debug logging to a file. Logs will be saved in `~/.ghi/logs/` directory with date-based rotation. This option is optional.

#### Example

Display all reviews from the last 30 days:
```sh
ghi pr review
```

Display reviews between specific dates:
```sh
ghi pr review --start-date 2023-01-01 --end-date 2023-12-31
```

Display reviews for a specific repository between specific dates:
```sh
ghi pr review --repo octocat/Hello-World --start-date 2023-01-01 --end-date 2023-12-31
```

Enable debug logging while displaying review history:

```sh
ghi pr review --debug
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
- `--debug` or `-d`: Enable debug logging to a file. Logs will be saved in `~/.ghi/logs/` directory with date-based rotation. This option is optional.

Example:
```sh
ghi auth set --db-url "libsql://your-database.turso.io" --auth-token "your-token" --username "your-github-username"
```

##### View Current Settings

```sh
ghi auth info
```

This displays your current database connection settings.

Enable debug logging while viewing current settings:

```sh
ghi auth info --debug
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
  - "jbrinkman"
```

## Global Flags

### Debug Mode
The `--debug` or `-d` flag is available on all commands and enables detailed logging to a file:

```sh
ghi pr --repo octocat/Hello-World --debug
ghi pr view --repo octocat/Hello-World --number 2856 --debug
ghi pr review --debug
ghi auth info --debug
```

When debug mode is enabled, detailed logs are written to files in the `~/.ghi/logs/` directory. Logs are automatically rotated daily with the naming format `ghi-YYYY-MM-DD.log`.

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

## Debugging

When using the `--debug` flag with any command, detailed logs are written to files in the `~/.ghi/logs/` directory. Logs are automatically rotated daily and named in the format `ghi-YYYY-MM-DD.log`. 

Log files contain detailed information about:
- Command execution and parameters
- API requests and responses (summaries only, not full content)
- Processing steps and their outcomes
- Database operations
- Error details

Only the most recent logs are kept; older logs are automatically compressed and eventually deleted based on these settings:
- Maximum log file size: 10 MB per file
- Maximum number of backup files: 5
- Maximum age of log files: 30 days

You can use these logs for troubleshooting issues with the CLI or to better understand the GitHub API responses.
