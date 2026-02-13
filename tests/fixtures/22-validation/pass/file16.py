class Product(ProductBase):
    created_at: datetime = Field(default_factory=datetime.utcnow)

    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat(),  # Returns "2024-01-15T12:34:56.789Z"
        }
