#!/bin/bash

tapes_path="$(dirname "$0")"

# verify that the "vhs" command is available
if ! command -v vhs &> /dev/null; then
    echo "vhs command not found. Please install vhs to proceed."
    exit 1
fi
# verify that the "minepack" command is available
if ! command -v minepack &> /dev/null; then
    echo "minepack command not found. Please install minepack to proceed."
    exit 1
fi
echo "Rendering all tapes in $tapes_path"

# init.tape:
vhs $tapes_path/init.tape

# add.tape:
vhs $tapes_path/add.tape

# list.tape:
vhs $tapes_path/list.tape

# link.tape:
# requires setup...
mkdir -p testProject/.minecraft
vhs $tapes_path/link.tape

# import.tape:
vhs $tapes_path/import.tape

# linkBisect.tape:
vhs $tapes_path/linkBisect.tape

# versionRevert.tape:
# requires setup...
cd testProject || exit 1
minepack version format increment
minepack version add --message "added create, rei, and sodium"
minepack rm rei #-y # note that -y is not actually in existence yet i just expect that it will be
minepack add emi
minepack add jei
minepack version add --message "replaced rei with emi/jei"
cd .. || exit 1
vhs $tapes_path/versionRevert.tape

# end of file
echo "Cleaning up..."
rm -rf testProject
echo "Done."