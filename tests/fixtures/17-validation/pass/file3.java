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
