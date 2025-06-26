# Minepack

A CLI tool for managing Minecraft modpacks, with extremely colorful output and exporting to modrinth's modpack format.

## Features
- **Create, search, add, remove, and list mods**
- **Dependency resolution**
- **Export to Modrinth .mrpack** format (with overrides and versioning)
- **Self-update** from GitHub releases

## Installation

### 1. Download a Release (Recommended)

Go to the [Releases page](https://github.com/ayeuhugyu/minepack/releases) and download the latest binary for your platform (e.g. `minepack-win-x64.exe`, `minepack-linux-x64`, etc). Place it somewhere in your PATH or run it directly.

### 2. Build from Source (Alternative)

#### a. Install [Bun](https://bun.sh/)
#### b. Clone and Install Dependencies

```
git clone https://github.com/ayeuhugyu/minepack.git
cd minepack
bun install
```

#### c. Compile Automatically (Recommended)

```
bun compile
```

This should be able to detect your platform and build the correct binary in the `build/` directory.
If it fails, try the most appropriate script of the build scripts in `package.json`

## Usage

You can run the CLI directly with Bun for development:

```
bun run src/index.ts <command> [...args]
```

Or use the compiled binary:

```
./build/minepack-win-x64.exe <command> [...args]
```

## Commands

| Command         | Description                                      | Example Usage / Key Flags                      |
|-----------------|--------------------------------------------------|------------------------------------------------|
| `init`          | Initialize a new minepack project                | `minepack init [path] [--force] [--verbose]`   |
| `search`        | Search for a mod on Modrinth                     | `minepack search <query> [--gameVersion <ver>] [--modloader <loader>] [--verbose]` |
| `add`           | Add a mod to your pack (with dependencies)       | `minepack add <mod> [--verbose]`               |
| `remove`        | Remove a mod, with dependency/orphan checks      | `minepack remove <mod> [--verbose]`            |
| `list`          | List all mods in your pack                       | `minepack list [full\|basic] [--hashes] [--urls] [--env] [--ids] [--size]` |
| `map`           | Show a dependency map of your pack               | `minepack map [--full] [--json] [--orphans] [--reverse] [--summary]` |
| `query`         | Check if a mod is in your pack                   | `minepack query <mod>`                         |
| `pack`          | Show pack metadata and total mod file size       | `minepack pack`                                |
| `export`        | Export as Modrinth .mrpack (with overrides)      | `minepack export [--required] [--verbose]`     |
| `selfupdate`    | Update minepack to the latest release            | `minepack selfupdate [--version <tag>]`        |
