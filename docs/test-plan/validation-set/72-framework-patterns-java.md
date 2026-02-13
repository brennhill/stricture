# 72 — Java Framework Patterns

**Purpose:** Validate route detection across Java web frameworks (Spring Boot, Jakarta EE, Micronaut).

**Coverage:** 3 frameworks × (1 perfect + 3 violations) = 12 test cases

---

## Spring Boot — Perfect Implementation

**File:** `src/main/java/com/example/api/UserController.java`

```java
package com.example.api;

import com.example.dto.CreateUserRequest;
import com.example.dto.UserResponse;
import com.example.service.UserService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/api/users")
public class UserController {

    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @GetMapping
    public ResponseEntity<List<UserResponse>> listUsers() {
        List<UserResponse> users = userService.getAllUsers();
        return ResponseEntity.ok(users);
    }

    @GetMapping("/{id}")
    public ResponseEntity<UserResponse> getUser(@PathVariable Long id) {
        UserResponse user = userService.getUserById(id);
        if (user == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(user);
    }

    @PostMapping
    public ResponseEntity<UserResponse> createUser(@Valid @RequestBody CreateUserRequest request) {
        UserResponse created = userService.createUser(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    @PutMapping("/{id}")
    public ResponseEntity<UserResponse> updateUser(
            @PathVariable Long id,
            @Valid @RequestBody CreateUserRequest request) {
        UserResponse updated = userService.updateUser(id, request);
        if (updated == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(updated);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteUser(@PathVariable Long id) {
        boolean deleted = userService.deleteUser(id);
        if (!deleted) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.noContent().build();
    }
}
```

**File:** `src/main/java/com/example/dto/CreateUserRequest.java`

```java
package com.example.dto;

import javax.validation.constraints.Email;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Size;

public class CreateUserRequest {

    @NotBlank(message = "Username is required")
    @Size(min = 3, max = 50, message = "Username must be between 3 and 50 characters")
    private String username;

    @NotBlank(message = "Email is required")
    @Email(message = "Email must be valid")
    private String email;

    @NotBlank(message = "Password is required")
    @Size(min = 8, message = "Password must be at least 8 characters")
    private String password;

    public String getUsername() {
        return username;
    }

    public void setUsername(String username) {
        this.username = username;
    }

    public String getEmail() {
        return email;
    }

    public void setEmail(String email) {
        this.email = email;
    }

    public String getPassword() {
        return password;
    }

    public void setPassword(String password) {
        this.password = password;
    }
}
```

**File:** `src/main/java/com/example/dto/UserResponse.java`

```java
package com.example.dto;

public class UserResponse {
    private Long id;
    private String username;
    private String email;

    public UserResponse(Long id, String username, String email) {
        this.id = id;
        this.username = username;
        this.email = email;
    }

    public Long getId() {
        return id;
    }

    public String getUsername() {
        return username;
    }

    public String getEmail() {
        return email;
    }
}
```

**File:** `src/main/java/com/example/service/UserService.java`

```java
package com.example.service;

import com.example.dto.CreateUserRequest;
import com.example.dto.UserResponse;
import com.example.model.User;
import com.example.repository.UserRepository;
import org.springframework.stereotype.Service;
import java.util.List;
import java.util.stream.Collectors;

@Service
public class UserService {

    private final UserRepository userRepository;

    public UserService(UserRepository userRepository) {
        this.userRepository = userRepository;
    }

    public List<UserResponse> getAllUsers() {
        return userRepository.findAll().stream()
                .map(this::toResponse)
                .collect(Collectors.toList());
    }

    public UserResponse getUserById(Long id) {
        return userRepository.findById(id)
                .map(this::toResponse)
                .orElse(null);
    }

    public UserResponse createUser(CreateUserRequest request) {
        User user = new User();
        user.setUsername(request.getUsername());
        user.setEmail(request.getEmail());
        user.setPassword(hashPassword(request.getPassword()));
        User saved = userRepository.save(user);
        return toResponse(saved);
    }

    public UserResponse updateUser(Long id, CreateUserRequest request) {
        return userRepository.findById(id)
                .map(user -> {
                    user.setUsername(request.getUsername());
                    user.setEmail(request.getEmail());
                    if (request.getPassword() != null) {
                        user.setPassword(hashPassword(request.getPassword()));
                    }
                    User saved = userRepository.save(user);
                    return toResponse(saved);
                })
                .orElse(null);
    }

    public boolean deleteUser(Long id) {
        if (userRepository.existsById(id)) {
            userRepository.deleteById(id);
            return true;
        }
        return false;
    }

    private UserResponse toResponse(User user) {
        return new UserResponse(user.getId(), user.getUsername(), user.getEmail());
    }

    private String hashPassword(String password) {
        // In production, use BCrypt or similar
        return "hashed_" + password;
    }
}
```

