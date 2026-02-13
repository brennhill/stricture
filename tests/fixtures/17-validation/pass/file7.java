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
