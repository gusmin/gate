#!/bin/bash

ID=$(which id)
CP=$(which cp)
CAT=$(which cat)
TAR=$(which tar)
CHSH=$(which chsh)
CHOWN=$(which chown)
MKDIR=$(which mkdir)
USERADD=$(sudo which useradd)

function template_config_usage() {
    ${CAT} <<EOF
You need to copy ${CONFIG_FILE_NAME}.template to ${CONFIG_FILE_NAME}
and complete the configuration before launching the installation
EOF
}

# get all env var from config file
set -o allexport
source ./config.sh
set +o allexport

function check_file() {
    if [[ -x ./${BIN_NAME} ]]; then
        if [[ ! -r ./${BIN_NAME} ]]; then
            echo "Binary file not readable"
            exit 1
        fi
    else
        echo "No binary file"
        exit 1
    fi
    if [[ -f ./${CONFIG_FILE_NAME} ]]; then
        if [[ ! -r ./${CONFIG_FILE_NAME} ]]; then
            echo "Config file not readable"
            exit 1
        fi
    else
        echo "No config file"
        template_config_usage
        exit 1
    fi
}

function create_user() {
    ${ID} -u ${GATE_USERNAME} >/dev/null 2>/dev/null
    if [[ ${?} -eq 1 ]]; then
        echo "Creating user ${GATE_USERNAME}..."
        sudo ${USERADD} --user-group --create-home --comment "Secure Gate" ${GATE_USERNAME}
        echo "User created."
    else
        echo "User ${GATE_USERNAME} already exist."
    fi
}

function launch_install() {
    set -e
    echo "Copy binary"
    sudo ${CP} ./${BIN_NAME} ${BIN_DEST}
    echo "Copy config"
    sudo ${MKDIR} -p ${CONFIG_DEST_DIR}
    sudo ${CP} ${CONFIG_FILE_NAME} ${CONFIG_DEST_DIR}
    echo "Create log folder"
    sudo ${MKDIR} -p ${LOG_FOLDER}
    sudo ${CHOWN} -R ${USER}: ${LOG_FOLDER}
    echo "Create translation folder"
    sudo ${MKDIR} -p ${TRANSLATIONS_DEST}
    sudo ${TAR} -zxf ${TRANSLATIONS_TAR} -C ${TRANSLATIONS_DEST}
    echo "Change the default shell of ${GATE_USERNAME} to ${BIN_NAME}"
    sudo ${CHSH} -s ${BIN_DEST}/${BIN_NAME} ${GATE_USERNAME}
    set +e
}

function main() {
    echo "Checking files"
    check_file
    echo "Checking User"
    create_user
    echo "Installing"
    launch_install
}

main "$@"
