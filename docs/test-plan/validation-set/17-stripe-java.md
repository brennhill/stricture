# 17 — Stripe Payments API (Java)

**Why included:** Java records, `@JsonProperty`, `HttpClient` (Java 11+), JUnit 5 `@ParameterizedTest`, `assertAll`, `Optional<T>`, sealed interfaces, pattern matching, modern exception handling.

**Language:** Java 17+
**Framework:** JUnit 5, java.net.http.HttpClient
**Domain:** Payment processing
**API:** Stripe Payments API (Charges, Customers)

---

## Manifest Fragment

```yaml
- id: stripe-java
  language: java
  framework: junit5
  domain: payments
  api_style: rest
  endpoints:
    - POST /v1/charges
    - GET /v1/charges/:id
    - GET /v1/charges (pagination)
    - POST /v1/customers
  features:
    - java_records
    - sealed_interfaces
    - pattern_matching
    - optional_handling
    - http_client
    - json_annotation
    - parametrized_tests
    - nested_test_classes
```

---

## PERFECT Implementation

```java
// StripeClient.java
package com.example.stripe;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.PropertyNamingStrategies;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.security.InvalidKeyException;
import java.security.NoSuchAlgorithmException;
import java.time.Duration;
import java.time.Instant;
import java.util.*;
import java.util.stream.Stream;
import java.util.stream.StreamSupport;

/**
 * Stripe API client using modern Java features.
 * Demonstrates records, sealed interfaces, Optional, and HttpClient.
 */
public final class StripeClient {
    private static final String BASE_URL = "https://api.stripe.com/v1";
    private static final Duration TIMEOUT = Duration.ofSeconds(30);
    private static final int MIN_AMOUNT = 50; // $0.50 minimum
    private static final int MAX_AMOUNT = 99999999; // $999,999.99 maximum

    private final HttpClient httpClient;
    private final String apiKey;
    private final ObjectMapper objectMapper;

    private StripeClient(Builder builder) {
        this.apiKey = Objects.requireNonNull(builder.apiKey, "API key required");
        this.httpClient = builder.httpClient != null
            ? builder.httpClient
            : HttpClient.newBuilder()
                .connectTimeout(TIMEOUT)
                .build();
        this.objectMapper = new ObjectMapper()
            .setPropertyNamingStrategy(PropertyNamingStrategies.SNAKE_CASE);
    }

    public static Builder builder() {
        return new Builder();
    }

    public static final class Builder {
        private String apiKey;
        private HttpClient httpClient;

        public Builder apiKey(String apiKey) {
            this.apiKey = apiKey;
            return this;
        }

        public Builder httpClient(HttpClient httpClient) {
            this.httpClient = httpClient;
            return this;
        }

        public StripeClient build() {
            return new StripeClient(this);
        }
    }

    // Domain models using records

    public record Charge(
        @JsonProperty("id") String id,
        @JsonProperty("object") String object,
        @JsonProperty("amount") int amount,
        @JsonProperty("amount_captured") int amountCaptured,
        @JsonProperty("amount_refunded") int amountRefunded,
        @JsonProperty("balance_transaction") String balanceTransaction,
        @JsonProperty("currency") String currency,
        @JsonProperty("customer") String customer,
        @JsonProperty("description") Optional<String> description,
        @JsonProperty("status") ChargeStatus status,
        @JsonProperty("failure_code") Optional<String> failureCode,
        @JsonProperty("failure_message") Optional<String> failureMessage,
        @JsonProperty("paid") boolean paid,
        @JsonProperty("refunded") boolean refunded,
        @JsonProperty("created") long created
    ) {
        // Compact constructor with validation
        public Charge {
            Objects.requireNonNull(id, "id required");
            Objects.requireNonNull(object, "object required");
            Objects.requireNonNull(currency, "currency required");
            Objects.requireNonNull(status, "status required");

            if (amount < MIN_AMOUNT || amount > MAX_AMOUNT) {
                throw new IllegalArgumentException(
                    "Amount must be between " + MIN_AMOUNT + " and " + MAX_AMOUNT
                );
            }

            if (!id.startsWith("ch_")) {
                throw new IllegalArgumentException("Invalid charge ID format");
            }

            if (currency.length() != 3) {
                throw new IllegalArgumentException("Currency must be 3-letter ISO code");
            }
        }

        public Instant getCreatedInstant() {
            return Instant.ofEpochSecond(created);
        }
    }

    public record Customer(
        @JsonProperty("id") String id,
        @JsonProperty("object") String object,
        @JsonProperty("email") Optional<String> email,
        @JsonProperty("name") Optional<String> name,
        @JsonProperty("description") Optional<String> description,
        @JsonProperty("balance") int balance,
        @JsonProperty("created") long created
    ) {
        public Customer {
            Objects.requireNonNull(id, "id required");
            Objects.requireNonNull(object, "object required");

            if (!id.startsWith("cus_")) {
                throw new IllegalArgumentException("Invalid customer ID format");
            }

            // Validate email format if present
            email.ifPresent(e -> {
                if (!e.contains("@") || !e.contains(".")) {
                    throw new IllegalArgumentException("Invalid email format");
                }
            });
        }
    }

    public record ChargeList(
        @JsonProperty("object") String object,
        @JsonProperty("data") List<Charge> data,
        @JsonProperty("has_more") boolean hasMore,
        @JsonProperty("url") String url
    ) {}

    public enum ChargeStatus {
        @JsonProperty("succeeded") SUCCEEDED,
        @JsonProperty("pending") PENDING,
        @JsonProperty("failed") FAILED
    }

    // Sealed interface for response handling
    public sealed interface StripeResponse<T> permits StripeResponse.Success, StripeResponse.Error {
        record Success<T>(T data) implements StripeResponse<T> {}
        record Error<T>(int statusCode, String message, Optional<String> code) implements StripeResponse<T> {}
    }

    public static final class StripeException extends Exception {
        private final int statusCode;
        private final Optional<String> code;

        public StripeException(int statusCode, String message, Optional<String> code) {
            super(message);
            this.statusCode = statusCode;
            this.code = code;
        }

        public int getStatusCode() {
            return statusCode;
        }

        public Optional<String> getCode() {
            return code;
        }
    }

    /**
     * Creates a charge with idempotency key support.
     */
    public StripeResponse<Charge> createCharge(
        int amount,
        String currency,
        String source,
        Optional<String> customerId,
        Optional<String> description,
        Optional<String> idempotencyKey
    ) throws IOException, InterruptedException {

        if (amount < MIN_AMOUNT || amount > MAX_AMOUNT) {
            return new StripeResponse.Error<>(400,
                "Amount must be between " + MIN_AMOUNT + " and " + MAX_AMOUNT,
                Optional.of("invalid_amount"));
        }

        var bodyParams = new HashMap<String, String>();
        bodyParams.put("amount", String.valueOf(amount));
        bodyParams.put("currency", currency);
        bodyParams.put("source", source);
        customerId.ifPresent(id -> bodyParams.put("customer", id));
        description.ifPresent(desc -> bodyParams.put("description", desc));

        var requestBuilder = HttpRequest.newBuilder()
            .uri(URI.create(BASE_URL + "/charges"))
            .timeout(TIMEOUT)
            .header("Authorization", "Bearer " + apiKey)
            .header("Content-Type", "application/x-www-form-urlencoded")
            .POST(buildFormBody(bodyParams));

        idempotencyKey.ifPresent(key ->
            requestBuilder.header("Idempotency-Key", key)
        );

        var request = requestBuilder.build();

        try {
            var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                var charge = objectMapper.readValue(response.body(), Charge.class);
                return new StripeResponse.Success<>(charge);
            } else {
                var errorBody = parseErrorResponse(response.body());
                return new StripeResponse.Error<>(
                    response.statusCode(),
                    errorBody.message(),
                    errorBody.code()
                );
            }
        } catch (IOException | InterruptedException e) {
            throw e;
        }
    }

    /**
     * Retrieves a charge by ID.
     */
    public StripeResponse<Charge> getCharge(String chargeId) throws IOException, InterruptedException {
        if (!chargeId.startsWith("ch_")) {
            return new StripeResponse.Error<>(400,
                "Invalid charge ID format",
                Optional.of("invalid_charge_id"));
        }

        var request = HttpRequest.newBuilder()
            .uri(URI.create(BASE_URL + "/charges/" + chargeId))
            .timeout(TIMEOUT)
            .header("Authorization", "Bearer " + apiKey)
            .GET()
            .build();

        try {
            var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                var charge = objectMapper.readValue(response.body(), Charge.class);
                return new StripeResponse.Success<>(charge);
            } else {
                var errorBody = parseErrorResponse(response.body());
                return new StripeResponse.Error<>(
                    response.statusCode(),
                    errorBody.message(),
                    errorBody.code()
                );
            }
        } catch (IOException | InterruptedException e) {
            throw e;
        }
    }

    /**
     * Lists all charges with pagination support.
     * Returns a Stream that automatically handles pagination.
     */
    public Stream<Charge> listCharges(int limit) throws IOException, InterruptedException {
        return StreamSupport.stream(
            Spliterators.spliteratorUnknownSize(
                new ChargeIterator(limit),
                Spliterator.ORDERED
            ),
            false
        );
    }

    private final class ChargeIterator implements Iterator<Charge> {
        private final int pageSize;
        private List<Charge> currentPage = new ArrayList<>();
        private int currentIndex = 0;
        private String startingAfter = null;
        private boolean hasMore = true;

        ChargeIterator(int pageSize) {
            this.pageSize = pageSize;
            fetchNextPage();
        }

        @Override
        public boolean hasNext() {
            if (currentIndex < currentPage.size()) {
                return true;
            }

            if (hasMore) {
                fetchNextPage();
                return currentIndex < currentPage.size();
            }

            return false;
        }

        @Override
        public Charge next() {
            if (!hasNext()) {
                throw new NoSuchElementException();
            }
            return currentPage.get(currentIndex++);
        }

        private void fetchNextPage() {
            try {
                var uriBuilder = new StringBuilder(BASE_URL + "/charges?limit=" + pageSize);
                if (startingAfter != null) {
                    uriBuilder.append("&starting_after=").append(startingAfter);
                }

                var request = HttpRequest.newBuilder()
                    .uri(URI.create(uriBuilder.toString()))
                    .timeout(TIMEOUT)
                    .header("Authorization", "Bearer " + apiKey)
                    .GET()
                    .build();

                var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

                if (response.statusCode() >= 200 && response.statusCode() < 300) {
                    var chargeList = objectMapper.readValue(response.body(), ChargeList.class);
                    currentPage = chargeList.data();
                    currentIndex = 0;
                    hasMore = chargeList.hasMore();

                    if (!currentPage.isEmpty()) {
                        startingAfter = currentPage.get(currentPage.size() - 1).id();
                    }
                } else {
                    hasMore = false;
                }
            } catch (IOException | InterruptedException e) {
                hasMore = false;
                throw new RuntimeException("Failed to fetch charges", e);
            }
        }
    }

    /**
     * Creates a customer.
     */
    public StripeResponse<Customer> createCustomer(
        Optional<String> email,
        Optional<String> name,
        Optional<String> description
    ) throws IOException, InterruptedException {

        var bodyParams = new HashMap<String, String>();
        email.ifPresent(e -> bodyParams.put("email", e));
        name.ifPresent(n -> bodyParams.put("name", n));
        description.ifPresent(d -> bodyParams.put("description", d));

        var request = HttpRequest.newBuilder()
            .uri(URI.create(BASE_URL + "/customers"))
            .timeout(TIMEOUT)
            .header("Authorization", "Bearer " + apiKey)
            .header("Content-Type", "application/x-www-form-urlencoded")
            .POST(buildFormBody(bodyParams))
            .build();

        try {
            var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                var customer = objectMapper.readValue(response.body(), Customer.class);
                return new StripeResponse.Success<>(customer);
            } else {
                var errorBody = parseErrorResponse(response.body());
                return new StripeResponse.Error<>(
                    response.statusCode(),
                    errorBody.message(),
                    errorBody.code()
                );
            }
        } catch (IOException | InterruptedException e) {
            throw e;
        }
    }

    /**
     * Verifies Stripe webhook signature using HMAC SHA-256.
     */
    public boolean verifyWebhookSignature(
        String payload,
        String signatureHeader,
        String webhookSecret
    ) {
        try {
            // Parse signature header: t=timestamp,v1=signature
            var parts = signatureHeader.split(",");
            String timestamp = null;
            String signature = null;

            for (var part : parts) {
                var keyValue = part.split("=", 2);
                if (keyValue.length == 2) {
                    if (keyValue[0].equals("t")) {
                        timestamp = keyValue[1];
                    } else if (keyValue[0].equals("v1")) {
                        signature = keyValue[1];
                    }
                }
            }

            if (timestamp == null || signature == null) {
                return false;
            }

            // Construct signed payload
            var signedPayload = timestamp + "." + payload;

            // Compute HMAC
            var mac = Mac.getInstance("HmacSHA256");
            var secretKey = new SecretKeySpec(
                webhookSecret.getBytes(StandardCharsets.UTF_8),
                "HmacSHA256"
            );
            mac.init(secretKey);
            var hash = mac.doFinal(signedPayload.getBytes(StandardCharsets.UTF_8));
            var computedSignature = bytesToHex(hash);

            return computedSignature.equals(signature);

        } catch (NoSuchAlgorithmException | InvalidKeyException e) {
            return false;
        }
    }

    private String bytesToHex(byte[] bytes) {
        var hexChars = new char[bytes.length * 2];
        for (int i = 0; i < bytes.length; i++) {
            int v = bytes[i] & 0xFF;
            hexChars[i * 2] = Character.forDigit(v >>> 4, 16);
            hexChars[i * 2 + 1] = Character.forDigit(v & 0x0F, 16);
        }
        return new String(hexChars);
    }

    private HttpRequest.BodyPublisher buildFormBody(Map<String, String> params) {
        var body = params.entrySet().stream()
            .map(entry -> entry.getKey() + "=" + entry.getValue())
            .reduce((a, b) -> a + "&" + b)
            .orElse("");
        return HttpRequest.BodyPublishers.ofString(body);
    }

    private record ErrorResponse(String message, Optional<String> code) {}

    private ErrorResponse parseErrorResponse(String body) {
        try {
            var tree = objectMapper.readTree(body);
            var error = tree.get("error");
            if (error != null) {
                var message = error.has("message")
                    ? error.get("message").asText()
                    : "Unknown error";
                var code = error.has("code")
                    ? Optional.of(error.get("code").asText())
                    : Optional.<String>empty();
                return new ErrorResponse(message, code);
            }
            return new ErrorResponse("Unknown error", Optional.empty());
        } catch (IOException e) {
            return new ErrorResponse("Failed to parse error response", Optional.empty());
        }
    }
}
```

