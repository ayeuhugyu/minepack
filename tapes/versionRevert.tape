Output versionRevert.gif

Set Shell "bash"
Set Width 1400
Set Height 900
Set FontSize 32

Sleep 2s

Type "cd testProject" Sleep 500ms Enter
Sleep 1s

Type "# Show current version history" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version show" Sleep 500ms Enter
Sleep 3s

Type "# List current mods" Sleep 500ms Enter
Sleep 500ms
Type "./minepack list" Sleep 500ms Enter
Sleep 3s

Type "# Revert to previous version" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version revert 1.0.0" Sleep 500ms Enter
Sleep 2s

Type "# Check version after revert" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version show" Sleep 500ms Enter
Sleep 3s

Type "# List mods after revert" Sleep 500ms Enter
Sleep 500ms
Type "./minepack list" Sleep 500ms Enter
Sleep 5s
