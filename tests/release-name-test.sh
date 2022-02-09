#!/bin/bash

function compare {
    echo -n "** ${3} | branchname=${1} suffix=${2} | "
    if [[ -z "$1" ]]; then
        old=$(tests/release-name.sh "$1")
        new=$(go run main.go ci release name --branchname "$1")
    else
        old=$(tests/release-name.sh "$1" "$2")
        new=$(go run main.go ci release name --branchname "$1" --release-suffix "$2")
    fi
    
    if [ "$new" = "$old" ]; then
        echo "OK"
        echo "${new}"
    else
        echo "DIFFER"
        echo "OLD: ${old}"
        echo "NEW: ${new}"
    fi
}

cd ..

# empty value test
compare "" "" "empty branchname"

# basic tests
compare "master" "" ""
compare "MASTER" "" "lower case"

# Alnum replacement test
compare "Te_3/%s^T" "" "alnum"
compare "T e s T" "" "alnum"

# release name >= 40 test 
compare "111111111122222222223333333333444444444" "" "39 chr"
compare "1111111111222222222233333333334444444444" "" "40 chr"
compare "11111111112222222222333333333344444444445" "" "41 chr"
compare "111111111122222222223333333333444444444455" "" "42 chr"
compare "11111111112222222222333333333344444444445555555555" "" "50 chr"
compare "111111111122222222223333333333444444444455555555556666666666" "" "60 chr"
compare "111111111122222222223333333333444444444455555555556666666666" "s" "60 chr + 1 chr suffix"

compare "111111111122222222223333333333444444444455555555556666666666" "aaaaaaaaaab" "60 chr + 11 chr suffix"
compare "111111111122222222223333333333444444444455555555556666666666" "aaaaaaaaaabb" "60 chr + 12 chr suffix"
compare "111111111122222222223333333333444444444455555555556666666666" "aaaaaaaaaabbb" "60 chr + 13 chr suffix"
compare "111111111122222222223333333333444444444455555555556666666666" "aaaaaaaaaabbbbbbbbbbccccccccccdddddddddd" "60 chr + 40 chr suffix"

compare "11111111112222222222" "aaaaaaaaaabbbbbbbbbbccccccccccdddddddddd" "20 chr + 40 chr suffix"