**Expected Behavior:**
- Stricture detects all 5 routes (GET /api/users, GET /api/users/{id}, POST /api/users, PUT /api/users/{id}, DELETE /api/users/{id})
- Recognizes @Valid annotations as input validation
- Understands ResponseEntity status codes (200, 201, 204, 404)
- Maps @RestController + @RequestMapping to route definitions
- NO violations reported

---

## Spring Boot — Violation 1: Missing @Valid

**File:** `src/main/java/com/example/api/ProductController.java`

```java
package com.example.api;

import com.example.dto.CreateProductRequest;
import com.example.dto.ProductResponse;
import com.example.service.ProductService;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/products")
public class ProductController {

    private final ProductService productService;

    public ProductController(ProductService productService) {
        this.productService = productService;
    }

    @PostMapping
    public ResponseEntity<ProductResponse> createProduct(@RequestBody CreateProductRequest request) {
        // VIOLATION: Missing @Valid annotation - input not validated
        ProductResponse created = productService.createProduct(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    @PutMapping("/{id}")
    public ResponseEntity<ProductResponse> updateProduct(
            @PathVariable Long id,
            @RequestBody CreateProductRequest request) {
        // VIOLATION: Missing @Valid annotation
        ProductResponse updated = productService.updateProduct(id, request);
        return ResponseEntity.ok(updated);
    }
}
```

**Expected Violation:**
```
POST /api/products: Input validation missing - @RequestBody lacks @Valid annotation
PUT /api/products/{id}: Input validation missing - @RequestBody lacks @Valid annotation
```

---

## Spring Boot — Violation 2: Wrong ResponseEntity Status

**File:** `src/main/java/com/example/api/OrderController.java`

```java
package com.example.api;

import com.example.dto.CreateOrderRequest;
import com.example.dto.OrderResponse;
import com.example.service.OrderService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import javax.validation.Valid;

@RestController
@RequestMapping("/api/orders")
public class OrderController {

    private final OrderService orderService;

    public OrderController(OrderService orderService) {
        this.orderService = orderService;
    }

    @PostMapping
    public ResponseEntity<OrderResponse> createOrder(@Valid @RequestBody CreateOrderRequest request) {
        OrderResponse created = orderService.createOrder(request);
        // VIOLATION: Should return 201 Created, not 200 OK
        return ResponseEntity.ok(created);
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<OrderResponse> deleteOrder(@PathVariable Long id) {
        orderService.deleteOrder(id);
        // VIOLATION: DELETE should return 204 No Content, not 200 with body
        return ResponseEntity.ok(new OrderResponse(id, "deleted"));
    }
}
```

**Expected Violation:**
```
POST /api/orders: Should return 201 Created for resource creation, not 200 OK
DELETE /api/orders/{id}: Should return 204 No Content for successful deletion, not 200 OK
```

---

## Spring Boot — Violation 3: Service Layer Issues

**File:** `src/main/java/com/example/api/PaymentController.java`

