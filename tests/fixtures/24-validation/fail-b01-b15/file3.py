item = client.get_item("item-001")
print(item.warehouse_id)  # ❌ AttributeError: 'Item' object has no attribute 'warehouse_id'