---

## Bug Cases (B01-B15)

### B01 — No Error Handling (TQ-error-path-coverage)

**Bug:** Missing try/catch on HttpClient.send(), allowing IOException/InterruptedException to propagate uncaught
**Expected violation:** `TQ-error-path-coverage`

```java
public StripeResponse<Charge> createChargeNoCatch(
    int amount,
    String currency,
    String source
) {
    var bodyParams = new HashMap<String, String>();
    bodyParams.put("amount", String.valueOf(amount));
    bodyParams.put("currency", currency);
    bodyParams.put("source", source);

    var request = HttpRequest.newBuilder()
        .uri(URI.create(BASE_URL + "/charges"))
        .timeout(TIMEOUT)
        .header("Authorization", "Bearer " + apiKey)
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(buildFormBody(bodyParams))
        .build();

    // BUG: No try/catch - IOException/InterruptedException unhandled
    var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

    if (response.statusCode() >= 200 && response.statusCode() < 300) {
        var charge = objectMapper.readValue(response.body(), Charge.class);
        return new StripeResponse.Success<>(charge);
    } else {
        var errorBody = parseErrorResponse(response.body());
        return new StripeResponse.Error<>(
            response.statusCode(),
            errorBody.message(),
            errorBody.code()
        );
    }
}
```

