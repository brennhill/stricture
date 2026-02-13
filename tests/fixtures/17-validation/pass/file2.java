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
