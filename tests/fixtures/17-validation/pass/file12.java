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
