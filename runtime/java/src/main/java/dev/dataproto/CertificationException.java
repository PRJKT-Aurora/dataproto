package dev.dataproto;

/**
 * Exception thrown when certification validation fails.
 *
 * <p>This exception is thrown by the DataProto runtime when:
 * <ul>
 *   <li>No certificate is provided</li>
 *   <li>The certificate signature is invalid</li>
 *   <li>The certificate has expired</li>
 *   <li>Required principles are not agreed to</li>
 * </ul>
 *
 * <p>To resolve this error:
 * <ol>
 *   <li>Apply for certification at https://dataproto.dev/certify</li>
 *   <li>Agree to the DataProto principles</li>
 *   <li>Pass the certification test suite</li>
 *   <li>Embed the issued certificate in your application</li>
 * </ol>
 */
public class CertificationException extends RuntimeException {

    public CertificationException(String message) {
        super(message);
    }

    public CertificationException(String message, Throwable cause) {
        super(message, cause);
    }
}
