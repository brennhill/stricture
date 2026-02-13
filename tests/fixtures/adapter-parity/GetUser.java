// GetUser.java — User service class.

package service;

public class UserService {
    public record User(int id, String name, String email) {}

    public User getUser(String id) {
        if (id == null || id.isEmpty()) {
            throw new IllegalArgumentException("get user: id is empty. Provide a valid user ID.");
        }
        return new User(1, "Alice", "alice@example.com");
    }

    public User createUser(String name, String email) {
        if (name == null || name.isEmpty()) {
            throw new IllegalArgumentException("create user: name is empty. Provide a name.");
        }
        return new User(2, name, email);
    }
}
