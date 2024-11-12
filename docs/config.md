# Configuration

The default location for the configuration file is operating system dependent. YAML and TOML are supported.

- On macOS, it is located at `~/.config/distillery.yaml`
- On Linux, it is located at `~/.config/distillery.yaml`.
- On Windows, it is located at `%APPDATA%\distillery.yaml`

The configuration file is optional. If it is not found, the default configuration is used.

## Aliases

You can configure aliases for your installation sources. This is useful if you don't want to type the whole
path all the time.

### Simple Alias Definition

```yaml
aliases:
  dist: ekristen/distillery
  aws-nuke: ekristen/aws-nuke
  age: filosottile/age
```

OR in TOML

```toml
[aliases]
dist = "ekristen/distillery"
aws-nuke = "ekristen/aws-nuke"
age = "filosottile/age"
```

### Alias with Version

```yaml
aliases:
  dist: ekristen/distillery
  aws-nuke: ekristen/aws-nuke
  age: filosottile/age@1.0.0
```

OR in TOML

```toml
[aliases]
dist = "ekristen/distillery"
aws-nuke = "ekristen/aws-nuke"
age = "filosottile/age@1.0.0"
```

### Alias with Version as Object

```yaml
aliases:
  dist: ekristen/distillery
  aws-nuke: ekristen/aws-nuke
  age:
    name: filosottile/age
    version: 1.0.0
```

OR in TOML

```toml
[aliases]
dist = "ekristen/distillery"
aws-nuke = "ekristen/aws-nuke"

[aliases.age]
name = "filosottile/age"
version = "1.0.0"
```