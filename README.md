# Distillery

**Status: alpha** -- things are not fully implemented, things are likely broken. Things are likely to change. Name
might change, but I like `dist` for the binary name.

## Overview

Without a doubt, [homebrew](https://brew.sh) has had a major impact on the macOS and even the linux ecosystem. It has made it easy
to install software and keep it up to date. However, it has been around for 15+ years and while it has evolved over time,
its core technology really hasn't changed, and 15+ years is an eternity in the tech world. Languages like [Go](https://golang.org)
and [Rust](https://www.rust-lang.org) have made it easy to compile binaries and distribute them without complicated
installers or dependencies. **I love homebrew**, but I think there's room for another tool.

**dist**illery is a tool that is designed to make it easy to install binaries on your system from multiple different
sources. It is designed to be simple and easy to use. It is **NOT** designed to be a package manager or handle complex
dependencies, that's where homebrew shines.

The goal of this project is to install binaries by leverage the collective power of all the developers out there that
are using tools like [goreleaser](https://goreleaser.com/) and [cargo-dist](https://github.com/axodotdev/cargo-dist)
and many others to pre-compile their software and put their binaries up on GitHub or GitLab.

## Features

- Simple to install binaries on your system from multiple sources
- No reliance on a centralized repository of metadata like package managers
- Support multiple platforms and architectures
- Support private repositories (this was a feature removed from homebrew)
- Support checksum verifications (if they exist)
- Support signatures verifications (if they exist) (**not implemented yet**)

## Install

## MacOS/Linux

1. Set your path `export PATH=$HOME/.distillery/bin:$PATH`
2. Download the latest release from the [releases page](https://github.com/ekristen/distillery/releases)
3. Extract and Run `./dist install ekristen/distillery`
4. Delete `./dist` and the .tar.gz, now use `dist` normally
5. Run `dist install owner/repo` to install a binary from GitHub Repository

## Windows

1. [Set Your Path](#set-your-path)
2. Download the latest release from the [releases page](https://github.com/ekristen/distillery/releases)
3. Extract and Run `.\dist.exe install ekristen/distillery`
4. Delete `.\dist.exe` and the .zip, now use `dist` normally
5. Run `dist install owner/repo` to install a binary from GitHub Repository

### Set Your Path

#### For Current Session

```powershell
$env:Path = "C:\Users\<username>\.distillery\bin;" + $env:Path
```

#### For Current User

```powershell
[Environment]::SetEnvironmentVariable("Path", "C:\Users\<username>\.distillery\bin;" + $env:Path, [EnvironmentVariableTarget]::User)
```

### For System

```powershell
[Environment]::SetEnvironmentVariable("Path", "C:\Users\<username>\.distillery\bin;" + $env:Path, [EnvironmentVariableTarget]::Machine)
```

### Uninstall

1. Run `dist info`
2. Remove the directories listed under the cleanup section

### Examples

Install a specific version of a tool using `@version` syntax. `github` is the default scope, this implies
`github/ekristen/aws-nuke`

```console
dist install ekristen/aws-nuke@3.16.0
```

Install a tool from a specific owner and repository, in this case hashicorp. This will install the latest version.
However, because hashicorp hosts their binaries on their own domain, distillery has special handling for obtaining
the latest version from releases.hashicorp.com instead of GitHub.

```console
dist install hashicorp/terraform
```

Install a binary from GitLab.
```console
dist install gitlab/gitlab-org/gitlab-runner
```

Often times installing from GitHub or GitLab is sufficient, but if you are on a MacOS system and Homebrew
has the binary you want, you can install it using the `homebrew` scope. I would generally still recommend just
installing from GitHub or GitLab directly.

```console
dist install homebrew/opentofu
```

## Supported Platforms

- GitHub
- GitLab
- Homebrew (binaries only, if anything has a dependency, it will not work at this time)
- Hashicorp (special handling for their releases, pointing to github repos will automatically pass through)

### Authentication

Distillery supports authentication for GitHub and GitLab. There are CLI options to pass in a token, but the preferred
method is to set the `DISTILLERY_GITHUB_TOKEN` or `DISTILLERY_GITLAB_TOKEN` environment variables using a tool like
(direnv)[https://direnv.net/].

## Behaviors

- Caching of HTTP calls where possible (GitHub primarily)
- Caching of downloads
- Allow for multiple versions of a binary using `tool@version` syntax
- Running installation for any version will automatically update the default symlink to that version (i.e. switching versions)

## Directory Structure

- Binaries
  - Symlinks `$HOME/.distillery/bin` (this should be in your `$PATH` variable)
  - Binaries `$HOME/.distillery/opt` (this is where the raw binaries are stored and symlinked to)
    - `source/owner/repo/version/<binaries>`
      - example: `github/ekristen/aws-nuke/v2.15.0/aws-nuke`
      - example: `hashicorp/terraform/v0.14.7/terraform`
- Cache directory (downloads, http caching)
  - MacOS `$HOME/Library/Caches/distillery`
  - Linux `$HOME/.cache/distillery`
  - Windows `$HOME/AppData/Local/distillery`

### Caching

At the moment there are two discrete caches. One for HTTP requests and one for downloads. The HTTP cache is used to
store the ETag and Last-Modified headers from the server to determine if the file has changed. The download cache is
used to store the downloaded file. The download cache is not used to determine if the file has changed, that is done
by the HTTP cache.
