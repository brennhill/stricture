// index.ts -- Auth module public API.
// V12: Re-exports from another module's internal directory.

export { validateToken, generateToken, type TokenPayload } from "./token-validator";

// VIOLATION: This re-export reaches into the billing module's internal
// directory. Even though it goes through auth's own index.ts (which is the
// auth module's public API), it creates a dependency from the auth module
// into billing's internals -- violating billing's module boundary.
export { createPaymentIntent } from "../billing/internal/stripe-adapter";  // <-- VIOLATION
