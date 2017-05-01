#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin

cleanup() {
    if [ -f "${tmpfile}" ]; then
        rm -f "${tmpfile}"
    fi
}

trap "{ cleanup; }" EXIT SIGTERM

getCommits() {
    prevtag="${1}"
    tag="${2}"
    forceVer="${3:-}"
    local -a authors
    local ver="${tag}-1"
    local h

    echo "»»» Processing ${prevtag}..${tag}"
    numCommits=$(git --no-pager rev-list --count "${prevtag}".."${tag}")
    if ((numCommits>0)); then
        echo "	${numCommits} commits found"

        if [ "${tag}" == "HEAD" ]; then
            h=$(git rev-list --max-count=1 --abbrev-commit HEAD)
            if [ -z "${forceVer}" ]; then
                ver="${prevtag}~1.${h}"
            else
                ver="${forceVer}~1.${h}"
            fi

        fi

        echo "${pkgname} (${ver}) UNRELEASED; urgency=low" >> "${tmpfile}"

        authors=($(git log --format='%aN' "${prevtag}".."${tag}" | sort | uniq))
        for author in "${authors[@]}"; do
            echo "	Gathering commits from ${author}"
            {
                echo "  [ ${author} ]"
                git --no-pager log --author="${author}" --pretty=format:'  * %s' "${prevtag}".."${tag}"
                echo ""
            } >> "${tmpfile}"
        done

        git --no-pager log -n 1 --pretty='format:%n -- %aN <%aE>  %aD%n%n' "${tag}" >> "${tmpfile}"
    else
        echo "	0 commits found, skipping"
    fi
}

if [ ! -d "debian" ]; then
    echo "Directory ./debian not found"
    exit 1
fi

tmpfile=$(mktemp)

historyStart=($(git rev-list --max-parents=0 HEAD))
if ((${#historyStart[@]}>1)); then
    echo "»»» History starts with more than one commit. Using ${historyStart[-1]} as root, some changes may not get into debian/changes ."
fi
firstHash="${historyStart[-1]}"

pkgname=$(grep '^Package: ' debian/control | sed 's/^Package: //')
tags=($(git tag | sort -r -V))

if ((${#tags[@]}>0)); then
    tag=${tags[0]}
    untagged=$(git rev-list --count "${tag}"..HEAD)

    if ((untagged>0)); then
        echo "»»» Gathering untagged commits"
        getCommits "${tag}" HEAD
    fi

    for ((i=1; i<${#tags[@]}; i++)); do
        tag="${tags[${i}]}"
        nexttag="${tags[$((i-1))]}"
        getCommits "${tag}" "${nexttag}"
    done

    getCommits "${firstHash}" "${tags[-1]}"
else
    getCommits "${firstHash}" HEAD "0.0.0"
fi

mv "${tmpfile}" debian/changelog
