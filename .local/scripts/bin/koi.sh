#! /usr/bin/env bash

installed() {
	if ! command -v "$1" &>/dev/null; then
		echo "$2" >&2
		exit 1
	fi
}

installed hurl "Missing dependency: hurl"
installed jnv "Missing dependency: jnv"

isdarwin() {
	uname | grep -i darwin
}

date="date"
if isdarwin; then
	date=gdate
fi

if ! command -v "$date" &>/dev/null; then
	echo "Please install gnu date" >&2
	exit 1
fi

[ $# -eq 1 ] && FILE="$1"
HURL_ARGS=(--color --error-format=long)
while true; do
	[ -z "$1" ] && break

	case "$1" in
		-h|--help)
			cat - <<-"EOF"
			Usage:
			koi.sh -f <file> [-x <hurl arg>]

			# Examples
			koi.sh -h                                 # show this message
			koi.sh -f create.hurl                     # run file (default hurl flags: --color --error-format=long)
			koi.sh -f create.hurl -x --ignore-asserts # final hurl args: --color --error-format=long --ignore-asserts
			koi.sh -f create.hurl -x -x               # final hurl args: --color --error-format=long -x
			koi.sh <file>                             # same as `koi.sh -f <file>`
			EOF
			exit 1
			;;

		-f)
			shift
			FILE="$1"
			;;

		-x)
			shift
			HURL_ARGS+=("$1")
	esac

	shift
done

if [ -z "$FILE" ]; then
	echo "Missing file." >&2
	exit 1
fi

DATE_FMT="+%Y-%m-%dT%H:%M:%S%z"
NOW=$("$date" -u "$DATE_FMT")

HURL_KOI_NOW=$("$date" -u -d "$NOW" "$DATE_FMT") || exit
HURL_KOI_YESTERDAY=$("$date" -u -d "$NOW -1 days" "$DATE_FMT") || exit
HURL_KOI_TOMORROW=$("$date" -u -d "$NOW +1 days" "$DATE_FMT") || exit
HURL_KOI_TODAY=$("$date" -u -d "$NOW" "$DATE_FMT") || exit
HURL_KOI_UUID=$(uuidgen) || exit
HURL_KOI_RANDOM=$RANDOM
HURL_KOI_LOREM="Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
export HURL_KOI_YESTERDAY
export HURL_KOI_NOW
export HURL_KOI_TOMORROW
export HURL_KOI_TODAY
export HURL_KOI_UUID
export HURL_KOI_RANDOM
export HURL_KOI_LOREM

output=$(hurl "${HURL_ARGS[@]}" "$FILE") || exit
if ! jnv <<< "$output"; then
	echo "$output"
fi
