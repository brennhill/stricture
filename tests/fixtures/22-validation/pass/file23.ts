// BUG: No email format validation
interface ContactInfo {
    email: string;  // Should validate email format like Pydantic
    phone: string;
}

async function updateContact(productId: string, contact: ContactInfo): Promise<void> {
    const response = await fetch(`${API_BASE}/api/products/${productId}/contact`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(contact),
    });
    if (!response.ok) throw new Error(`HTTP ${response.status}`);
}

// Usage:
await updateContact("123", {
    email: "not-an-email",  // Invalid format — server will reject, client allows
    phone: "555-1234",
});
