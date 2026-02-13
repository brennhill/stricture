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
