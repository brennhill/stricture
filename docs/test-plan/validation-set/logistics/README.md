# 20 — Multi-Company Logistics Ecosystem

**Why included:** Cross-company microservices integration -- the "exotic" test case. Demonstrates how Stricture catches contract mismatches BETWEEN services owned by different teams and companies. A complex dropshipping e-commerce flow involving 7 companies, 3 languages, and 8 inter-service contracts.

## Companies

| # | Company | Language | Role | Based on |
|---|---------|----------|------|----------|
| 1 | ShopStream | TypeScript | E-commerce storefront | Shopify API patterns |
| 2 | PayCore | Go | Payment processing | Stripe API patterns |
| 3 | VaultStore | Go | Warehouse management | ShipBob API patterns |
| 4 | LabelForge | TypeScript | Shipping label generation | EasyPost API patterns |
| 5 | SwiftHaul | Go | Last-mile delivery | FedEx/UPS API patterns |
| 6 | OceanBridge | Python | International shipping | Flexport/Maersk API patterns |
| 7 | PingWave | TypeScript | Notifications | Twilio + SendGrid patterns |

## Order Flow

```
Customer → ShopStream → PayCore (charge)
                     → VaultStore (reserve inventory)
                          → LabelForge (create label)
                               → SwiftHaul (schedule pickup)
                                    → OceanBridge (international handoff)
                                         → PingWave (customs notification)
                                              → ShopStream (delivery callback)
PayCore → ShopStream (payment webhook)
```

## Files

| File | Lines | Content |
|------|-------|---------|
| [companies.md](companies.md) | ~490 | API specs for all 7 companies (endpoints, request/response schemas) |
| [contracts.md](contracts.md) | ~540 | 8 inter-service contracts (C1-C8) + Stricture manifest |
| [perfect.md](perfect.md) | ~1,225 | Correct integration code for all 7 services (zero violations) |
| [mismatches.md](mismatches.md) | ~800 | 10 cross-company mismatch scenarios (M01-M10) |

## Key Conventions

| Concern | Convention | Notes |
|---------|-----------|-------|
| Money | Integer cents | Except OceanBridge customs (string decimal) |
| Weight | Grams internally | OceanBridge expects kg (conversion required) |
| Dimensions | Millimeters internally | LabelForge/SwiftHaul expect cm (conversion required) |
| Country | ISO 3166-1 alpha-2 | Always uppercase 2-letter |
| Timestamps | ISO 8601 UTC | Always include timezone |
| Currency codes | Lowercase | Except OceanBridge (uppercase) |
| Status enums | snake_case | Except SwiftHaul (SCREAMING_SNAKE_CASE) |

## Cross-Company Mismatches (M01-M10)

| ID | Mismatch | Boundary | Primary Rule |
|----|----------|----------|-------------|
| M01 | Float dollars vs integer cents | ShopStream ↔ PayCore | CTR-strictness-parity |
| M02 | Weight grams vs kilograms | VaultStore → OceanBridge | CTR-request-shape |
| M03 | Dimensions mm sent as cm | VaultStore → LabelForge | CTR-strictness-parity |
| M04 | Country name vs ISO code | SwiftHaul → OceanBridge | CTR-request-shape |
| M05 | Enum case mismatch (snake vs SCREAMING) | SwiftHaul → ShopStream | CTR-strictness-parity |
| M06 | String decimal vs integer cents | SwiftHaul → OceanBridge | CTR-request-shape |
| M07 | Missing webhook signature verification | PingWave → ShopStream | CTR-manifest-conformance |
| M08 | Timezone-naive datetime | OceanBridge → PingWave | CTR-strictness-parity |
| M09 | Date format degradation across hops | PayCore → ShopStream → VaultStore | CTR-response-shape |
| M10 | Implicit callback contract (undocumented) | PingWave → ShopStream | CTR-manifest-conformance |
