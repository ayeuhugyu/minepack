export const projectReadmeContent = `# Minepack Project Guide

First and foremost, DO NOT EDIT FILES IN THE ROOT OR STUBS DIRECTORY OF YOUR PROJECT MANUALLY. Any file ending with .mp.json is a Minepack file and manually editing them can cause issues if improperly formatted.
Use minepack commands to manage your project files.

If you wish to change the modloader, game version, or any other properties of your project, you can run the \`minepack init\` command again with the \`--force\` flag. This will overwrite the existing pack file with the new properties you provide.

To add mods to your project, use \`minepack add\` (see help for more details). This command will add stubs for whatever mod you specify.

To add custom files to your project, place them in the \`overrides\` directory. Any files placed in this directory will be copied to the modpack root when exported.

For example, to add configs for a mod, create a directory in \`overrides\` named \`config\` and place the config files in there.
Custom .jar mod files can be added in a similar vain, by placing them in a folder called \`mods\` in the \`overrides\` directory. 
Once again, any and all files or directories in the \`overrides\` folder will just be exactly copied to the modpack root when exported. Anything. At all.
`