```java
package com.example.api;

import com.example.dto.PaymentRequest;
import com.example.dto.PaymentResponse;
import com.example.repository.PaymentRepository;
import com.example.model.Payment;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import javax.validation.Valid;

@RestController
@RequestMapping("/api/payments")
public class PaymentController {

    private final PaymentRepository paymentRepository;

    public PaymentController(PaymentRepository paymentRepository) {
        this.paymentRepository = paymentRepository;
    }

    @PostMapping
    public ResponseEntity<PaymentResponse> processPayment(@Valid @RequestBody PaymentRequest request) {
        // VIOLATION: Controller directly accessing repository - missing service layer
        Payment payment = new Payment();
        payment.setAmount(request.getAmount());
        payment.setCardNumber(request.getCardNumber());
        Payment saved = paymentRepository.save(payment);
        return ResponseEntity.status(HttpStatus.CREATED)
                .body(new PaymentResponse(saved.getId(), saved.getAmount()));
    }

    @GetMapping("/{id}")
    public ResponseEntity<PaymentResponse> getPayment(@PathVariable Long id) {
        // VIOLATION: Business logic in controller instead of service layer
        return paymentRepository.findById(id)
                .map(payment -> ResponseEntity.ok(
                        new PaymentResponse(payment.getId(), payment.getAmount())))
                .orElse(ResponseEntity.notFound().build());
    }
}
```

**Expected Violation:**
```
POST /api/payments: Controller directly accessing repository - should use service layer
GET /api/payments/{id}: Controller directly accessing repository - should use service layer
```

---

## Jakarta EE (JAX-RS) — Perfect Implementation

**File:** `src/main/java/com/example/resource/UserResource.java`

```java
package com.example.resource;

import com.example.dto.CreateUserRequest;
import com.example.dto.UserResponse;
import com.example.service.UserService;
import javax.inject.Inject;
import javax.validation.Valid;
import javax.ws.rs.*;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;
import java.util.List;

@Path("/users")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class UserResource {

    @Inject
    private UserService userService;

    @GET
    public Response listUsers() {
        List<UserResponse> users = userService.getAllUsers();
        return Response.ok(users).build();
    }

    @GET
    @Path("/{id}")
    public Response getUser(@PathParam("id") Long id) {
        UserResponse user = userService.getUserById(id);
        if (user == null) {
            return Response.status(Response.Status.NOT_FOUND).build();
        }
        return Response.ok(user).build();
    }

    @POST
    public Response createUser(@Valid CreateUserRequest request) {
        UserResponse created = userService.createUser(request);
        return Response.status(Response.Status.CREATED).entity(created).build();
    }

    @PUT
    @Path("/{id}")
    public Response updateUser(@PathParam("id") Long id, @Valid CreateUserRequest request) {
        UserResponse updated = userService.updateUser(id, request);
        if (updated == null) {
            return Response.status(Response.Status.NOT_FOUND).build();
        }
        return Response.ok(updated).build();
    }

    @DELETE
    @Path("/{id}")
    public Response deleteUser(@PathParam("id") Long id) {
        boolean deleted = userService.deleteUser(id);
        if (!deleted) {
            return Response.status(Response.Status.NOT_FOUND).build();
        }
        return Response.noContent().build();
    }
}
```

**Expected Behavior:**
- Stricture detects all 5 routes (GET /users, GET /users/{id}, POST /users, PUT /users/{id}, DELETE /users/{id})
- Recognizes @Valid for input validation
- Understands JAX-RS Response.Status codes (200, 201, 204, 404)
- Maps @Path + HTTP method annotations to routes
- NO violations reported

---

## Jakarta EE (JAX-RS) — Violation 1: Missing @Consumes

**File:** `src/main/java/com/example/resource/ProductResource.java`

