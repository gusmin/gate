#!/bin/bash

RM=$(which rm)
CHSH=$(which chsh)

# get all env var from service env file
set -o allexport
source ./config.sh
set +o allexport

echo "Changing default binary of ${GATE_USERNAME}"
sudo ${CHSH} -s /bin/bash "${GATE_USERNAME}"
sudo ${RM} -rf "${CONFIG_DEST_DIR}"
sudo ${RM} -rf "${BIN_DEST}/${BIN_NAME}"
echo "Uninstall complete"