**Why Stricture catches this:** TQ-error-path-coverage requires all network I/O calls to have explicit error handling. HttpClient.send() throws checked exceptions that must be caught and converted to domain errors.

---

### B02 — No Status Code Check (CTR-status-code-handling)

**Bug:** Missing HTTP status code validation, assumes all responses are successful
**Expected violation:** `CTR-status-code-handling`

```java
public Charge getChargeNoStatusCheck(String chargeId)
    throws IOException, InterruptedException {

    var request = HttpRequest.newBuilder()
        .uri(URI.create(BASE_URL + "/charges/" + chargeId))
        .timeout(TIMEOUT)
        .header("Authorization", "Bearer " + apiKey)
        .GET()
        .build();

    var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

    // BUG: No status code check - blindly deserializes any response
    var charge = objectMapper.readValue(response.body(), Charge.class);
    return charge;
}
```

**Why Stricture catches this:** CTR-status-code-handling requires HTTP responses to be validated before deserialization. Error responses (4xx/5xx) have different shapes than success responses and will cause deserialization failures.

---

### B03 — Shallow Test Assertions (TQ-no-shallow-assertions)

**Bug:** Test only checks assertNotNull instead of validating field values
**Expected violation:** `TQ-no-shallow-assertions`

```java
@Test
void testCreateChargeShallowAssertions() throws Exception {
    var client = StripeClient.builder()
        .apiKey("sk_test_123")
        .httpClient(mockHttpClient)
        .build();

    var response = client.createCharge(
        2000,
        "usd",
        "tok_visa",
        Optional.empty(),
        Optional.of("Test charge"),
        Optional.empty()
    );

    // BUG: Only checks non-null, doesn't validate fields
    assertNotNull(response);
    assertTrue(response instanceof StripeResponse.Success);
}
```

