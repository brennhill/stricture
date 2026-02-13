@Test
void testCreateChargeShallowAssertions() throws Exception {
    var client = StripeClient.builder()
        .apiKey("sk_test_123")
        .httpClient(mockHttpClient)
        .build();

    var response = client.createCharge(
        2000,
        "usd",
        "tok_visa",
        Optional.empty(),
        Optional.of("Test charge"),
        Optional.empty()
    );

    // BUG: Only checks non-null, doesn't validate fields
    assertNotNull(response);
    assertTrue(response instanceof StripeResponse.Success);
}
