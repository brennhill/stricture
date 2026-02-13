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
