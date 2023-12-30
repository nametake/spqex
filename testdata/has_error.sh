#!/bin/bash

read -r input

result=$(echo "$input" | xargs echo -n)

if echo "$result" | grep -q "HAS_ERROR"; then
	echo -n "COMMAND ERROR" 1>&2
	exit 1
else
	echo -n "${result//TABLE/TABLE_A}"
fi
