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
