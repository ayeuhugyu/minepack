# minepack

A modern Minecraft modpack manager and CLI tool that feels similar to real package managers.

## Features
- **Cross-platform CLI** for managing Minecraft modpacks
- Supports mods, resourcepacks, shaderpacks, datapacks, plugins, and more
- Modrinth API integration for searching, adding, and updating content
- Content-type-aware folder management
- Export to Modrinth `.mrpack` format
- Import from Packwiz projects

## Installation

### Download prebuilt binaries
- Visit the [Releases page](https://github.com/ayeuhugyu/minepack/releases) to download the latest version for your OS.
- Place the binary somewhere in your PATH (e.g. `/usr/local/bin`).

### Build from source
1. [Install Bun](https://bun.sh/)
2. Clone this repo:
   ```sh
   git clone https://github.com/ayeuhugyu/minepack.git
   cd minepack
   ```
3. Compile:
   ```sh
   bun compile
   ```
4. The binary will be in `build/`.

## Usage

Run `minepack help` for full command/flag info.

### Commands

- `minepack init` — Initialize a new modpack in the current directory. Use `--force` to re-initialize.
- `minepack add <mod>` — Add a mod/content by Modrinth URL, ID, or search term. Supports direct URLs and all content types. Use `--download` to fetch the file directly.
- `minepack list` — List all content in the pack. Supports filtering by type, side, etc.
- `minepack query <mod>` — Query if a mod/content exists in the pack (by name, filename, or URL).
- `minepack remove <mod>` — Remove a mod/content from the pack (by name, filename, or URL).
- `minepack search <term>` — Search Modrinth for mods/content.
- `minepack update` — Update all mods/content to the latest version for the current Minecraft version and modloader. Prompts if a mod can't be updated.
- `minepack export modrinth` — Export the pack to Modrinth `.mrpack` format. Supports filtering by side and downloading all files into overrides.
- `minepack import packwiz [--input DIR] [--output DIR]` — Import a Packwiz project into minepack format. Converts all supported content and copies overrides.
- `minepack help` — Show help for all commands and flags.

### Content Types Supported
- mods
- resourcepacks
- shaderpacks
- datapacks
- plugins

### How it works
- Each content type is stored in its own folder (e.g. `mods/`, `resourcepacks/`, etc.)
- Each mod/content is represented by a `.json` stub (or direct file) with all metadata
- All actions are logged verbosely for transparency
- Exported packs are Modrinth-compliant and ready for upload

### Import/Export
- **Import:** Converts Packwiz projects (including all supported content types and overrides) to minepack format
- **Export:** Creates a Modrinth `.mrpack` with correct metadata, file sizes, and loader mapping

## Adding Configs or Other Local Files

The .mrpack format has a system for adding local files, Minepack will take any files you put anywhere inside of it that it doesn't recognize and replicate them to the overrides folder.

## Contributing
PRs and issues welcome! See the code for details.

## License
MIT