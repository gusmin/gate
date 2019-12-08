#!/bin/bash

CP=$(which cp)
RM=$(which rm)
MV=$(which mv)
CAT=$(which cat)
TAR=$(which tar)
SED=$(which sed)
MKDIR=$(which mkdir)

SCRIPT_NAME=$(basename $0)

name="securegate-gate"
bin_dir="bin"
binary_name="securegate-gate"
release_scripts_dir="scripts"
release_config_dir="config"

releases_dir="releases"
${MKDIR} -p ${releases_dir}

function usage() {
  ${CAT} <<EOF
Usage: ${SCRIPT_NAME} version
EOF
  exit 1
}

function release() {
  release_name="${name}-v${version}"
  release_dir=${release_name}

  ${RM} -rf ${release_dir} ${releases_dir}/${release_dir}
  ${MKDIR} -p ${release_dir}

  ${CP} ${bin_dir}/${binary_name} ${release_dir}/${binary_name}
  ${CP} linux/${release_scripts_dir}/* ${release_dir}
  ${CP} linux/${release_config_dir}/* ${release_dir}

  ${TAR} cvf ${releases_dir}/${release_name}.tar ${release_dir}
  ${MV} ${release_dir} ${releases_dir}
}

[[ ${#} -lt 1 ]] && usage

version=${1}

release
