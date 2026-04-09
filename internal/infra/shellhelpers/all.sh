#!/bin/bash

# MB CLI shell helpers. Loads all helpers.
# Ex.: . "$MB_HELPERS_PATH/all.sh"
# Uses MB_HELPERS_PATH (and not dirname "$0") because when sourced, $0 is the plugin script.

. "${MB_HELPERS_PATH}/os.sh"
. "${MB_HELPERS_PATH}/log.sh"
. "${MB_HELPERS_PATH}/string.sh"
. "${MB_HELPERS_PATH}/memory.sh"
. "${MB_HELPERS_PATH}/kubernetes.sh"
. "${MB_HELPERS_PATH}/flatpak.sh"
. "${MB_HELPERS_PATH}/snap.sh"
. "${MB_HELPERS_PATH}/homebrew.sh"
. "${MB_HELPERS_PATH}/github.sh"
. "${MB_HELPERS_PATH}/sudo.sh"
. "${MB_HELPERS_PATH}/shell-rc.sh"
. "${MB_HELPERS_PATH}/ensure.sh"
. "${MB_HELPERS_PATH}/context.sh"
