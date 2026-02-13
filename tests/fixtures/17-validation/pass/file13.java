public String formatChargeDescription(Charge charge) {
    // BUG: Calls .get() without checking .isPresent()
    // Throws NullPointerException if description is empty
    return "Charge " + charge.id() + ": " + charge.description().get();
}

@Test
void testFormatChargeDescriptionCrashes() {
    var charge = new Charge(
        "ch_123",
        "charge",
        2000,
        2000,
        0,
        "txn_123",
        "usd",
        "cus_123",
        Optional.empty(), // No description
        ChargeStatus.SUCCEEDED,
        Optional.empty(),
        Optional.empty(),
        true,
        false,
        1640000000L
    );

    // BUG: Crashes with NoSuchElementException
    formatChargeDescription(charge);
}
