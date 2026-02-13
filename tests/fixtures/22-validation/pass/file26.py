class ProductBase(BaseModel):
    description: Optional[str] = Field(None, max_length=2000)  # Can be None
