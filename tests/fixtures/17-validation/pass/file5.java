@Test
void testCreateChargeOnlyHappyPath() throws Exception {
    var client = StripeClient.builder()
        .apiKey("sk_test_123")
        .httpClient(mockHttpClient)
        .build();

    // BUG: Only tests success case, no negative tests
    var response = client.createCharge(
        2000,
        "usd",
        "tok_visa",
        Optional.empty(),
        Optional.empty(),
        Optional.empty()
    );

    assertTrue(response instanceof StripeResponse.Success);
    var charge = ((StripeResponse.Success<Charge>) response).data();
    assertEquals(2000, charge.amount());
    assertEquals("usd", charge.currency());
}