**Why Stricture catches this:** TQ-no-shallow-assertions requires field-level validation in tests. Tests must verify amount, currency, status, and other critical fields match expected values, not just check for non-null.

---

### B04 — Missing Negative Tests (TQ-negative-cases)

**Bug:** No parameterized tests for error codes (invalid_amount, card_declined, etc.)
**Expected violation:** `TQ-negative-cases`

```java
@Test
void testCreateChargeOnlyHappyPath() throws Exception {
    var client = StripeClient.builder()
        .apiKey("sk_test_123")
        .httpClient(mockHttpClient)
        .build();

    // BUG: Only tests success case, no negative tests
    var response = client.createCharge(
        2000,
        "usd",
        "tok_visa",
        Optional.empty(),
        Optional.empty(),
        Optional.empty()
    );

    assertTrue(response instanceof StripeResponse.Success);
    var charge = ((StripeResponse.Success<Charge>) response).data();
    assertEquals(2000, charge.amount());
    assertEquals("usd", charge.currency());
}
```

**Why Stricture catches this:** TQ-negative-cases requires @ParameterizedTest coverage for error scenarios: amount_too_small, invalid_currency, card_declined, insufficient_funds, etc. Each error code path must be tested.

---

### B05 — Request Missing Required Fields (CTR-request-shape)

