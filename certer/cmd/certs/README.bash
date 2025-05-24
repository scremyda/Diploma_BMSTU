# ──────────────────────────────────────────────────────────
# 1. Подготовить файлы расширений (в рабочей директории):
# ──────────────────────────────────────────────────────────

cat > v3_ca.ext << 'EOF'
[v3_ca]
basicConstraints=critical,CA:TRUE
keyUsage=critical,keyCertSign,cRLSign
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid:always,issuer
EOF

cat > v3_server.ext << 'EOF'
[ v3_server ]
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage=digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:localhost,IP:127.0.0.1
EOF

# ──────────────────────────────────────────────────────────
# 2. Сгенерировать корневой CA
# ──────────────────────────────────────────────────────────

# 2.1. Приватный ключ CA
openssl genrsa -out ca.key 4096

# 2.2. CSR для CA
openssl req -new \
  -key ca.key \
  -out ca.csr \
  -subj "/C=RU/ST=Moscow/L=Moscow/O=MyLocalCA/OU=Dev/CN=My Local Dev CA"

# 2.3. Самоподписать CSR и получить CA-сертификат с CA:TRUE
openssl x509 -req \
  -in ca.csr \
  -signkey ca.key \
  -out ca.crt \
  -days 3650 \
  -sha256 \
  -extfile v3_ca.ext \
  -extensions v3_ca

############################################
# 3. Конвертировать CA-ключ в PKCS#8 для Go
############################################

openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt \
  -in ca.key \
  -out ca_pkcs8.key

####################################################
# 4. Вставить CA-сертификат и PKCS#8-ключ в config
####################################################
cat << 'INSTR'
Откройте ваш config.yaml и добавьте в раздел Certificates -> "localhost":
  CaCert: |-
    -----BEGIN CERTIFICATE-----
    (содержимое файла ca.crt)
    -----END CERTIFICATE-----
  CaKey: |-
    -----BEGIN PRIVATE KEY-----
    (содержимое файла ca_pkcs8.key)
    -----END PRIVATE KEY-----
Не забудьте также задать Organization и ValidFor.
INSTR

## Ubuntu/Debian:
sudo cp ca.crt /usr/local/share/ca-certificates/my-local-ca-2.crt
sudo update-ca-certificates