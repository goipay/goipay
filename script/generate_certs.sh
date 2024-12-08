#!/bin/bash

pathToCertFolder="./cert"
pathToCertServerFolder="$pathToCertFolder/server"
pathToCertClientFolder="$pathToCertFolder/client"

## Server
mkdir -p "$pathToCertServerFolder"

# Create a Private Key for the Server CA
openssl genrsa -out "$pathToCertServerFolder/ca.key" 4096
 
# Create a Self-Signed Certificate for the Server CA
openssl req -x509 -new -nodes -key "$pathToCertServerFolder/ca.key" -sha256 -subj "/C=US/ST=New York/L=New York City/O=Goipay CA Inc./CN=Goipay Server Root CA" -days 36500 -out "$pathToCertServerFolder/ca.crt"

# Create a Private Key for the Server
openssl genrsa -out "$pathToCertServerFolder/server.key" 4096
 
# Create a Certificate Signing Request (CSR) config for the Server
cat > "$pathToCertServerFolder/server.ext" <<EOF
[ req ]
default_bits       = 2048
default_keyfile    = server.key
default_md         = sha256
prompt             = no
distinguished_name = req_distinguished_name
x509_extensions    = v3_req
 
[ req_distinguished_name ]
C                  = US
ST                 = Test Cape
L                  = Test Town
O                  = Goipay
OU                 = Finance
CN                 = goipay.github.io
 
[ v3_req ]
keyUsage           = keyEncipherment, dataEncipherment
extendedKeyUsage   = serverAuth
subjectAltName     = @alt_names
 
[ alt_names ]
DNS.1              = localhost
DNS.1              = backend-processor
IP.1               = 127.0.0.1
IP.2               = 0.0.0.0
EOF

# Create a Certificate Signing Request (CSR) for the Server
openssl req -new -key "$pathToCertServerFolder/server.key" -out "$pathToCertServerFolder/server.csr" -config "$pathToCertServerFolder/server.ext"
 
# Sign the Server CSR with the CA Certificate to Generate the Server Certificate
openssl x509 -req -in "$pathToCertServerFolder/server.csr" -CA "$pathToCertServerFolder/ca.crt" -CAkey "$pathToCertServerFolder/ca.key" -CAcreateserial -out "$pathToCertServerFolder/server.crt" -days 36500 -sha256 -extensions v3_req -extfile "$pathToCertServerFolder/server.ext"


## Client 
mkdir -p "$pathToCertClientFolder"

# Create a Private Key for the Client CA
openssl genrsa -out "$pathToCertClientFolder/ca.key" 4096
 
# Create a Self-Signed Certificate for the Client CA
openssl req -x509 -new -nodes -key "$pathToCertClientFolder/ca.key" -sha256 -subj "/C=US/ST=New York/L=New York City/O=Goipay CA Inc./CN=Goipay Client Root CA" -days 36500 -out "$pathToCertClientFolder/ca.crt"

# Create a Private Key for the Client
openssl genrsa -out "$pathToCertClientFolder/client.key" 4096
 
# Create a Certificate Signing Request (CSR) config for the Client
cat > "$pathToCertClientFolder/client.ext" <<EOF
[ req ]
default_bits       = 2048
default_keyfile    = client.key
default_md         = sha256
prompt             = no
distinguished_name = req_distinguished_name
x509_extensions    = v3_req
 
[ req_distinguished_name ]
C                  = US
ST                 = Test Cape
L                  = Test Town
O                  = Goipay
OU                 = Finance
CN                 = goipay.github.io
 
[ v3_req ]
keyUsage           = keyEncipherment, dataEncipherment
extendedKeyUsage   = clientAuth
subjectAltName     = @alt_names
 
[ alt_names ]
DNS.1              = localhost
DNS.1              = backend-processor
IP.1               = 127.0.0.1
IP.2               = 0.0.0.0
EOF

# Create a Certificate Signing Request (CSR) for the Client
openssl req -new -key "$pathToCertClientFolder/client.key" -out "$pathToCertClientFolder/client.csr" -config "$pathToCertClientFolder/client.ext"
 
# Sign the Client CSR with the CA Certificate to Generate the Client Certificate
openssl x509 -req -in "$pathToCertClientFolder/client.csr" -CA "$pathToCertClientFolder/ca.crt" -CAkey "$pathToCertClientFolder/ca.key" -CAcreateserial -out "$pathToCertClientFolder/client.crt" -days 36500 -sha256 -extensions v3_req -extfile "$pathToCertClientFolder/client.ext"