```java
package com.example.resource;

import com.example.dto.CreateProductRequest;
import com.example.dto.ProductResponse;
import com.example.service.ProductService;
import javax.inject.Inject;
import javax.validation.Valid;
import javax.ws.rs.*;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;

@Path("/products")
@Produces(MediaType.APPLICATION_JSON)
public class ProductResource {

    @Inject
    private ProductService productService;

    @POST
    // VIOLATION: Missing @Consumes annotation - unclear what content types are accepted
    public Response createProduct(@Valid CreateProductRequest request) {
        ProductResponse created = productService.createProduct(request);
        return Response.status(Response.Status.CREATED).entity(created).build();
    }

    @PUT
    @Path("/{id}")
    // VIOLATION: Missing @Consumes annotation
    public Response updateProduct(@PathParam("id") Long id, @Valid CreateProductRequest request) {
        ProductResponse updated = productService.updateProduct(id, request);
        return Response.ok(updated).build();
    }
}
```

**Expected Violation:**
```
POST /products: Missing @Consumes annotation - should specify accepted content types
PUT /products/{id}: Missing @Consumes annotation - should specify accepted content types
```

---

## Jakarta EE (JAX-RS) — Violation 2: Wrong Status Codes

**File:** `src/main/java/com/example/resource/OrderResource.java`

```java
package com.example.resource;

import com.example.dto.CreateOrderRequest;
import com.example.dto.OrderResponse;
import com.example.service.OrderService;
import javax.inject.Inject;
import javax.validation.Valid;
import javax.ws.rs.*;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;

@Path("/orders")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class OrderResource {

    @Inject
    private OrderService orderService;

    @POST
    public Response createOrder(@Valid CreateOrderRequest request) {
        OrderResponse created = orderService.createOrder(request);
        // VIOLATION: Should return 201 Created, not 200 OK
        return Response.ok(created).build();
    }

    @DELETE
    @Path("/{id}")
    public Response deleteOrder(@PathParam("id") Long id) {
        orderService.deleteOrder(id);
        // VIOLATION: DELETE should return 204 No Content, not 200 with body
        return Response.ok().entity("Deleted").build();
    }

    @GET
    @Path("/{id}")
    public Response getOrder(@PathParam("id") Long id) {
        OrderResponse order = orderService.getOrderById(id);
        if (order == null) {
            // VIOLATION: Should return 404 Not Found, not 204 No Content
            return Response.noContent().build();
        }
        return Response.ok(order).build();
    }
}
```

**Expected Violation:**
```
POST /orders: Should return 201 Created for resource creation, not 200 OK
DELETE /orders/{id}: Should return 204 No Content without body, not 200 OK
GET /orders/{id}: Should return 404 Not Found for missing resource, not 204 No Content
```

---

## Jakarta EE (JAX-RS) — Violation 3: No Input Validation

**File:** `src/main/java/com/example/resource/PaymentResource.java`

```java
package com.example.resource;

import com.example.dto.PaymentRequest;
import com.example.dto.PaymentResponse;
import com.example.service.PaymentService;
import javax.inject.Inject;
import javax.ws.rs.*;
import javax.ws.rs.core.MediaType;
import javax.ws.rs.core.Response;

@Path("/payments")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class PaymentResource {

    @Inject
    private PaymentService paymentService;

    @POST
    public Response processPayment(PaymentRequest request) {
        // VIOLATION: Missing @Valid annotation - input not validated
        PaymentResponse response = paymentService.processPayment(request);
        return Response.status(Response.Status.CREATED).entity(response).build();
    }

    @PUT
    @Path("/{id}")
    public Response updatePayment(@PathParam("id") Long id, PaymentRequest request) {
        // VIOLATION: Missing @Valid annotation
        PaymentResponse updated = paymentService.updatePayment(id, request);
        return Response.ok(updated).build();
    }
}
```

**Expected Violation:**
```
POST /payments: Input validation missing - request parameter lacks @Valid annotation
PUT /payments/{id}: Input validation missing - request parameter lacks @Valid annotation
```

---

## Micronaut — Perfect Implementation

**File:** `src/main/java/com/example/controller/UserController.java`

