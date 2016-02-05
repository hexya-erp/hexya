#!/usr/bin/env bash

YEP_SRC=github.com/npiganeau/yep

function find_path {
    for path in ${GOPATH//:/ }; do
        if [ -d "$path/src/$1" ]; then
            FOUND_PATH="$path/src/$1";
            return 0;
        fi
    done
    FOUND_PATH="";
}

# First let's find the YEP directory
echo "Starting YEP build..."
find_path ${YEP_SRC}
if [ -z "$FOUND_PATH" ]; then
    echo "YEP directory not found";
    exit 1;
else
    echo "YEP directory found at $FOUND_PATH"
fi

YEP_PATH=${FOUND_PATH}

# Remove modules
echo "Removing previously installed modules"
rm -rf "$YEP_PATH/modules"

# Copy all addons into modules
echo "Copying addons to the modules directory"
while read addon_path; do
#    find_path ${addon_path}
#    if [ -z "$FOUND_PATH" ]; then
#        echo "- Addon directory $addon_path not found, skipping...";
#    else
    echo "- $addon_path"
    package=${addon_path##*/}
    ${GOPATH}/bin/gomvpkg -from "$addon_path" -to "$YEP_SRC/modules/$package" -vcs_mv_cmd "cp -a {{.Src}} {{.Dst}}"
#    fi
done <"$YEP_PATH/modules.list"

exit 0