**Bug:** Charge creation missing "currency" field in request body
**Expected violation:** `CTR-request-shape`

```java
public StripeResponse<Charge> createChargeMissingCurrency(
    int amount,
    String source
) throws IOException, InterruptedException {

    // BUG: Missing required "currency" field
    var bodyParams = new HashMap<String, String>();
    bodyParams.put("amount", String.valueOf(amount));
    bodyParams.put("source", source);

    var request = HttpRequest.newBuilder()
        .uri(URI.create(BASE_URL + "/charges"))
        .timeout(TIMEOUT)
        .header("Authorization", "Bearer " + apiKey)
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(buildFormBody(bodyParams))
        .build();

    var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

    if (response.statusCode() >= 200 && response.statusCode() < 300) {
        var charge = objectMapper.readValue(response.body(), Charge.class);
        return new StripeResponse.Success<>(charge);
    } else {
        var errorBody = parseErrorResponse(response.body());
        return new StripeResponse.Error<>(
            response.statusCode(),
            errorBody.message(),
            errorBody.code()
        );
    }
}
```

**Why Stricture catches this:** CTR-request-shape validates request payloads against API manifests. The Stripe API requires "currency" for all charge creations. Missing required fields cause 400 errors.

---

### B06 — Response Type Mismatch (CTR-response-shape)

**Bug:** Charge record missing "livemode" field that exists in Stripe API responses
**Expected violation:** `CTR-response-shape`

```java
// BUG: Missing "livemode" field from Stripe API response
public record ChargeIncomplete(
    @JsonProperty("id") String id,
    @JsonProperty("object") String object,
    @JsonProperty("amount") int amount,
    @JsonProperty("currency") String currency,
    @JsonProperty("status") ChargeStatus status,
    @JsonProperty("paid") boolean paid,
    @JsonProperty("created") long created
    // Missing: livemode, metadata, receipt_url, etc.
) {}

public StripeResponse<ChargeIncomplete> createChargeIncompleteModel(
    int amount,
    String currency,
    String source
) throws IOException, InterruptedException {
    // Implementation uses ChargeIncomplete instead of Charge
    // API returns fields that aren't captured in the model
}
```

