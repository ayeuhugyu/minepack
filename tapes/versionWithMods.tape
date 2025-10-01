Output versionWithMods.gif

Set Shell "bash"
Set Width 1400
Set Height 900
Set FontSize 32

Sleep 2s

Type "cd testProject" Sleep 500ms Enter
Sleep 1s

Type "# Initialize version tracking" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version format semver" Sleep 500ms Enter
Sleep 1s
Type "./minepack version set 1.0.0 -m 'Initial version'" Sleep 500ms Enter
Sleep 2s

Type "# Add some mods" Sleep 500ms Enter
Sleep 500ms
Type "./minepack add sodium" Sleep 500ms Enter
Sleep 3s
Enter
Sleep 2s

Type "./minepack add lithium" Sleep 500ms Enter
Sleep 3s
Enter
Sleep 2s

Type "# Increment version after adding mods" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version patch add 1" Sleep 500ms Enter
Sleep 2s

Type "# List mods and show version" Sleep 500ms Enter
Sleep 500ms
Type "./minepack list" Sleep 500ms Enter
Sleep 3s

Type "./minepack version show" Sleep 500ms Enter
Sleep 5s
