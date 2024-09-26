mkdir -p testcerts
cd testcerts
openssl genrsa -out root.key 2048
openssl req -new -x509 -days 365 -key root.key -subj "/C=CN/ST=GD/L=SZ/O=Moov, Inc./CN=Moov Root CA" -out root.crt
openssl req -newkey rsa:2048 -nodes -keyout server.key -subj "/C=CN/ST=GD/L=SZ/O=Moov, Inc./CN=localhost" -out server.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:localhost") -days 365 -in server.csr -CA root.crt -CAkey root.key -CAcreateserial -out server.crt
openssl req -newkey rsa:2048 -nodes -keyout client.key -subj "/C=CN/ST=GD/L=SZ/O=Moov, Inc./CN=moov" -out client.csr
openssl x509 -req -extfile <(printf "subjectAltName=DNS:localhost") -days 365 -in client.csr -CA root.crt -CAkey root.key -CAcreateserial -out client.crt