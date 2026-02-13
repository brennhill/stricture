// DEFECT B03: Shallow assertion — only checks ID exists
func TestCreateOrder_Success(t *testing.T) {
	// ...
	var order models.Order
	err := json.NewDecoder(w.Body).Decode(&order)
	require.NoError(t, err)

	assert.NotNil(t, order.ID)  // BUG: Should validate it's a valid UUID
	assert.Equal(t, reqBody.CustomerID, order.CustomerID)
	assert.Equal(t, reqBody.TotalAmount, order.TotalAmount)
}