```java
package com.example.controller;

import com.example.dto.CreateUserRequest;
import com.example.dto.UserResponse;
import com.example.service.UserService;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import io.micronaut.http.annotation.*;
import javax.validation.Valid;
import java.util.List;

@Controller("/api/users")
public class UserController {

    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @Get
    public HttpResponse<List<UserResponse>> listUsers() {
        List<UserResponse> users = userService.getAllUsers();
        return HttpResponse.ok(users);
    }

    @Get("/{id}")
    public HttpResponse<UserResponse> getUser(@PathVariable Long id) {
        UserResponse user = userService.getUserById(id);
        if (user == null) {
            return HttpResponse.notFound();
        }
        return HttpResponse.ok(user);
    }

    @Post
    public HttpResponse<UserResponse> createUser(@Valid @Body CreateUserRequest request) {
        UserResponse created = userService.createUser(request);
        return HttpResponse.status(HttpStatus.CREATED).body(created);
    }

    @Put("/{id}")
    public HttpResponse<UserResponse> updateUser(
            @PathVariable Long id,
            @Valid @Body CreateUserRequest request) {
        UserResponse updated = userService.updateUser(id, request);
        if (updated == null) {
            return HttpResponse.notFound();
        }
        return HttpResponse.ok(updated);
    }

    @Delete("/{id}")
    public HttpResponse<Void> deleteUser(@PathVariable Long id) {
        boolean deleted = userService.deleteUser(id);
        if (!deleted) {
            return HttpResponse.notFound();
        }
        return HttpResponse.noContent();
    }
}
```

**Expected Behavior:**
- Stricture detects all 5 routes (GET /api/users, GET /api/users/{id}, POST /api/users, PUT /api/users/{id}, DELETE /api/users/{id})
- Recognizes @Valid for input validation
- Understands Micronaut HttpResponse status methods (ok, created, notFound, noContent)
- Maps @Controller + HTTP method annotations to routes
- NO violations reported

---

## Micronaut — Violation 1: Missing @Valid

**File:** `src/main/java/com/example/controller/ProductController.java`

```java
package com.example.controller;

import com.example.dto.CreateProductRequest;
import com.example.dto.ProductResponse;
import com.example.service.ProductService;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import io.micronaut.http.annotation.*;

@Controller("/api/products")
public class ProductController {

    private final ProductService productService;

    public ProductController(ProductService productService) {
        this.productService = productService;
    }

    @Post
    public HttpResponse<ProductResponse> createProduct(@Body CreateProductRequest request) {
        // VIOLATION: Missing @Valid annotation - input not validated
        ProductResponse created = productService.createProduct(request);
        return HttpResponse.status(HttpStatus.CREATED).body(created);
    }

    @Put("/{id}")
    public HttpResponse<ProductResponse> updateProduct(
            @PathVariable Long id,
            @Body CreateProductRequest request) {
        // VIOLATION: Missing @Valid annotation
        ProductResponse updated = productService.updateProduct(id, request);
        return HttpResponse.ok(updated);
    }
}
```

**Expected Violation:**
```
POST /api/products: Input validation missing - @Body lacks @Valid annotation
PUT /api/products/{id}: Input validation missing - @Body lacks @Valid annotation
```

---

## Micronaut — Violation 2: Wrong HttpResponse Status

**File:** `src/main/java/com/example/controller/OrderController.java`

```java
package com.example.controller;

import com.example.dto.CreateOrderRequest;
import com.example.dto.OrderResponse;
import com.example.service.OrderService;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.annotation.*;
import javax.validation.Valid;

@Controller("/api/orders")
public class OrderController {

    private final OrderService orderService;

    public OrderController(OrderService orderService) {
        this.orderService = orderService;
    }

    @Post
    public HttpResponse<OrderResponse> createOrder(@Valid @Body CreateOrderRequest request) {
        OrderResponse created = orderService.createOrder(request);
        // VIOLATION: Should return 201 Created, not 200 OK
        return HttpResponse.ok(created);
    }

    @Delete("/{id}")
    public HttpResponse<String> deleteOrder(@PathVariable Long id) {
        orderService.deleteOrder(id);
        // VIOLATION: DELETE should return 204 No Content, not 200 with body
        return HttpResponse.ok("Deleted");
    }

    @Get("/{id}")
    public HttpResponse<OrderResponse> getOrder(@PathVariable Long id) {
        OrderResponse order = orderService.getOrderById(id);
        if (order == null) {
            // VIOLATION: Should return 404 Not Found, not 204 No Content
            return HttpResponse.noContent();
        }
        return HttpResponse.ok(order);
    }
}
```

