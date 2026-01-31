# DataProto Certification Principles

## Overview

DataProto certification ensures that implementations using the DataProto schema system uphold user-centric data principles. Certification is **required** to use DataProto in production—the runtime enforces this with hard checks.

## The Five Principles

### 1. Data Sovereignty

**User data must be stored in a user-controlled location.**

✅ **Compliant:**
- Data stored on user's device
- Data stored on user's self-hosted server
- Data stored in user's personal cloud storage (iCloud, Google Drive, etc.)
- Data stored in user-selected location with full access

❌ **Non-compliant:**
- Data stored on developer/company servers without user control
- Data stored in locations user cannot access or export from
- Data held hostage behind subscription paywalls

### 2. No Surveillance

**No behavioral tracking, profiling, or manipulation of users.**

✅ **Compliant:**
- Analytics only for debugging (crash reports, error logs)
- Anonymized, aggregated usage statistics (opt-in)
- No personal data in analytics

❌ **Non-compliant:**
- Tracking user behavior to build profiles
- Selling or sharing user data with third parties
- A/B testing that manipulates user behavior for profit
- Advertising based on user data
- Any form of dark patterns

### 3. Export/Delete

**Users must be able to export and delete their data.**

✅ **Compliant:**
- One-click export to standard formats (JSON, CSV, iCal, etc.)
- Complete data deletion on request
- No "soft delete" that keeps data hidden
- Export includes all user data, not just a subset

❌ **Non-compliant:**
- No export functionality
- Export to proprietary formats only
- Data deletion that doesn't actually delete
- Keeping "anonymized" copies after deletion request

### 4. Interoperability

**Works with the Aurora ecosystem and standard protocols.**

✅ **Compliant:**
- Uses standard data formats (protobuf, JSON)
- Implements gRPC interfaces correctly
- Can sync with other Aurora-compatible systems
- Respects schema versioning

❌ **Non-compliant:**
- Proprietary extensions that break compatibility
- Refusing to sync with other implementations
- Vendor lock-in tactics

### 5. No Extraction

**User data stays on user's device/server.**

✅ **Compliant:**
- All processing happens locally
- Sync only between user's own devices
- Optional cloud backup to user-controlled storage
- Encryption in transit and at rest

❌ **Non-compliant:**
- Uploading user data to developer servers
- "Free tier" that requires data sharing
- Analytics that include user content
- ML training on user data without explicit consent

---

## Certification Process

### Step 1: Application

Submit your application at `https://dataproto.dev/certify` with:

- Application name and identifier
- Description of how you use DataProto
- Acknowledgment of the five principles
- Contact information for certification team

### Step 2: Test Key

You'll receive a test certificate valid for development. This allows you to build and test your implementation.

### Step 3: Implementation

Build your application using the DataProto runtime. The generated code includes principle checks:

```java
// Generated repository - certification required
public CalendarEventRepository(DataProtoRuntime runtime) {
    runtime.requireCertified(); // HARD ENFORCEMENT
    this.runtime = runtime;
}
```

### Step 4: Test Suite

Run the certification test suite against your implementation:

```bash
dataprotoc certify --app your-app-identifier --tests all
```

Tests verify:
- Data export functionality works
- Data deletion is complete
- No unauthorized network requests
- Schema compliance

### Step 5: Production Certificate

After passing the test suite, you'll receive a production certificate:

```json
{
  "iss": "dataproto.aurora.dev",
  "sub": "com.example.yourapp",
  "exp": 1767225600,
  "principles": {
    "data_sovereignty": true,
    "no_surveillance": true,
    "export_delete": true,
    "interoperability": true,
    "no_extraction": true
  },
  "test_suite_version": "1.0.0"
}
```

### Step 6: Embed Certificate

Add the certificate to your application:

```java
Certificate cert = Certificate.loadFromResource("dataproto.cert");
DataProtoRuntime runtime = DataProtoRuntime.builder()
    .databasePath("app_data.db")
    .certificate(cert)
    .build();
```

---

## Annual Renewal

Certificates expire after one year. Renewal requires:

1. Re-running the test suite
2. Confirming continued compliance
3. Updating to latest DataProto version (if security updates)

---

## Revocation

Certificates may be revoked if:

- Principle violations are discovered
- App is reported and verified as non-compliant
- Security vulnerabilities are not patched

Revoked certificates are added to a public revocation list. The runtime checks this list periodically.

---

## Development Mode

For local development and testing, you can skip certification:

```java
// ONLY for development
DataProtoRuntime runtime = DataProtoRuntime.development("test.db");
```

This creates a development-only certificate that:
- Cannot be used in production builds
- Is rejected by the test suite
- Logs a warning on every use

---

## FAQ

**Q: Why is certification required?**

A: DataProto is designed for user-centric applications. The certification requirement ensures that all apps using DataProto uphold the same user protection standards. This builds trust in the ecosystem.

**Q: Is certification free?**

A: Yes, certification is free for open-source projects and individual developers. Commercial applications pay a nominal annual fee to support the certification infrastructure.

**Q: What if I disagree with a principle?**

A: The principles are non-negotiable. If your business model conflicts with user data protection, DataProto is not the right choice for your application.

**Q: Can I use DataProto without certification for internal tools?**

A: Yes, internal-only tools that never handle user data can use development mode. However, any app that processes personal data requires certification.

**Q: How do I report a violation?**

A: Email violations@dataproto.dev with evidence. All reports are investigated and kept confidential.
