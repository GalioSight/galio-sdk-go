#! /bin/sh
if [[ $(uname) == 'Darwin' ]]; then
    sed -i "" 's/,omitempty//g' "$1"
    sed -E -i "" 's/(json\:\"([a-z_]*).*")`/\1 yaml:"\2"`/g' "$1"
fi
if [[ $(uname) == 'Linux' ]]; then
    sed -i 's/,omitempty//g' "$1"
    sed -E -i 's/(json\:\"([a-z_]*).*")`/\1 yaml:"\2"`/g' "$1"
fi