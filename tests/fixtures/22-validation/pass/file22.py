from pydantic import BaseModel, EmailStr

class ContactInfo(BaseModel):
    email: EmailStr  # Pydantic validates email format
    phone: str
