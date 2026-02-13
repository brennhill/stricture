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
