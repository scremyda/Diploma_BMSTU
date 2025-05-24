### Команда создать серты с неправильным CN
openssl req -x509 -nodes -newkey rsa:2048 \
-keyout localhost.key -out localhost.crt \
-days 365 -subj "/CN=bad.cn" \
-addext "subjectAltName = IP:88.88.88.88"
