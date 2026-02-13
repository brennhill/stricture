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
