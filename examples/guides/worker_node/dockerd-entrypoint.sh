#!/bin/sh
set -eu

# CUSTOM: Set minimum API version for Docker daemon for Portainer compatibility
export DOCKER_MIN_API_VERSION=1.41

# CUSTOM: Allow configuring Docker API host via environment variable
# Default to 0.0.0.0:2376 for backward compatibility with official image
# Usage: -e DOCKER_API_HOST=tcp://192.168.1.10:2376
if [ -z "${DOCKER_API_HOST+set}" ]; then
	DOCKER_API_HOST="tcp://0.0.0.0:2376"
fi

# CUSTOM: Allow configuring Docker API port for non-TLS mode
# Default to 0.0.0.0:2375 for backward compatibility with official image
# Usage: -e DOCKER_API_HOST_NOTLS=tcp://192.168.1.10:2375
if [ -z "${DOCKER_API_HOST_NOTLS+set}" ]; then
	DOCKER_API_HOST_NOTLS="tcp://0.0.0.0:2375"
fi

_tls_ensure_private() {
	local f="$1"; shift
	[ -s "$f" ] || openssl genrsa -out "$f" 4096
}
_tls_san() {
	{
		ip -oneline address | awk '{ gsub(/\/.+$/, "", $4); print "IP:" $4 }'
		{
			cat /etc/hostname
			echo 'docker'
			echo 'localhost'
			hostname -f
			hostname -s
		} | sed 's/^/DNS:/'
		[ -z "${DOCKER_TLS_SAN:-}" ] || echo "$DOCKER_TLS_SAN"
	} | sort -u | xargs printf '%s,' | sed "s/,\$//"
}
_tls_generate_certs() {
	local dir="$1"; shift

	# CUSTOM: Check if pre-generated certificates exist and are valid
	# If server, client, and CA certificates all exist, use them (externally managed certs)
	if [ -s "$dir/server/ca.pem" ] && [ -s "$dir/server/cert.pem" ] && [ -s "$dir/server/key.pem" ] && \
	   [ -s "$dir/client/ca.pem" ] && [ -s "$dir/client/cert.pem" ] && [ -s "$dir/client/key.pem" ] && \
	   [ -s "$dir/ca/cert.pem" ]; then
		echo "Using pre-generated certificates from $dir (externally managed)"
		openssl verify -CAfile "$dir/server/ca.pem" "$dir/server/cert.pem" || return 1
		openssl verify -CAfile "$dir/client/ca.pem" "$dir/client/cert.pem" || return 1
		return 0
	fi

	# Original behavior: if server/{ca,key,cert}.pem && !ca/key.pem, do NOTHING except verify (user likely managing CA themselves)
	# if ca/key.pem || !ca/cert.pem, generate CA public if necessary
	# if ca/key.pem, generate server public
	# if ca/key.pem, generate client public
	# (regenerating public certs every startup to account for SAN/IP changes and/or expiration)

	if [ -s "$dir/server/ca.pem" ] && [ -s "$dir/server/cert.pem" ] && [ -s "$dir/server/key.pem" ] && [ ! -s "$dir/ca/key.pem" ]; then
		echo "Using existing server certificates (external CA)"
		openssl verify -CAfile "$dir/server/ca.pem" "$dir/server/cert.pem" || return 1
		return 0
	fi

	# https://github.com/FiloSottile/mkcert/issues/174
	local certValidDays='825'

	if [ -s "$dir/ca/key.pem" ] || [ ! -s "$dir/ca/cert.pem" ]; then
		# if we either have a CA private key or do *not* have a CA public key, then we should create/manage the CA
		mkdir -p "$dir/ca"
		_tls_ensure_private "$dir/ca/key.pem"
		openssl req -new -key "$dir/ca/key.pem" \
			-out "$dir/ca/cert.pem" \
			-subj '/CN=docker:dind CA' \
			-x509 \
			-days "$certValidDays" \
			-addext keyUsage=critical,digitalSignature,keyCertSign
	fi

	if [ -s "$dir/ca/key.pem" ]; then
		# if we have a CA private key, we should create/manage a server key
		mkdir -p "$dir/server"
		_tls_ensure_private "$dir/server/key.pem"
		openssl req -new -key "$dir/server/key.pem" \
			-out "$dir/server/csr.pem" \
			-subj '/CN=docker:dind server'
		cat > "$dir/server/openssl.cnf" <<-EOF
			[ x509_exts ]
			extendedKeyUsage = serverAuth
			subjectAltName = $(_tls_san)
		EOF
		openssl x509 -req \
				-in "$dir/server/csr.pem" \
				-CA "$dir/ca/cert.pem" \
				-CAkey "$dir/ca/key.pem" \
				-CAcreateserial \
				-out "$dir/server/cert.pem" \
				-days "$certValidDays" \
				-extfile "$dir/server/openssl.cnf" \
				-extensions x509_exts
		cp "$dir/ca/cert.pem" "$dir/server/ca.pem"
		openssl verify -CAfile "$dir/server/ca.pem" "$dir/server/cert.pem"
	fi

	if [ -s "$dir/ca/key.pem" ]; then
		# if we have a CA private key, we should create/manage a client key
		mkdir -p "$dir/client"
		_tls_ensure_private "$dir/client/key.pem"
		chmod 0644 "$dir/client/key.pem" # openssl defaults to 0600 for the private key, but this one needs to be shared with arbitrary client contexts
		openssl req -new \
				-key "$dir/client/key.pem" \
				-out "$dir/client/csr.pem" \
				-subj '/CN=docker:dind client'
		cat > "$dir/client/openssl.cnf" <<-'EOF'
			[ x509_exts ]
			extendedKeyUsage = clientAuth
		EOF
		openssl x509 -req \
				-in "$dir/client/csr.pem" \
				-CA "$dir/ca/cert.pem" \
				-CAkey "$dir/ca/key.pem" \
				-CAcreateserial \
				-out "$dir/client/cert.pem" \
				-days "$certValidDays" \
				-extfile "$dir/client/openssl.cnf" \
				-extensions x509_exts
		cp "$dir/ca/cert.pem" "$dir/client/ca.pem"
		openssl verify -CAfile "$dir/client/ca.pem" "$dir/client/cert.pem"
	fi
}