**Why Stricture catches this:** CTR-response-shape validates response models against API manifests. All fields returned by the API must be present in the model, even if marked as Optional. Missing fields indicate incomplete understanding of the API contract.

---

### B07 — Wrong Field Types (CTR-manifest-conformance)

**Bug:** Using double for amount instead of long (loses precision, violates API contract)
**Expected violation:** `CTR-manifest-conformance`

```java
// BUG: amount should be int/long (cents), not double
public record ChargeWrongTypes(
    @JsonProperty("id") String id,
    @JsonProperty("object") String object,
    @JsonProperty("amount") double amount, // WRONG: should be int
    @JsonProperty("currency") String currency,
    @JsonProperty("status") ChargeStatus status,
    @JsonProperty("paid") boolean paid,
    @JsonProperty("created") long created
) {
    public ChargeWrongTypes {
        Objects.requireNonNull(id, "id required");
        Objects.requireNonNull(currency, "currency required");
        Objects.requireNonNull(status, "status required");

        // BUG: Validation uses double comparison (imprecise)
        if (amount < 0.50 || amount > 999999.99) {
            throw new IllegalArgumentException("Invalid amount");
        }
    }
}
```

**Why Stricture catches this:** CTR-manifest-conformance requires field types to match API specifications exactly. Stripe amounts are integers in cents to avoid floating-point precision issues. Using double causes rounding errors and violates the API contract.

---

### B08 — Incomplete Enum Handling (CTR-strictness-parity)

**Bug:** ChargeStatus enum missing FAILED state, only has SUCCEEDED and PENDING
**Expected violation:** `CTR-strictness-parity`

```java
// BUG: Missing FAILED status
public enum ChargeStatusIncomplete {
    @JsonProperty("succeeded") SUCCEEDED,
    @JsonProperty("pending") PENDING
    // Missing: FAILED
}

public record ChargeIncompleteEnum(
    @JsonProperty("id") String id,
    @JsonProperty("amount") int amount,
    @JsonProperty("currency") String currency,
    @JsonProperty("status") ChargeStatusIncomplete status // Incomplete enum
) {
    public ChargeIncompleteEnum {
        Objects.requireNonNull(status, "status required");

        // BUG: Logic assumes only SUCCEEDED/PENDING, ignores FAILED
        if (status == ChargeStatusIncomplete.SUCCEEDED) {
            // Process successful charge
        } else if (status == ChargeStatusIncomplete.PENDING) {
            // Wait for charge to complete
        }
        // Missing: FAILED case handling
    }
}
```

**Why Stricture catches this:** CTR-strictness-parity requires enums to include all values documented in the API. Incomplete enums cause deserialization failures when the API returns undocumented values (e.g., "failed" status).

---

### B09 — Missing Range Validation (CTR-strictness-parity)

**Bug:** No validation that amount meets Stripe's minimum requirement (50 cents)
**Expected violation:** `CTR-strictness-parity`

```java
public StripeResponse<Charge> createChargeNoRangeCheck(
    int amount,
    String currency,
    String source
) throws IOException, InterruptedException {

    // BUG: No range validation - allows amounts < 50 cents
    var bodyParams = new HashMap<String, String>();
    bodyParams.put("amount", String.valueOf(amount));
    bodyParams.put("currency", currency);
    bodyParams.put("source", source);

    var request = HttpRequest.newBuilder()
        .uri(URI.create(BASE_URL + "/charges"))
        .timeout(TIMEOUT)
        .header("Authorization", "Bearer " + apiKey)
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(buildFormBody(bodyParams))
        .build();

    var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

    if (response.statusCode() >= 200 && response.statusCode() < 300) {
        var charge = objectMapper.readValue(response.body(), Charge.class);
        return new StripeResponse.Success<>(charge);
    } else {
        var errorBody = parseErrorResponse(response.body());
        return new StripeResponse.Error<>(
            response.statusCode(),
            errorBody.message(),
            errorBody.code()
        );
    }
}
```

**Why Stricture catches this:** CTR-strictness-parity requires client-side validation to match API constraints. Stripe enforces minimum amounts (50 cents for USD) to prevent invalid charges. Missing validation causes unnecessary round-trips and poor error messages.

---

### B10 — Format Not Validated (CTR-strictness-parity)

