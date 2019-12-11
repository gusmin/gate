#!/bin/bash

if [[ ${BASH_SOURCE} = $0 ]]; then
  echo "This script should only be sourced"
  exit 1
fi

# User
GATE_USERNAME="secure"

# Config
CONFIG_DEST_DIR="/etc/securegate/gate"
CONFIG_FILE_NAME="config.json"

# Translations
TRANSLATIONS_TAR="translations.tgz"
TRANSLATIONS_DEST="/var/lib/securegate/gate/"

# Log
LOG_FOLDER

# Binary
BIN_NAME="securegate-gate"
BIN_DEST="/usr/bin"
