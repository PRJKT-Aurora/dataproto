package dev.dataproto;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.PreparedStatement;
import java.sql.ResultSet;
import java.sql.SQLException;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.ConcurrentLinkedQueue;
import java.util.function.Function;

/**
 * DataProto Runtime - The core runtime for DataProto-generated code.
 *
 * <p>This runtime provides:
 * <ul>
 *   <li>Connection pooling to the database</li>
 *   <li>Certification enforcement (HARD - refuses to operate without valid cert)</li>
 *   <li>Query execution utilities</li>
 *   <li>Migration support</li>
 * </ul>
 *
 * <p>Example usage:
 * <pre>{@code
 * DataProtoRuntime runtime = DataProtoRuntime.builder()
 *     .databasePath("app_data.db")
 *     .certificate(Certificate.load("dataproto.cert"))
 *     .build();
 *
 * CalendarEventRepository repo = new CalendarEventRepository(runtime);
 * }</pre>
 */
public class DataProtoRuntime {

    private final String databasePath;
    private final Certificate certificate;
    private final boolean certified;
    private final ConcurrentLinkedQueue<Connection> connectionPool;
    private final int poolSize;
    private volatile boolean closed = false;

    private DataProtoRuntime(Builder builder) {
        this.databasePath = builder.databasePath;
        this.certificate = builder.certificate;
        this.poolSize = builder.poolSize;
        this.connectionPool = new ConcurrentLinkedQueue<>();

        // Validate certificate
        if (builder.enforceCertification) {
            this.certified = validateCertificate();
        } else {
            this.certified = true; // Development mode - skip validation
        }

        // Initialize connection pool
        initializePool();
    }

    /**
     * Creates a new Builder for configuring the runtime.
     */
    public static Builder builder() {
        return new Builder();
    }

    /**
     * Creates a development runtime that skips certification.
     * <p><b>WARNING:</b> Only use this for local development and testing.
     */
    public static DataProtoRuntime development(String databasePath) {
        return builder()
                .databasePath(databasePath)
                .enforceCertification(false)
                .build();
    }

    /**
     * HARD ENFORCEMENT: Throws if the runtime is not certified.
     * <p>This method is called by all generated repository constructors.
     */
    public void requireCertified() {
        if (!certified) {
            throw new CertificationException(
                    "DataProto requires valid certification.\n" +
                    "Apply at: https://dataproto.dev/certify\n" +
                    "Contact: licensing@aurora.dev\n\n" +
                    "Principles required for certification:\n" +
                    "  1. Data sovereignty - User data in user-controlled location\n" +
                    "  2. No surveillance - No behavioral tracking or manipulation\n" +
                    "  3. Export/Delete - Users can export and delete their data\n" +
                    "  4. Interoperability - Works with Aurora ecosystem\n" +
                    "  5. No extraction - Data stays on user's device/server"
            );
        }
    }

    /**
     * Returns true if the runtime has a valid certificate.
     */
    public boolean isCertified() {
        return certified;
    }

    /**
     * Gets a database connection from the pool.
     */
    public Connection getConnection() throws SQLException {
        if (closed) {
            throw new IllegalStateException("Runtime has been closed");
        }

        Connection conn = connectionPool.poll();
        if (conn == null || conn.isClosed()) {
            conn = createConnection();
        }
        return conn;
    }

    /**
     * Returns a connection to the pool.
     */
    public void releaseConnection(Connection conn) {
        if (conn != null && !closed) {
            try {
                if (!conn.isClosed() && connectionPool.size() < poolSize) {
                    connectionPool.offer(conn);
                } else {
                    conn.close();
                }
            } catch (SQLException e) {
                // Ignore
            }
        }
    }