**Bug:** No validation that charge ID starts with "ch_" prefix
**Expected violation:** `CTR-strictness-parity`

```java
public StripeResponse<Charge> getChargeNoFormatCheck(String chargeId)
    throws IOException, InterruptedException {

    // BUG: No format validation - accepts any string
    var request = HttpRequest.newBuilder()
        .uri(URI.create(BASE_URL + "/charges/" + chargeId))
        .timeout(TIMEOUT)
        .header("Authorization", "Bearer " + apiKey)
        .GET()
        .build();

    var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

    if (response.statusCode() >= 200 && response.statusCode() < 300) {
        var charge = objectMapper.readValue(response.body(), Charge.class);
        return new StripeResponse.Success<>(charge);
    } else {
        var errorBody = parseErrorResponse(response.body());
        return new StripeResponse.Error<>(
            response.statusCode(),
            errorBody.message(),
            errorBody.code()
        );
    }
}
```

**Why Stricture catches this:** CTR-strictness-parity requires format validation for resource IDs. All Stripe charge IDs start with "ch_" prefix. Missing validation allows invalid IDs to reach the API, causing unnecessary network calls and unclear errors.

---

### B11 — Precision Loss on Currency (CTR-strictness-parity)

**Bug:** Using double for monetary calculations causes floating-point precision errors
**Expected violation:** `CTR-strictness-parity`

```java
public class ChargeCalculatorBroken {
    // BUG: Double arithmetic for money causes precision loss
    public double calculateTotal(List<Double> amounts) {
        return amounts.stream()
            .reduce(0.0, Double::sum);
    }

    public StripeResponse<Charge> createChargeFromDouble(
        double amountInDollars,
        String currency,
        String source
    ) throws IOException, InterruptedException {

        // BUG: Converting double to int loses precision
        // 19.99 * 100 = 1998.9999999999998 -> 1998 cents
        int amountInCents = (int) (amountInDollars * 100);

        var bodyParams = new HashMap<String, String>();
        bodyParams.put("amount", String.valueOf(amountInCents));
        bodyParams.put("currency", currency);
        bodyParams.put("source", source);

        var request = HttpRequest.newBuilder()
            .uri(URI.create(BASE_URL + "/charges"))
            .timeout(TIMEOUT)
            .header("Authorization", "Bearer " + "apiKey")
            .header("Content-Type", "application/x-www-form-urlencoded")
            .POST(buildFormBody(bodyParams))
            .build();

        var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        // ... rest of implementation
    }
}
```

**Why Stricture catches this:** CTR-strictness-parity requires exact decimal arithmetic for monetary values. Floating-point arithmetic causes rounding errors (0.1 + 0.2 != 0.3). Use int/long for cents or BigDecimal for precision.

---

### B12 — Nullable Field Crashes (CTR-response-shape)

**Bug:** Accessing Optional field without checking presence causes NullPointerException
**Expected violation:** `CTR-response-shape`

```java
public String formatChargeDescription(Charge charge) {
    // BUG: Calls .get() without checking .isPresent()
    // Throws NullPointerException if description is empty
    return "Charge " + charge.id() + ": " + charge.description().get();
}

@Test
void testFormatChargeDescriptionCrashes() {
    var charge = new Charge(
        "ch_123",
        "charge",
        2000,
        2000,
        0,
        "txn_123",
        "usd",
        "cus_123",
        Optional.empty(), // No description
        ChargeStatus.SUCCEEDED,
        Optional.empty(),
        Optional.empty(),
        true,
        false,
        1640000000L
    );

    // BUG: Crashes with NoSuchElementException
    formatChargeDescription(charge);
}
```

**Why Stricture catches this:** CTR-response-shape requires safe handling of optional fields. Optional.get() without isPresent() check causes runtime crashes. Use orElse(), orElseThrow(), or ifPresent() instead.

---

### B13 — Missing Webhook Verification (CTR-request-shape)

**Bug:** Webhook handler doesn't verify Stripe-Signature header before processing payload
**Expected violation:** `CTR-request-shape`

