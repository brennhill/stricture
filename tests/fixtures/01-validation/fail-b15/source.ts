// B15: Read-then-update on customer without any concurrency protection.

class StripeClientB15 {
  private readonly baseUrl = "https://api.stripe.com";
  private readonly apiKey: string;

  constructor(apiKey: string) {
    this.apiKey = apiKey;
  }

  async getCustomer(customerId: string): Promise<StripeResult<StripeCustomer>> {
    validateCustomerId(customerId);

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/customers/${customerId}`, {
        headers: { Authorization: `Bearer ${this.apiKey}` },
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCustomer };
  }

  async updateCustomer(
    customerId: string,
    updates: Partial<Pick<StripeCustomer, "email" | "name" | "description">>
  ): Promise<StripeResult<StripeCustomer>> {
    validateCustomerId(customerId);

    const body = new URLSearchParams();
    if (updates.email !== undefined) body.set("email", updates.email ?? "");
    if (updates.name !== undefined) body.set("name", updates.name ?? "");
    if (updates.description !== undefined) body.set("description", updates.description ?? "");

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/customers/${customerId}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCustomer };
  }

  // BUG: Read-modify-write without any concurrency protection.
  // Two concurrent calls to addTagToCustomer can overwrite each other:
  //
  // Timeline:
  //   T1: reads customer.metadata = { tier: "gold" }
  //   T2: reads customer.metadata = { tier: "gold" }
  //   T1: writes metadata = { tier: "gold", region: "us" }
  //   T2: writes metadata = { tier: "gold", cohort: "2024" }
  //   Result: metadata = { tier: "gold", cohort: "2024" }
  //   Lost: region: "us" (T1's write was silently overwritten)
  //
  // The manifest does not provide an ETag or version field for optimistic
  // locking, so the client must use Idempotency-Key or serialize access.

  async addMetadataToCustomer(
    customerId: string,
    key: string,
    value: string
  ): Promise<StripeResult<StripeCustomer>> {
    // Step 1: Read current customer
    const current = await this.getCustomer(customerId);
    if (!current.ok) return current;

    // BUG: No lock, no version check, no idempotency key.
    // Between the read above and the write below, another process
    // can modify the customer, and this write will silently overwrite
    // those changes.

    // Step 2: Merge new metadata with existing
    const updatedMetadata = { ...current.data.metadata, [key]: value };

    // Step 3: Write back (overwrites any concurrent modifications)
    const body = new URLSearchParams();
    for (const [k, v] of Object.entries(updatedMetadata)) {
      body.set(`metadata[${k}]`, v);
    }

    let response: Response;
    try {
      response = await fetch(`${this.baseUrl}/v1/customers/${customerId}`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
          "Content-Type": "application/x-www-form-urlencoded",
        },
        body: body.toString(),
      });
    } catch (err) {
      throw new Error(`Network error: ${(err as Error).message}`);
    }

    if (!response.ok) {
      const errorBody = (await response.json()) as StripeError;
      return { ok: false, status: response.status, error: errorBody };
    }

    return { ok: true, data: (await response.json()) as StripeCustomer };
  }
}
