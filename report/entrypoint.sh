#!/usr/bin/env bash
# use supplied year and week or assign defaults
NEBULA_REPORT_YEAR="${NEBULA_REPORT_YEAR:=$(date +"%Y")}"
NEBULA_REPORT_WEEK="${NEBULA_REPORT_WEEK:=$(($(date +"%W")-1))}"
NETWORK_NAME="${NETWORK_NAME:-ipfs}"

REPORT_DIR_BASE="${REPORT_DIR_BASE:-reports}"
REPORT_DIR="${REPORT_DIR:-$REPORT_DIR_BASE/$NEBULA_REPORT_YEAR/calendar-week-$NEBULA_REPORT_WEEK/$NETWORK_NAME}"

echo "Starting nebula report production for the $NETWORK_NAME network, $NEBULA_REPORT_YEAR, week $NEBULA_REPORT_WEEK"

WORK_DIR="$(mktemp -d)"

# Generate report
poetry run python main.py "$WORK_DIR"
if [[ $? != 0 ]]
then
 	echo "Could not generate the report"
	exit 1
fi

echo "Copying report files to $REPORT_DIR"

mkdir -p "$REPORT_DIR"
cp -r "$WORK_DIR"/* "$REPORT_DIR"