```java
public class WebhookHandlerInsecure {
    private final ObjectMapper objectMapper = new ObjectMapper();

    // BUG: No signature verification - accepts any POST
    public void handleWebhook(String payload) throws IOException {
        var event = objectMapper.readTree(payload);
        var eventType = event.get("type").asText();

        if (eventType.equals("charge.succeeded")) {
            var charge = event.get("data").get("object");
            var chargeId = charge.get("id").asText();
            var amount = charge.get("amount").asInt();

            // Process charge without verifying authenticity
            processSuccessfulCharge(chargeId, amount);
        }
    }

    private void processSuccessfulCharge(String chargeId, int amount) {
        // Critical business logic executed without auth
        System.out.println("Processing charge: " + chargeId);
    }
}
```

**Why Stricture catches this:** CTR-request-shape requires webhook signature verification before processing. Without HMAC validation of Stripe-Signature header, attackers can forge webhooks to trigger unauthorized actions.

---

### B14 — Pagination Terminated Early (CTR-response-shape)

**Bug:** Iterator ignores hasMore flag and stops after first page
**Expected violation:** `CTR-response-shape`

```java
private final class ChargeIteratorBroken implements Iterator<Charge> {
    private final int pageSize;
    private List<Charge> currentPage = new ArrayList<>();
    private int currentIndex = 0;
    private boolean fetched = false;

    ChargeIteratorBroken(int pageSize) {
        this.pageSize = pageSize;
    }

    @Override
    public boolean hasNext() {
        if (currentIndex < currentPage.size()) {
            return true;
        }

        // BUG: Only fetches first page, ignores hasMore
        if (!fetched) {
            fetchFirstPage();
            fetched = true;
            return currentIndex < currentPage.size();
        }

        return false;
    }

    @Override
    public Charge next() {
        if (!hasNext()) {
            throw new NoSuchElementException();
        }
        return currentPage.get(currentIndex++);
    }

    private void fetchFirstPage() {
        try {
            var request = HttpRequest.newBuilder()
                .uri(URI.create(BASE_URL + "/charges?limit=" + pageSize))
                .timeout(TIMEOUT)
                .header("Authorization", "Bearer " + "apiKey")
                .GET()
                .build();

            var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                var chargeList = objectMapper.readValue(response.body(), ChargeList.class);
                currentPage = chargeList.data();
                currentIndex = 0;
                // BUG: Ignores chargeList.hasMore() - never fetches page 2+
            }
        } catch (IOException | InterruptedException e) {
            throw new RuntimeException("Failed to fetch charges", e);
        }
    }
}
```

**Why Stricture catches this:** CTR-response-shape requires proper pagination handling. List endpoints return has_more to indicate additional pages. Ignoring this flag causes incomplete data retrieval.

---

### B15 — Race Condition (CTR-request-shape)

**Bug:** No Idempotency-Key header causes duplicate charges on network retry
**Expected violation:** `CTR-request-shape`

```java
public StripeResponse<Charge> createChargeNoIdempotency(
    int amount,
    String currency,
    String source,
    int maxRetries
) throws IOException, InterruptedException {

    var bodyParams = new HashMap<String, String>();
    bodyParams.put("amount", String.valueOf(amount));
    bodyParams.put("currency", currency);
    bodyParams.put("source", source);

    var request = HttpRequest.newBuilder()
        .uri(URI.create(BASE_URL + "/charges"))
        .timeout(TIMEOUT)
        .header("Authorization", "Bearer " + apiKey)
        .header("Content-Type", "application/x-www-form-urlencoded")
        // BUG: No Idempotency-Key header
        .POST(buildFormBody(bodyParams))
        .build();

    // BUG: Retries without idempotency key create duplicate charges
    for (int i = 0; i < maxRetries; i++) {
        try {
            var response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                var charge = objectMapper.readValue(response.body(), Charge.class);
                return new StripeResponse.Success<>(charge);
            } else if (response.statusCode() >= 500) {
                // Retry on server error - may create duplicate charge
                Thread.sleep(1000 * (i + 1));
                continue;
            } else {
                var errorBody = parseErrorResponse(response.body());
                return new StripeResponse.Error<>(
                    response.statusCode(),
                    errorBody.message(),
                    errorBody.code()
                );
            }
        } catch (IOException e) {
            if (i == maxRetries - 1) throw e;
            Thread.sleep(1000 * (i + 1));
        }
    }

    throw new IOException("Max retries exceeded");
}
```

**Why Stricture catches this:** CTR-request-shape requires Idempotency-Key header for non-idempotent operations. Without it, network timeouts or 5xx errors cause retries that create duplicate charges. Stripe uses this header to safely deduplicate requests.

---
