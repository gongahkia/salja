[![](https://img.shields.io/badge/salja_1.0.0-passing-green)](https://github.com/gongahkia/salja/releases/tag/1.0.0)
![](https://github.com/gongahkia/salja/actions/workflows/ci.yml/badge.svg)

# `Salja`

[All-in-one](https://www.collinsdictionary.com/us/dictionary/english/all-in-one#:~:text=All-in-one%20means%20having,blend%20of%20stocks%20and%20bonds.) converter between [calendar and task management apps](#support) with [native cloud sync](#architecture) and [conflict detection](#architecture), served as a [CLI tool](#architecture) *(now also available as an [**MCP server**](#mcp-server))*.

## Stack

* *Scripting*: [Go](https://go.dev/), [Cobra](https://github.com/spf13/cobra), [go-keyring](https://github.com/zalando/go-keyring) 
* *TUI*: [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), [Bubbles](https://github.com/charmbracelet/bubbles)
* *MCP*: [mcp-go](https://github.com/mark3labs/mcp-go)
* *Parsing*: [go-ical](https://github.com/emersion/go-ical) 
* *Config files*: [TOML](https://github.com/BurntSushi/toml) 
* *Lint*: [golangci-lint](https://golangci-lint.run/)
* *Build*: [GoReleaser](https://goreleaser.com/), [Docker](https://www.docker.com/)

## Support

| Format | Extension | Events | Tasks | Recurrence | Subtasks |
|---|---|---|---|---|---|
| **ICS** | `.ics` | yes | yes | yes | no |
| **Google Calendar** | `.csv` | yes | no | no | no |
| **Outlook** | `.csv` | yes | no | no | no |
| **Todoist** | `.csv` | no | yes | no | yes |
| **TickTick** | `.csv` | no | yes | yes | yes |
| **Notion** | `.csv` | no | yes | no | no |
| **Asana** | `.csv` | no | yes | no | no |
| **Trello** | `.json` | no | yes | no | yes |
| **OmniFocus** | `.taskpaper` | no | yes | no | yes |
| **Apple Calendar** | native | yes | no | no | no |
| **Apple Reminders** | native | no | yes | no | no |

## What `Salja` can do *([at the moment](https://github.com/gongahkia/salja/issues))*

1. **Cloud Sync (OAuth)**: Push/pull to Google Calendar, Microsoft Outlook, Todoist, TickTick, and Notion via authenticated API calls with PKCE OAuth2 flow, token refresh, and secure keyring storage.
2. **Conflict Detection**: Fuzzy duplicate detection using UID matching, Levenshtein title distance, and date proximity heuristics. Configurable resolution strategies: `ask`, `prefer-source`, `prefer-target`, `skip-conflicts`, `fail-on-conflict`.
3. **Fidelity Checking**: Pre-conversion warnings when the target format can't represent source data (subtasks, recurrence rules, reminders, timezones). Modes: `warn` (default), `error`, `silent`.
4. **Streaming CSV/ICS parsing**: 
5. **Locale-aware date parsing**: 
6. **Shell completion**: 
7. **Native AppleScript**:

## Usage

1. First run any of the below commands to get `Salja` on your local machine.
    1. Homebrew

    ```console
    $ brew install gongahkia/salja/salja
    ```

    2. Go install

    ```console
    $ go install github.com/gongahkia/salja/cmd/salja@latest
    ```

    3. Build from source

    ```console
    $ git clone https://github.com/gongahkia/salja.git
    $ cd salja && make install
    ```

    4. Nix build

    ```console
    $ nix build .#salja
    ```

    5. Docker

    ```console
    $ docker build -t salja .
    $ docker run --rm salja convert input.ics output.csv --to gcal
    $ docker compose run --rm salja convert input.ics output.csv --to gcal
    $ docker compose run --rm salja-mcp
    ```

    6. Arch linux

    ```console
    # arch linux
    $ makepkg -si
    ```

2. Then execute the below commands for calendar/task conversion.

```console
$ salja convert calendar.ics tasks.csv --to todoist # auto-detect formats from file extensions
$ salja convert data.csv output.ics --from gcal --to ics # explicit formats
$ salja convert input.ics output.csv --to gcal --dry-run # dry-run preview
$ salja convert input.ics output.csv --to todoist --fidelity error # strict mode (fail on data loss)
$ salja convert new.ics existing.ics --merge # merge with conflict detection
$ salja convert tasks.ics output --to apple-calendar --calendar "Work" # apple calendar (macOS)
```

4. Alernatively use the below commands for cloud sync.

```console
$ salja auth login google # authenticate with google
$ salja auth login notion # authenticate with notion

$ salja sync push calendar.ics --to google # push local file to google cloud
$ salja sync push tasks.csv --to todoist --dry-run # push local file to todoist cloud

$ salja sync pull --from google --output calendar.ics # pull from google cloud to local file
$ salja sync pull --from todoist --output tasks.csv --start 2026-01-01 --end 2026-06-01 # pull from todoist cloud

$ salja auth status # check auth status
```

5. Additionally run any of the below commands.

```console
$ salja list-formats # list supported formats

$ salja validate calendar.ics # validate a file

$ salja diff old.ics new.ics --format table # diff two files

$ salja config init # config initialisation
$ salja config path # config path

$ source <(salja completion bash) # shell completion for bash
$ salja completion zsh > ~/.zfunc/_salja # shell completion for zsh
```

6. Finally run `Salja`'s TUI with the following command.

```console
$ salja # launches TUI
$ salja tui # does the same thing as above
```

## MCP Server

`salja-mcp` is `Salja`'s [MCP server](https://modelcontextprotocol.io/) that exposes `Salja`'s capabilities to your AI agent via *stdio*.

1. Run the below to install `salja-mcp`.

```console
$ go install github.com/gongahkia/salja/cmd/salja-mcp@latest # install via go install
$ docker run --rm -i --entrypoint salja-mcp salja # install via docker
```

2. Your AI agent then has access to the following skills.

| Tool | Description |
|---|---|
| `convert` | Convert between calendar/task formats |
| `validate` | Validate a file and return format, item count, field coverage |
| `diff` | Compare two calendar/task files |
| `list_formats` | List all supported formats with capabilities |
| `sync_push` | Push local data to a cloud service |
| `sync_pull` | Pull data from a cloud service |
| `auth_status` | Check authentication status for all services |

### Resources

| URI | Description |
|---|---|
| `salja://formats` | All registered format metadata |
| `salja://config` | Current configuration values |
| `salja://auth/{service}` | Auth state for a specific service |

## Architecture

<div align="center">
    <img src="./asset/reference/architecture.png">
</div>

## Reference

The name `Salja` is in reference to...