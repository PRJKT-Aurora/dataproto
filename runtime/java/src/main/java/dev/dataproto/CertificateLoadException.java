package dev.dataproto;

/**
 * Exception thrown when a certificate cannot be loaded or parsed.
 */
public class CertificateLoadException extends Exception {

    public CertificateLoadException(String message) {
        super(message);
    }

    public CertificateLoadException(String message, Throwable cause) {
        super(message, cause);
    }
}