**Expected Violation:**
```
POST /api/orders: Should return 201 Created for resource creation, not 200 OK
DELETE /api/orders/{id}: Should return 204 No Content without body, not 200 OK
GET /api/orders/{id}: Should return 404 Not Found for missing resource, not 204 No Content
```

---

## Micronaut — Violation 3: Repository in Controller

**File:** `src/main/java/com/example/controller/PaymentController.java`

```java
package com.example.controller;

import com.example.dto.PaymentRequest;
import com.example.dto.PaymentResponse;
import com.example.repository.PaymentRepository;
import com.example.model.Payment;
import io.micronaut.http.HttpResponse;
import io.micronaut.http.HttpStatus;
import io.micronaut.http.annotation.*;
import javax.validation.Valid;

@Controller("/api/payments")
public class PaymentController {

    private final PaymentRepository paymentRepository;

    public PaymentController(PaymentRepository paymentRepository) {
        this.paymentRepository = paymentRepository;
    }

    @Post
    public HttpResponse<PaymentResponse> processPayment(@Valid @Body PaymentRequest request) {
        // VIOLATION: Controller directly accessing repository - missing service layer
        Payment payment = new Payment();
        payment.setAmount(request.getAmount());
        payment.setCardNumber(request.getCardNumber());
        Payment saved = paymentRepository.save(payment);
        return HttpResponse.status(HttpStatus.CREATED)
                .body(new PaymentResponse(saved.getId(), saved.getAmount()));
    }

    @Get("/{id}")
    public HttpResponse<PaymentResponse> getPayment(@PathVariable Long id) {
        // VIOLATION: Business logic in controller instead of service layer
        return paymentRepository.findById(id)
                .map(payment -> HttpResponse.ok(
                        new PaymentResponse(payment.getId(), payment.getAmount())))
                .orElse(HttpResponse.notFound());
    }

    @Delete("/{id}")
    public HttpResponse<Void> deletePayment(@PathVariable Long id) {
        // VIOLATION: Direct repository access without service layer
        paymentRepository.deleteById(id);
        return HttpResponse.noContent();
    }
}
```

**Expected Violation:**
```
POST /api/payments: Controller directly accessing repository - should use service layer
GET /api/payments/{id}: Controller directly accessing repository - should use service layer
DELETE /api/payments/{id}: Controller directly accessing repository - should use service layer
```

---

## Summary

**Total Test Cases:** 12
- Spring Boot: 4 (1 perfect + 3 violations)
- Jakarta EE (JAX-RS): 4 (1 perfect + 3 violations)
- Micronaut: 4 (1 perfect + 3 violations)

**Coverage:**
- Route detection across 3 major Java frameworks
- Input validation patterns (@Valid, @RequestBody, @Body)
- HTTP status code correctness (200, 201, 204, 404)
- Service layer architecture (controllers should not access repositories directly)
- Framework-specific annotations (@RestController, @Path, @Controller)
- Dependency injection patterns (@Autowired, @Inject, constructor injection)

**Expected Stricture Behavior:**
- Detect all routes with correct HTTP methods and paths
- Flag missing @Valid annotations on request bodies
- Flag incorrect HTTP status codes for CRUD operations
- Flag direct repository access in controllers
- Flag missing @Consumes/@Produces annotations in JAX-RS
- Understand framework-specific response builders (ResponseEntity, Response, HttpResponse)