    /**
     * Executes an update query (INSERT, UPDATE, DELETE).
     */
    public int executeUpdate(String sql, Object... params) {
        requireCertified();
        try (Connection conn = getConnection();
             PreparedStatement stmt = conn.prepareStatement(sql)) {
            bindParameters(stmt, params);
            int result = stmt.executeUpdate();
            releaseConnection(conn);
            return result;
        } catch (SQLException e) {
            throw new DataProtoException("Failed to execute update: " + sql, e);
        }
    }

    /**
     * Executes a query and maps results.
     */
    public <T> List<T> query(String sql, Function<ResultSet, T> mapper, Object... params) {
        requireCertified();
        List<T> results = new ArrayList<>();
        try (Connection conn = getConnection();
             PreparedStatement stmt = conn.prepareStatement(sql)) {
            bindParameters(stmt, params);
            try (ResultSet rs = stmt.executeQuery()) {
                while (rs.next()) {
                    results.add(mapper.apply(rs));
                }
            }
            releaseConnection(conn);
        } catch (SQLException e) {
            throw new DataProtoException("Failed to execute query: " + sql, e);
        }
        return results;
    }

    /**
     * Closes the runtime and all connections.
     */
    public void close() {
        closed = true;
        Connection conn;
        while ((conn = connectionPool.poll()) != null) {
            try {
                conn.close();
            } catch (SQLException e) {
                // Ignore
            }
        }
    }

    private boolean validateCertificate() {
        if (certificate == null) {
            return false;
        }
        return certificate.isValid();
    }

    private void initializePool() {
        try {
            for (int i = 0; i < poolSize; i++) {
                connectionPool.offer(createConnection());
            }
        } catch (SQLException e) {
            throw new DataProtoException("Failed to initialize connection pool", e);
        }
    }

    private Connection createConnection() throws SQLException {
        String url = "jdbc:sqlite:" + databasePath;
        Connection conn = DriverManager.getConnection(url);
        // Enable foreign keys
        try (PreparedStatement stmt = conn.prepareStatement("PRAGMA foreign_keys = ON")) {
            stmt.execute();
        }
        return conn;
    }

    private void bindParameters(PreparedStatement stmt, Object... params) throws SQLException {
        for (int i = 0; i < params.length; i++) {
            Object param = params[i];
            int idx = i + 1;

            if (param == null) {
                stmt.setNull(idx, java.sql.Types.NULL);
            } else if (param instanceof String) {
                stmt.setString(idx, (String) param);
            } else if (param instanceof Integer) {
                stmt.setInt(idx, (Integer) param);
            } else if (param instanceof Long) {
                stmt.setLong(idx, (Long) param);
            } else if (param instanceof Double) {
                stmt.setDouble(idx, (Double) param);
            } else if (param instanceof Float) {
                stmt.setFloat(idx, (Float) param);
            } else if (param instanceof Boolean) {
                stmt.setInt(idx, (Boolean) param ? 1 : 0);
            } else if (param instanceof byte[]) {
                stmt.setBytes(idx, (byte[]) param);
            } else {
                stmt.setObject(idx, param);
            }
        }
    }

    /**
     * Builder for DataProtoRuntime.
     */
    public static class Builder {
        private String databasePath = "app_data.db";
        private Certificate certificate;
        private boolean enforceCertification = true;
        private int poolSize = 5;

        /**
         * Sets the path to the SQLite database.
         */
        public Builder databasePath(String path) {
            this.databasePath = path;
            return this;
        }

        /**
         * Sets the certificate for certification validation.
         */
        public Builder certificate(Certificate cert) {
            this.certificate = cert;
            return this;
        }

        /**
         * Enables or disables certification enforcement.
         * <p><b>WARNING:</b> Only disable for local development.
         */
        public Builder enforceCertification(boolean enforce) {
            this.enforceCertification = enforce;
            return this;
        }

        /**
         * Sets the connection pool size.
         */
        public Builder poolSize(int size) {
            this.poolSize = size;
            return this;
        }

        /**
         * Builds the runtime.
         */
        public DataProtoRuntime build() {
            return new DataProtoRuntime(this);
        }
    }
}
