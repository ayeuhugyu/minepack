Output versionBasic.gif

Set Shell "bash"
Set Width 1200
Set Height 800
Set FontSize 32

Sleep 2s

Type "cd testProject" Sleep 500ms Enter
Sleep 1s

Type "# Set version format to semver" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version format semver" Sleep 500ms Enter
Sleep 2s

Type "# Set initial version" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version set 1.0.0 -m 'Initial release'" Sleep 500ms Enter
Sleep 2s

Type "# Increment patch version" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version patch add 1" Sleep 500ms Enter
Sleep 2s

Type "# Increment minor version" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version minor add 1" Sleep 500ms Enter
Sleep 2s

Type "# Show version history" Sleep 500ms Enter
Sleep 500ms
Type "./minepack version show" Sleep 500ms Enter
Sleep 5s
