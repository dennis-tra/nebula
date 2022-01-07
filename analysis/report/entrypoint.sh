#!/usr/bin/env bash

# params:
#  <IPFS_API> . i.e. /ip4/x.x.x.x/tcp/5001
# this can be passed as an argument or environment variable.
#
# 1. Generate reports
# 2. Pin to IPFS
# 3. Copy to MFS
# 4. Pubish to IPNS

if [[ ! $# -ne 1  ]]
then
	IPFS=$1
elif [[ ! -z "${IPFS_API}" ]]
then
	IPFS="${IPFS_API}"
else
	echo "Sorry, I don't know where to upload reports. This is a configuration error. Will not continue." >&2
	exit 1
fi


# this is the directory where reports will be generated.
mkdir reports

# Generate report
poetry run python ./gen_report.py || echo "Error generating reports. I'll publish what I got." && echo "could not generate the report today" > reports/report-error

REPORTDIR=nimbus-$(date +"%y-%m-%d")
mv reports "${REPORTDIR}"

# Pin to IPFS
REPORTCID=$(ipfs --api="${IPFS}" add -Qr "${REPORTDIR}")
echo "pinned report with CID ${REPORTCID}"

# Copy to MFS
ipfs --api="${IPFS}" files cp -p "/ipfs/${REPORTCID}" "/nimbus-reports/${REPORTDIR}"
MFSHASH=$(ipfs --api="${IPFS}" files stat --hash "/nimbus-reports")

# Publish to IPNS
ipfs --api="${IPFS}" name publish "${MFSHASH}"
