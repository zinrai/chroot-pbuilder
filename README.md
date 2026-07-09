# chroot-pbuilder

chroot-pbuilder is a tool designed to simplify the creation, update, and login process for Debian GNU/Linux chroot environments using pbuilder. It streamlines the management of isolated Debian-based development and testing environments.

## Features

- Easy creation and management of chroot environments using pbuilder
- Support for multiple distributions and architectures
- Custom role definition capability
- Support for additional pbuilder options

## Note

- The base.tgz file is created in the current directory.
- The bindmounts directory is created at `$HOME/.chroot-pbuilder/`.

## Prerequisites

- `sudo` command
- `pbuilder` command
- sudo privileges

Please ensure that both `sudo` and `pbuilder` are installed and available in your system's PATH. chroot-pbuilder will check for these commands at startup and exit with an error message if they are not found.

## Installation

Build the program:

```
$ go build
```

## Usage

chroot-pbuilder provides three main commands: `create`, `update`, and `login`. Each command requires the `--distribution` flag, and accepts optional `--architecture` and `--role` flags.

### Creating a chroot environment

```
$ chroot-pbuilder create --distribution sid --architecture amd64
```

### Updating a chroot environment

```
$ chroot-pbuilder update --distribution bullseye --architecture arm64
```

### Logging into a chroot environment

```
$ chroot-pbuilder login --distribution sid --architecture amd64 -- --save-after-login
```

## Options

- `--distribution`: (Required) The distribution to use
- `--architecture`: (Optional, default: amd64) The architecture to use
- `--role`: (Optional) Custom role name
- `--force`: (For create command only) Overwrite existing base.tgz

Additional pbuilder options can be specified after `--`.

## License

This project is licensed under the [MIT License](./LICENSE).
