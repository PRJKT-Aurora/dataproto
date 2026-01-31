package dev.dataproto;

/**
 * General exception for DataProto runtime errors.
 */
public class DataProtoException extends RuntimeException {

    public DataProtoException(String message) {
        super(message);
    }

    public DataProtoException(String message, Throwable cause) {
        super(message, cause);
    }
}
