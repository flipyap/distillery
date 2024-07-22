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

## Install

1. Set your path `export PATH=$HOME/.distillery/bin:$PATH`
2. Download the latest release from the [releases page](https://github.com/ekristen/distillery/releases)
3. Extract and Run `./dist install ekristen/distillery`
4. Delete `./dist` and the .tar.gz
5. Run `dist install owner/repo` to install a binary from GitHub Repository

## Goals

- Make it simple to install binaries on your system from multiple sources
- Do not rely on a centralized repository of metadata like package managers
- Support binary verifications and signatures if they exist
- Support multiple platforms and architectures

## Supported Platforms

- GitHub
- GitLab
- Homebrew (binaries only)
- Hashicorp

## Needed Before 1.0

- [ ] implement signature verification
- [ ] implement hash verification
- [ ] implement uninstall
- [ ] implement cleanup
