#!/bin/sh

export SERVER_SSL_KEY=${SERVER_SSL_KEY:-ssl/server.key}
export SERVER_SSL_CRT=${SERVER_SSL_CRT:-ssl/server.crt}
SERVER_SSL_CSR=ssl/service.csr
SERVER_SSL_CA_KEY=ssl/service_ca.key
SERVER_SSL_CA_CRT=ssl/service_ca.crt

if [ -f "$SERVER_SSL_KEY" ] && [ -f "$SERVER_SSL_CRT" ]; then
    echo "service ssl crt and key already exist"
elif [ "$SERVER_USE_SSL" = "true" ]; then
    echo "Gen service ssl key and crt"
    openssl genrsa -out ${SERVER_SSL_CA_KEY} 4096
    openssl req \
        -new -x509 -days 3650 \
        -key ${SERVER_SSL_CA_KEY} \
        -subj "/C=US/ST=NY/L=NY/O=PentAGI/OU=Project/CN=PentAGI CA" \
        -out ${SERVER_SSL_CA_CRT}
    openssl req \
        -newkey rsa:4096 \
        -sha256 \
        -nodes \
        -keyout ${SERVER_SSL_KEY} \
        -subj "/C=US/ST=NY/L=NY/O=PentAGI/OU=Project/CN=localhost" \
        -out ${SERVER_SSL_CSR}

    echo "subjectAltName=DNS:pentagi.local" > extfile.tmp
    echo "keyUsage=critical,digitalSignature,keyAgreement" >> extfile.tmp

    openssl x509 -req \
        -days 730 \
        -extfile extfile.tmp \
        -in ${SERVER_SSL_CSR} \
        -CA ${SERVER_SSL_CA_CRT} -CAkey ${SERVER_SSL_CA_KEY} -CAcreateserial \
        -out ${SERVER_SSL_CRT}

    rm extfile.tmp

    cat ${SERVER_SSL_CA_CRT} >> ${SERVER_SSL_CRT}

    chmod g+r ${SERVER_SSL_KEY}
    chmod g+r ${SERVER_SSL_CA_KEY}
fi

exec "$@"
