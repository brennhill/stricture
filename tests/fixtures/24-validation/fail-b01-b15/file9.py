# B12-null-safety-violation/python-client/client.py
def get_item_description_upper(self, item_id: str) -> str:
    """Get item description in uppercase"""
    item = self.get_item(item_id)

    # ❌ description can be None, will raise AttributeError
    return item.description.upper()
