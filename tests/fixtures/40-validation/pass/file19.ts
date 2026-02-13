// src/services/profile-service.ts
export interface UserProfile {
  userId: string;
  name: string;
  contact: {
    email: string;
    phone: string;
    address: {
      street: string;
      city: string;
      zip: string;
    };
  };
}

export function getProfile(userId: string): UserProfile {
  return database.fetchProfile(userId);
}
