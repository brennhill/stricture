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
