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

```
curl -fsSL https://bun.sh/install | bash
```

Or see the [Bun install docs](https://bun.sh/docs/installation) for your platform.

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

This will attempt to detect your platform and build the correct binary in the `build/` directory.

#### d. Manual Compilation (If auto-detect fails)

- **Windows x64:**
```
bun compile:win-x64
```
- **Mac x64:**
```
bun compile:mac-x64
```
- **Mac ARM64:**
```
bun compile:mac-arm64
```
- **Linux x64:**
```
bun compile:linux-x64
```
- **Linux ARM64:**
```
bun compile:linux-arm64
```
- **All Platforms All At Once:**
```
bun compile:all
```

The compiled binaries will be in the `build/` directory.

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

## License
MIT
