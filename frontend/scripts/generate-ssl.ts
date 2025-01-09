import { execSync } from 'node:child_process';
import { chmodSync, existsSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';

interface SSLPaths {
    sslDir: string;
    serverKey: string;
    serverCert: string;
    serverCsr: string;
    caKey: string;
    caCert: string;
}

const SSL_PATHS: SSLPaths = {
    sslDir: join(process.cwd(), 'ssl'),
    serverKey: join(process.cwd(), 'ssl', 'server.key'),
    serverCert: join(process.cwd(), 'ssl', 'server.crt'),
    serverCsr: join(process.cwd(), 'ssl', 'server.csr'),
    caKey: join(process.cwd(), 'ssl', 'ca.key'),
    caCert: join(process.cwd(), 'ssl', 'ca.crt'),
};

const executeCommand = (command: string): void => {
    try {
        execSync(command, { stdio: 'inherit' });
    } catch (error) {
        console.error(`Error executing command: ${command}`);
        throw error;
    }
};

export const generateCertificates = (): void => {
    // Create ssl directory if it doesn't exist
    if (!existsSync(SSL_PATHS.sslDir)) {
        mkdirSync(SSL_PATHS.sslDir, { recursive: true });
    }

    // Check if certificates already exist
    if (existsSync(SSL_PATHS.serverKey) && existsSync(SSL_PATHS.serverCert)) {
        console.log('SSL certificates already exist');
        return;
    }

    console.log('Generating SSL certificates...');

    // Generate CA key
    executeCommand(`openssl genrsa -out ${SSL_PATHS.caKey} 4096`);

    // Generate CA certificate
    executeCommand(
        `openssl req -new -x509 -days 3650 -key ${SSL_PATHS.caKey} \
    -subj "/C=US/ST=NY/L=NY/O=PentAGI/OU=Project/CN=PentAGI CA" \
    -out ${SSL_PATHS.caCert}`,
    );

    // Generate server key and CSR
    executeCommand(
        `openssl req -newkey rsa:4096 -sha256 -nodes \
    -keyout ${SSL_PATHS.serverKey} \
    -subj "/C=US/ST=NY/L=NY/O=PentAGI/OU=Project/CN=localhost" \
    -out ${SSL_PATHS.serverCsr}`,
    );

    // Create temporary configuration file
    const extFile = join(SSL_PATHS.sslDir, 'extfile.tmp');
    const extFileContent = ['subjectAltName=DNS:pentagi.local', 'keyUsage=critical,digitalSignature,keyAgreement'].join(
        '\n',
    );

    executeCommand(`echo "${extFileContent}" > ${extFile}`);

    // Sign the certificate
    executeCommand(
        `openssl x509 -req -days 730 \
    -extfile ${extFile} \
    -in ${SSL_PATHS.serverCsr} \
    -CA ${SSL_PATHS.caCert} \
    -CAkey ${SSL_PATHS.caKey} \
    -CAcreateserial \
    -out ${SSL_PATHS.serverCert}`,
    );

    // Append CA certificate to server certificate
    executeCommand(`cat ${SSL_PATHS.caCert} >> ${SSL_PATHS.serverCert}`);

    // Set group read permissions
    chmodSync(SSL_PATHS.serverKey, '0640');
    chmodSync(SSL_PATHS.caKey, '0640');

    // Remove temporary files
    executeCommand(`rm ${extFile}`);

    console.log('SSL certificates generated successfully');
};
