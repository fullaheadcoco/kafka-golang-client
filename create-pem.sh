#!/bin/bash

set -o nounset \
    -o errexit

printf "Deleting previous (if any)..."
rm -rf secrets
rm -rf tmp
mkdir secrets
mkdir -p tmp
echo " OK!"

# Generate CA key and certificate
printf "Creating CA..."
openssl req -new -x509 -keyout tmp/datahub-ca.key -out tmp/datahub-ca.crt -days 365 -subj '/CN=ca.datahub/OU=test/O=datahub/L=paris/C=fr' -passin pass:datahub -passout pass:datahub >/dev/null 2>&1
echo " OK!"

for i in 'kafka-1' 'kafka-2' 'kafka-3' 'kafka-ui' 'localhost' # 'schema-registry'
do
	printf "Creating cert, keystore, and PEM files for $i..."
	
	# Generate a private key for each service in PEM format
	openssl genpkey -algorithm RSA -out secrets/$i.key.pem -pass pass:datahub >/dev/null 2>&1

	# Create a certificate signing request (CSR)
	openssl req -new -key secrets/$i.key.pem -out tmp/$i.csr -subj "/CN=$i/OU=test/O=datahub/L=paris/C=fr" -passin pass:datahub >/dev/null 2>&1

	# Sign the CSR with the CA to produce a certificate
	openssl x509 -req -in tmp/$i.csr -CA tmp/datahub-ca.crt -CAkey tmp/datahub-ca.key -CAcreateserial -out secrets/$i.crt.pem -days 365 -passin pass:datahub >/dev/null 2>&1

	# Convert the PEM key to PKCS12 format, required for Java KeyStore
	openssl pkcs12 -export -in secrets/$i.crt.pem -inkey secrets/$i.key.pem -out tmp/$i.p12 -name $i -CAfile tmp/datahub-ca.crt -caname root -password pass:datahub >/dev/null 2>&1

	# Create the Java KeyStore (.jks) from the PKCS12 file
	keytool -importkeystore -deststorepass datahub -destkeypass datahub -destkeystore secrets/$i.keystore.jks -srckeystore tmp/$i.p12 -srcstoretype PKCS12 -srcstorepass datahub -alias $i >/dev/null 2>&1

	# Import the CA certificate into the truststore
	keytool -keystore secrets/$i.truststore.jks -alias CARoot -import -noprompt -file tmp/datahub-ca.crt -storepass datahub >/dev/null 2>&1
	
	echo " OK!"
done

# Save credentials to a file
echo "datahub" > secrets/cert_creds

# Cleanup temporary files if needed
# rm -rf tmp

echo "SUCCEEDED"