# no arguments passed
# or first arg is `-f` or `--some-option`
if [ "$#" -eq 0 ] || [ "${1#-}" != "$1" ]; then
	# set "dockerSocket" to the default "--host" *unix socket* value (for both standard or rootless)
	uid="$(id -u)"
	if [ "$uid" = '0' ]; then
		dockerSocket='unix:///var/run/docker.sock'
	else
		# if we're not root, we must be trying to run rootless
		: "${XDG_RUNTIME_DIR:=/run/user/$uid}"
		dockerSocket="unix://$XDG_RUNTIME_DIR/docker.sock"
	fi
	case "${DOCKER_HOST:-}" in
		unix://*)
			dockerSocket="$DOCKER_HOST"
			;;
	esac

	# CUSTOM: Build DNS server arguments if DOCKER_DNS_SERVERS is set
	# Format: "8.8.8.8,8.8.4.4" or "192.168.1.100"
	DNS_ARGS=""
	if [ -n "${DOCKER_DNS_SERVERS:-}" ]; then
		# Split comma-separated DNS servers and add --dns for each (POSIX-compatible)
		OLD_IFS="$IFS"
		IFS=','
		for dns in $DOCKER_DNS_SERVERS; do
			# Trim whitespace
			dns=$(echo "$dns" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
			if [ -n "$dns" ]; then
				DNS_ARGS="$DNS_ARGS --dns=$dns"
			fi
		done
		IFS="$OLD_IFS"
	fi

	# add our default arguments
	if [ -n "${DOCKER_TLS_CERTDIR:-}" ]; then
		_tls_generate_certs "$DOCKER_TLS_CERTDIR"
		# generate certs and use TLS if requested/possible (default in 19.03+)
		# CUSTOM: Use DOCKER_API_HOST instead of hardcoded tcp://0.0.0.0:2376
		set -- dockerd \
			--host="$dockerSocket" \
			--host="$DOCKER_API_HOST" \
			--tlsverify \
			--tlscacert "$DOCKER_TLS_CERTDIR/server/ca.pem" \
			--tlscert "$DOCKER_TLS_CERTDIR/server/cert.pem" \
			--tlskey "$DOCKER_TLS_CERTDIR/server/key.pem" \
			$DNS_ARGS \
			"$@"
		# CUSTOM: Extract IP:PORT from DOCKER_API_HOST for rootless mode
		DOCKER_API_BIND="${DOCKER_API_HOST#tcp://}"
		DOCKER_API_PORT="${DOCKER_API_BIND##*:}"
		DOCKERD_ROOTLESS_ROOTLESSKIT_FLAGS="${DOCKERD_ROOTLESS_ROOTLESSKIT_FLAGS:-} -p ${DOCKER_API_BIND}:${DOCKER_API_PORT}/tcp"
	else
		# TLS disabled (-e DOCKER_TLS_CERTDIR='') or missing certs
		# CUSTOM: Use DOCKER_API_HOST_NOTLS instead of hardcoded tcp://0.0.0.0:2375
		set -- dockerd \
			--host="$dockerSocket" \
			--host="$DOCKER_API_HOST_NOTLS" \
			$DNS_ARGS \
			"$@"
		# CUSTOM: Extract IP:PORT from DOCKER_API_HOST_NOTLS for rootless mode
		DOCKER_API_BIND_NOTLS="${DOCKER_API_HOST_NOTLS#tcp://}"
		DOCKER_API_PORT_NOTLS="${DOCKER_API_BIND_NOTLS##*:}"
		DOCKERD_ROOTLESS_ROOTLESSKIT_FLAGS="${DOCKERD_ROOTLESS_ROOTLESSKIT_FLAGS:-} -p ${DOCKER_API_BIND_NOTLS}:${DOCKER_API_PORT_NOTLS}/tcp"
	fi
fi

if [ "$1" = 'dockerd' ]; then
	# explicitly remove Docker's default PID file to ensure that it can start properly if it was stopped uncleanly (and thus didn't clean up the PID file)
	find /run /var/run -iname 'docker*.pid' -delete || :

	# XXX inject "docker-init" (tini) as pid1 to workaround https://github.com/docker-library/docker/issues/318 (zombie container-shim processes)
	set -- docker-init -- "$@"

	iptablesLegacy=
	if [ -n "${DOCKER_IPTABLES_LEGACY+x}" ]; then
		# let users choose explicitly to legacy or not to legacy
		iptablesLegacy="$DOCKER_IPTABLES_LEGACY"
		if [ -n "$iptablesLegacy" ]; then
			modprobe ip_tables || :
			modprobe ip6_tables || :
		else
			modprobe nf_tables || :
		fi
	elif (
		# https://git.netfilter.org/iptables/tree/iptables/nft-shared.c?id=f5cf76626d95d2c491a80288bccc160c53b44e88#n420
		# https://github.com/docker-library/docker/pull/468#discussion_r1442131459
		for f in /proc/net/ip_tables_names /proc/net/ip6_tables_names /proc/net/arp_tables_names; do
			if b="$(cat "$f")" && [ -n "$b" ]; then
				exit 0
			fi
		done
		exit 1
	); then
		# if we already have any "legacy" iptables rules, we should always use legacy
		iptablesLegacy=1
	elif ! iptables -nL > /dev/null 2>&1; then
		# if iptables fails to run, chances are high the necessary kernel modules aren't loaded (perhaps the host is using xtables, for example)
		# https://github.com/docker-library/docker/issues/350
		# https://github.com/moby/moby/issues/26824
		# https://github.com/docker-library/docker/pull/437#issuecomment-1854900620
		modprobe nf_tables || :
		if ! iptables -nL > /dev/null 2>&1; then
			# might be host has no nf_tables, but Alpine is all-in now (so let's try a legacy fallback)
			modprobe ip_tables || :
			modprobe ip6_tables || :
			if /usr/local/sbin/.iptables-legacy/iptables -nL > /dev/null 2>&1; then
				iptablesLegacy=1
			fi
		fi
	fi
	if [ -n "$iptablesLegacy" ]; then
		# see https://github.com/docker-library/docker/issues/463 (and the dind Dockerfile where this directory is set up)
		export PATH="/usr/local/sbin/.iptables-legacy:$PATH"
	fi
	iptables --version # so users can see whether it's legacy or not

	uid="$(id -u)"
	if [ "$uid" != '0' ]; then
		# if we're not root, we must be trying to run rootless
		if ! command -v rootlesskit > /dev/null; then
			echo >&2 "error: attempting to run rootless dockerd but missing 'rootlesskit' (perhaps the 'docker:dind-rootless' image variant is intended?)"
			exit 1
		fi
		user="$(id -un 2>/dev/null || :)"
		if ! grep -qE "^($uid${user:+|$user}):" /etc/subuid || ! grep -qE "^($uid${user:+|$user}):" /etc/subgid; then
			echo >&2 "error: attempting to run rootless dockerd but missing necessary entries in /etc/subuid and/or /etc/subgid for $uid"
			exit 1
		fi
		: "${XDG_RUNTIME_DIR:=/run/user/$uid}"
		export XDG_RUNTIME_DIR
		if ! mkdir -p "$XDG_RUNTIME_DIR" || [ ! -w "$XDG_RUNTIME_DIR" ] || ! mkdir -p "$HOME/.local/share/docker" || [ ! -w "$HOME/.local/share/docker" ]; then
			echo >&2 "error: attempting to run rootless dockerd but need writable HOME ($HOME) and XDG_RUNTIME_DIR ($XDG_RUNTIME_DIR) for user $uid"
			exit 1
		fi
		if [ -f /proc/sys/kernel/unprivileged_userns_clone ] && unprivClone="$(cat /proc/sys/kernel/unprivileged_userns_clone)" && [ "$unprivClone" != '1' ]; then
			echo >&2 "error: attempting to run rootless dockerd but need 'kernel.unprivileged_userns_clone' (/proc/sys/kernel/unprivileged_userns_clone) set to 1"
			exit 1
		fi
		if [ -f /proc/sys/user/max_user_namespaces ] && maxUserns="$(cat /proc/sys/user/max_user_namespaces)" && [ "$maxUserns" = '0' ]; then
			echo >&2 "error: attempting to run rootless dockerd but need 'user.max_user_namespaces' (/proc/sys/user/max_user_namespaces) set to a sufficiently large value"
			exit 1
		fi
		# TODO overlay support detection?
		exec rootlesskit \
			--net="${DOCKERD_ROOTLESS_ROOTLESSKIT_NET:-vpnkit}" \
			--mtu="${DOCKERD_ROOTLESS_ROOTLESSKIT_MTU:-1500}" \
			--disable-host-loopback \
			--port-driver=builtin \
			--copy-up=/etc \
			--copy-up=/run \
			${DOCKERD_ROOTLESS_ROOTLESSKIT_FLAGS:-} \
			"$@"
	elif [ -x '/usr/local/bin/dind' ]; then
		# if we have the (mostly defunct now) Docker-in-Docker wrapper script, use it
		set -- '/usr/local/bin/dind' "$@"
	fi
else
	# if it isn't `dockerd` we're trying to run, pass it through `docker-entrypoint.sh` so it gets `DOCKER_HOST` set appropriately too
	set -- docker-entrypoint.sh "$@"
fi

exec "$@"
