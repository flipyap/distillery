# Distillery

**Status: alpha** -- things are not fully implemented, things are likely broken. Things are likely to change. Name
might change, but I like `dist` for the binary name.

## Overview

Without a doubt, [homebrew](https://brew.sh) has had a major impact on the macOS ecosystem. It has made it easy to 
install software and keep it up to date. It has been around for 15 years and while it has evolved over time, its core
technology hasn't changed, and 15 year is an eternity in the tech world. I love homebrew, but I think there's room for
another tool.

The goal of this project is to leverage the collective power of all the developers out there that are using tools like
[goreleaser](https://goreleaser.com/) and [cargo-dist](https://github.com/axodotdev/cargo-dist) and many others to 
pre-compile their software and put their binaries up on GitHub or GitLab and install the binaries.

## Needed Before 1.0

- [ ] make defaults configurable (bin directory for example)
- [ ] finish bin re-org under `opt/`
- [ ] detect already installed versions
- [ ] implement signature verification
- [ ] implement uninstall
- [ ] implement cleanup

## Install

1. Set your path `export PATH=$HOME/.distillery/bin:$PATH`
2. Download the latest release from the [releases page](https://github.com/ekristen/distillery/releases)
3. Extract and Run `./dist install ekristen/distillery`
4. Delete `./dist` and the .tar.gz, now use `dist` normally
5. Run `dist install owner/repo` to install a binary from GitHub Repository

### Uninstall

1. Simply remove `$HOME/.distillery/bin` from your path
2. Remove `$HOME/.distillery` directory
3. Optionally remove cache directory (varies by OS, viewable by the `info` command)
4. Done

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
installing from GitHub.

```console
dist install homebrew/opentofu
```

## Goals

- Make it simple to install binaries on your system from multiple sources
- Do not rely on a centralized repository of metadata like package managers
- Support binary verifications and signatures if they exist
- Support multiple platforms and architectures

## Supported Platforms

- GitHub
- GitLab
- Homebrew (binaries only, if anything has a dependency, it will not work at this time)
- Hashicorp

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
