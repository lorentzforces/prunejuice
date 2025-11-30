#!/usr/bin/env bash
set -u

# create a super-basic directory of stuff that might be useful to run the command on to test it
# "local-only-files" is gitignored to help support this

DIR_LOCATION=${1:-}

main() {
	if [[ -z "$DIR_LOCATION" ]]; then
		>&2 echo "need to be given a location for the test directory"
		return 1
	fi

	if [[ -e "$DIR_LOCATION" ]]; then
		>&2 echo "provided path $DIR_LOCATION already exists, exiting"
		return 1
	fi

	mkdir -p "$DIR_LOCATION"
	mkdir -p "$DIR_LOCATION/one_day_ago.d"
	mkdir -p "$DIR_LOCATION/one_hour_ago.d"
	mkdir -p "$DIR_LOCATION/one_day_future.d"
	touch --date='1 day ago' "$DIR_LOCATION/one_day_ago.d"
	touch --date='1 hour ago' "$DIR_LOCATION/one_hour_ago.d"
	touch --date='1 day' "$DIR_LOCATION/one_day_future.d"

	touch --date='1 day ago' "$DIR_LOCATION/one_day_ago.file"
	touch --date='2 days ago' "$DIR_LOCATION/two_day_ago.file"
	touch --date='1 hour ago' "$DIR_LOCATION/one_hour_ago.file"
	touch --date='1 hour' "$DIR_LOCATION/one_hour_future.file"
	touch --date='1 day' "$DIR_LOCATION/one_day_future.file"
}
main
