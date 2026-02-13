# Python expects:
{"stock_count": 150, "warehouse_id": "warehouse-west", "created_at": "..."}

# Go sends:
{"stockCount": 150, "warehouseId": "warehouse-west", "createdAt": "..."}

# Pydantic validation error:
# Field required: stock_count, warehouse_id, created_at
