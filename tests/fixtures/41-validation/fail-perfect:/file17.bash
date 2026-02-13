stricture analyze test/

ERROR: test/payment-processor.anti1.test.ts
  TQ-no-shallow-assertions (lines 14-22, 28-30, 38)
  9 assertions use toBeDefined() without value verification

ERROR: test/payment-processor.anti2.test.ts
  TQ-negative-cases (lines 5-104)
  10 success tests, 0 error cases
  Missing: negative amount, zero amount, invalid currency, invalid transaction ID

WARNING: test/payment-processor.anti3.test.ts
  TQ-mock-scope (lines 6-11)
  Global mock affects all 4 tests. Move to beforeEach or test scope

WARNING: test/data-transformer.anti4.test.ts
  TQ-test-naming (lines 6, 12, 18, 24, 30, 36, 42)
  7 generic test names. Use descriptive names explaining what is tested

ERROR: test/payment-processor.anti5.test.ts
  TQ-no-shallow-assertions (lines 6-58)
  5 tests with 0 assertions

WARNING: test/payment-processor.anti6.test.ts
  TQ-assertion-depth (lines 15-23, 30-31, 43, 55-57)
  12 assertions use expect.any() without value verification

ERROR: test/auth-middleware.anti7.test.ts
  TQ-test-isolation (lines 7-9)
  Shared mutable state: 'auth', 'testToken', 'testUser'
  Tests are order-dependent

WARNING: test/pagination.anti8.test.ts
  TQ-boundary-tested (lines 5-42)
  Missing boundary tests: page ≤ 0, page > totalPages, limit ≤ 0, empty array

WARNING: test/payment-processor.anti9.test.ts
  TQ-return-type-verified (lines 6-18, 21-27, 44-50, 53-58)
  4 tests verify partial return types (checks 1-2 of 8 fields)

ERROR: test/payment-processor.anti10.test.ts
  Multiple violations: 8/10 anti-patterns detected
  Critical: This test provides false confidence

Summary:
  Files analyzed: 10
  Tests analyzed: 63
  Violations: 10 ERROR, 5 WARNING
  Coverage achieved: 100%
  Actual validation: ~15%
