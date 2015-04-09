#/bin/sh

set -ex

HELIX_VERSION=0.6.4

mkdir -p ${HELIX_INSTALL_ROOT}
if [ ! -f ${HELIX_INSTALL_ROOT}/helix-core-${HELIX_VERSION}-pkg.tar ]; then
    wget --quiet http://apache.cs.utah.edu/helix/${HELIX_VERSION}/binaries/helix-core-${HELIX_VERSION}-pkg.tar -O ${HELIX_INSTALL_ROOT}/helix-core-${HELIX_VERSION}-pkg.tar
fi

# unpack helix
mkdir -p ${HELIX_INSTALL_ROOT}/helix
tar xvf ${HELIX_INSTALL_ROOT}/helix-core-${HELIX_VERSION}-pkg.tar -C ${HELIX_INSTALL_ROOT}/helix --strip-components 1
