#!/bin/bash
set -euo pipefail
IFS=$'\n\t'

PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin

# Print help
help() {
    cat << EOF
Usage: ${0##*/} [-u USER] HOST[:PORT]…
Install latest .deb on HOST logging in as USER
    -h          display help
    -u USER     optional username to connect as, you will be asked for password
                if it is required (default is root)
    HOST        one or more hostnames to work on (IP address is also fine)
    PORT        optional SSH port (default is 22)
EOF
}

# Cleanup
cleanup() {
    if [ -f "${playbook_file:-}" ]; then
        rm -f "${playbook_file}"
    fi

    if [ -f "${hosts_file:-}" ]; then
        rm -f "${hosts_file}"
    fi
}

trap "{ cleanup; }" EXIT SIGTERM

opt=""
while getopts ':hu:' opt; do
    case "${opt}" in
        h)
            help
            exit 0
            ;;
        u)
            user="${OPTARG:-root}"
            shift 2
            ;;
        ':')
            echo "-u requires username!" >&2
            help
            exit 1
            ;;
        '?')
            echo "Invalid option -${OPTARG}!" >&2
            help >&2
            exit 1
            ;;
    esac
done

if (( ${#} < 1 )); then
    help >&2
    exit 1
fi

# Find latest tag
tag=$(git tag | sort -V | tail -n1)
if [ -z "${tag}" ]; then
    echo "Can't find latest tag!"
    exit 2
fi

deb="$(pwd)/../beauties_${tag}-1_amd64.deb"
if [ ! -f "${deb}" ]; then
    echo "${deb} is not a file!" >&2
    help >&2
    exit 1
fi

# Create and handle temporary hosts file
hosts_file="$(mktemp)"

echo "[hosts]" > "${hosts_file}"

# Iterate over remaining arguments
for host; do
    echo "${host}" >> "${hosts_file}"
done

# Create and handle temporary playbook file
playbook_file="$(mktemp)"
temp_deb_filename="$(mktemp -u)"
deb_name=$(/usr/bin/basename "${deb}")

cat <<EOF > "${playbook_file}"
- hosts: all
  gather_facts: no
  tasks:
      - name: 'Copy ${deb_name} to remote host as ${temp_deb_filename}'
        copy: src='${deb}' dest='${temp_deb_filename}' mode='0600'

      - name: 'Install ${deb_name}'
        apt: deb='${temp_deb_filename}'
        become: true
      
      - name: Cleanup
        file: path='${temp_deb_filename}' state='absent'

      - name: Restart
        systemd: name=beauties state=restarted
EOF

runopts=("-u" "${user:-root}")

if [ "${user:-root}" != "root" ]; then
    runopts+=("-K")
fi
    
# Call ansible-playbook
ANSIBLE_NOCOWS=1 ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i "${hosts_file}" "${runopts[@]}" "${playbook_file}"
