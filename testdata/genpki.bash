#!/usr/bin/env bash

# Generate a 2048-bit RSA private key.
openssl genrsa -out server-key.pem 2048

# Generate a CSR for a private key with servername `localhost`.
openssl req -new -sha256 -key server-key.pem -out server-csr.pem

# Create a self-signed certificate.
openssl x509 -req -in server-csr.pem -signkey server-key.pem -out server-cert.pem