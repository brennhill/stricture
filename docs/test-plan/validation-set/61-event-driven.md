# 61 — Event-Driven Architecture (WebSocket + Message Queue)

**Category:** Event-Driven Architecture
**Complexity:** High
**Languages:** TypeScript (WebSocket server + client), Go (message queue producer), TypeScript (message queue consumer)
**Last Updated:** 2026-02-13

---

## Overview

Tests event-driven communication patterns using non-HTTP protocols:
- **Part 1:** WebSocket chat server with bidirectional messaging
- **Part 2:** Message queue producer/consumer with versioned events

**Key Validation Areas:**
- Message type enumeration and exhaustive handling
- Bidirectional WebSocket communication
- Event schema versioning and backward compatibility
- Consumer idempotency and offset tracking
- Error handling for connection failures
- Authentication and authorization for WebSocket connections

---

## Part 1: WebSocket Chat Server

### Perfect Implementation Requirements

**Server (TypeScript):**
```typescript
// ws-chat-server.ts
import { WebSocketServer, WebSocket } from 'ws';

// Message type enumeration
enum MessageType {
  JOIN = 'join',
  MESSAGE = 'message',
  LEAVE = 'leave',
  TYPING = 'typing',
  ERROR = 'error',
  PING = 'ping',
  PONG = 'pong'
}

// Message interfaces
interface BaseMessage {
  type: MessageType;
  timestamp: number;
  user_id: string;
}

interface JoinMessage extends BaseMessage {
  type: MessageType.JOIN;
  username: string;
  room_id: string;
}

interface ChatMessage extends BaseMessage {
  type: MessageType.MESSAGE;
  content: string;
  room_id: string;
}

interface LeaveMessage extends BaseMessage {
  type: MessageType.LEAVE;
  room_id: string;
}

interface TypingMessage extends BaseMessage {
  type: MessageType.TYPING;
  room_id: string;
  is_typing: boolean;
}

interface ErrorMessage extends BaseMessage {
  type: MessageType.ERROR;
  error_code: string;
  error_message: string;
}

type IncomingMessage = JoinMessage | ChatMessage | LeaveMessage | TypingMessage;
type OutgoingMessage = IncomingMessage | ErrorMessage;

// Client connection tracking
interface Client {
  ws: WebSocket;
  user_id: string;
  username: string;
  rooms: Set<string>;
  authenticated: boolean;
}

class ChatServer {
  private wss: WebSocketServer;
  private clients: Map<WebSocket, Client> = new Map();
  private rooms: Map<string, Set<WebSocket>> = new Map();

  constructor(port: number) {
    this.wss = new WebSocketServer({ port });
    this.wss.on('connection', this.handleConnection.bind(this));
  }

  private handleConnection(ws: WebSocket): void {
    console.log('New connection');

    // Initialize client (not authenticated yet)
    const client: Client = {
      ws,
      user_id: '',
      username: '',
      rooms: new Set(),
      authenticated: false
    };
    this.clients.set(ws, client);

    // Set up event handlers
    ws.on('message', (data: Buffer) => this.handleMessage(ws, data));
    ws.on('error', (error: Error) => this.handleError(ws, error));
    ws.on('close', () => this.handleClose(ws));
    ws.on('ping', () => ws.pong());
  }

  private handleMessage(ws: WebSocket, data: Buffer): void {
    const client = this.clients.get(ws);
    if (!client) {
      return;
    }

    let message: IncomingMessage;
    try {
      message = JSON.parse(data.toString());
    } catch {
      this.sendError(ws, 'INVALID_JSON', 'Failed to parse message');
      return;
    }

    // Validate message structure
    if (!message.type || !message.user_id) {
      this.sendError(ws, 'MISSING_FIELDS', 'Message must have type and user_id');
      return;
    }

    // Authentication check (except for JOIN)
    if (message.type !== MessageType.JOIN && !client.authenticated) {
      this.sendError(ws, 'UNAUTHORIZED', 'Must join before sending messages');
      return;
    }

    // Exhaustive message type handling
    switch (message.type) {
      case MessageType.JOIN:
        this.handleJoin(ws, message as JoinMessage);
        break;
      case MessageType.MESSAGE:
        this.handleChatMessage(ws, message as ChatMessage);
        break;
      case MessageType.LEAVE:
        this.handleLeave(ws, message as LeaveMessage);
        break;
      case MessageType.TYPING:
        this.handleTyping(ws, message as TypingMessage);
        break;
      case MessageType.PING:
        this.handlePing(ws);
        break;
      default:
        // Exhaustiveness check - should never reach here
        const _exhaustive: never = message.type;
        this.sendError(ws, 'UNKNOWN_TYPE', `Unknown message type: ${message.type}`);
    }
  }

  private handleJoin(ws: WebSocket, message: JoinMessage): void {
    const client = this.clients.get(ws);
    if (!client) return;

    // Validate JOIN message fields
    if (!message.username || !message.room_id) {
      this.sendError(ws, 'INVALID_JOIN', 'JOIN requires username and room_id');
      return;
    }

    // Authenticate client
    client.user_id = message.user_id;
    client.username = message.username;
    client.authenticated = true;
    client.rooms.add(message.room_id);

    // Add to room
    if (!this.rooms.has(message.room_id)) {
      this.rooms.set(message.room_id, new Set());
    }
    this.rooms.get(message.room_id)!.add(ws);

    // Broadcast join to room
    this.broadcastToRoom(message.room_id, {
      type: MessageType.JOIN,
      user_id: message.user_id,
      username: message.username,
      room_id: message.room_id,
      timestamp: Date.now()
    }, ws);

    console.log(`User ${message.username} joined room ${message.room_id}`);
  }

  private handleChatMessage(ws: WebSocket, message: ChatMessage): void {
    const client = this.clients.get(ws);
    if (!client) return;

    // Validate room membership
    if (!client.rooms.has(message.room_id)) {
      this.sendError(ws, 'NOT_IN_ROOM', `Not a member of room ${message.room_id}`);
      return;
    }

    // Validate content
    if (!message.content || message.content.trim().length === 0) {
      this.sendError(ws, 'EMPTY_MESSAGE', 'Message content cannot be empty');
      return;
    }

    // Broadcast to room (including sender)
    this.broadcastToRoom(message.room_id, {
      type: MessageType.MESSAGE,
      user_id: client.user_id,
      content: message.content,
      room_id: message.room_id,
      timestamp: Date.now()
    });
  }

  private handleLeave(ws: WebSocket, message: LeaveMessage): void {
    const client = this.clients.get(ws);
    if (!client) return;

    // Remove from room
    client.rooms.delete(message.room_id);
    const room = this.rooms.get(message.room_id);
    if (room) {
      room.delete(ws);
      if (room.size === 0) {
        this.rooms.delete(message.room_id);
      }
    }

    // Broadcast leave to room
    this.broadcastToRoom(message.room_id, {
      type: MessageType.LEAVE,
      user_id: client.user_id,
      room_id: message.room_id,
      timestamp: Date.now()
    }, ws);

    console.log(`User ${client.username} left room ${message.room_id}`);
  }

  private handleTyping(ws: WebSocket, message: TypingMessage): void {
    const client = this.clients.get(ws);
    if (!client) return;

    // Validate room membership
    if (!client.rooms.has(message.room_id)) {
      return;
    }

    // Broadcast typing status to room (excluding sender)
    this.broadcastToRoom(message.room_id, {
      type: MessageType.TYPING,
      user_id: client.user_id,
      room_id: message.room_id,
      is_typing: message.is_typing,
      timestamp: Date.now()
    }, ws);
  }

  private handlePing(ws: WebSocket): void {
    ws.send(JSON.stringify({
      type: MessageType.PONG,
      timestamp: Date.now()
    }));
  }

  private handleError(ws: WebSocket, error: Error): void {
    console.error('WebSocket error:', error);
    this.sendError(ws, 'CONNECTION_ERROR', error.message);
  }

  private handleClose(ws: WebSocket): void {
    const client = this.clients.get(ws);
    if (!client) return;

    // Remove from all rooms
    for (const room_id of client.rooms) {
      const room = this.rooms.get(room_id);
      if (room) {
        room.delete(ws);
        if (room.size === 0) {
          this.rooms.delete(room_id);
        }

        // Broadcast leave
        this.broadcastToRoom(room_id, {
          type: MessageType.LEAVE,
          user_id: client.user_id,
          room_id,
          timestamp: Date.now()
        });
      }
    }

    this.clients.delete(ws);
    console.log(`Connection closed for user ${client.username}`);
  }

  private broadcastToRoom(room_id: string, message: OutgoingMessage, exclude?: WebSocket): void {
    const room = this.rooms.get(room_id);
    if (!room) return;

    const payload = JSON.stringify(message);
    for (const client_ws of room) {
      if (client_ws !== exclude && client_ws.readyState === WebSocket.OPEN) {
        client_ws.send(payload);
      }
    }
  }

  private sendError(ws: WebSocket, error_code: string, error_message: string): void {
    const message: ErrorMessage = {
      type: MessageType.ERROR,
      user_id: '',
      error_code,
      error_message,
      timestamp: Date.now()
    };
    ws.send(JSON.stringify(message));
  }
}

// Start server
const server = new ChatServer(8080);
console.log('WebSocket chat server running on ws://localhost:8080');
```

**Client (TypeScript):**
```typescript
// ws-chat-client.ts
import WebSocket from 'ws';

class ChatClient {
  private ws: WebSocket | null = null;
  private user_id: string;
  private username: string;
  private connected: boolean = false;
  private reconnect_timeout: NodeJS.Timeout | null = null;

  constructor(user_id: string, username: string) {
    this.user_id = user_id;
    this.username = username;
  }

  connect(url: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(url);

      this.ws.on('open', () => {
        console.log('Connected to server');
        this.connected = true;
        resolve();
      });

      this.ws.on('message', (data: Buffer) => {
        this.handleMessage(data);
      });

      this.ws.on('error', (error: Error) => {
        console.error('WebSocket error:', error);
        reject(error);
      });

      this.ws.on('close', () => {
        console.log('Disconnected from server');
        this.connected = false;
        this.attemptReconnect(url);
      });
    });
  }

  private attemptReconnect(url: string): void {
    if (this.reconnect_timeout) {
      return;
    }

    console.log('Attempting reconnect in 5s...');
    this.reconnect_timeout = setTimeout(() => {
      this.reconnect_timeout = null;
      this.connect(url).catch(() => {
        this.attemptReconnect(url);
      });
    }, 5000);
  }

  private handleMessage(data: Buffer): void {
    let message: any;
    try {
      message = JSON.parse(data.toString());
    } catch {
      console.error('Failed to parse message');
      return;
    }

    // Exhaustive message type handling
    switch (message.type) {
      case 'join':
        console.log(`${message.username} joined ${message.room_id}`);
        break;
      case 'message':
        console.log(`[${message.room_id}] ${message.user_id}: ${message.content}`);
        break;
      case 'leave':
        console.log(`${message.user_id} left ${message.room_id}`);
        break;
      case 'typing':
        if (message.is_typing) {
          console.log(`${message.user_id} is typing...`);
        }
        break;
      case 'error':
        console.error(`Error [${message.error_code}]: ${message.error_message}`);
        break;
      case 'pong':
        console.log('Received pong');
        break;
      default:
        console.warn(`Unknown message type: ${message.type}`);
    }
  }

  joinRoom(room_id: string): void {
    if (!this.ws || !this.connected) {
      console.error('Not connected');
      return;
    }

    this.ws.send(JSON.stringify({
      type: 'join',
      user_id: this.user_id,
      username: this.username,
      room_id,
      timestamp: Date.now()
    }));
  }

  sendMessage(room_id: string, content: string): void {
    if (!this.ws || !this.connected) {
      console.error('Not connected');
      return;
    }

    this.ws.send(JSON.stringify({
      type: 'message',
      user_id: this.user_id,
      room_id,
      content,
      timestamp: Date.now()
    }));
  }

  setTyping(room_id: string, is_typing: boolean): void {
    if (!this.ws || !this.connected) {
      return;
    }

    this.ws.send(JSON.stringify({
      type: 'typing',
      user_id: this.user_id,
      room_id,
      is_typing,
      timestamp: Date.now()
    }));
  }

  leaveRoom(room_id: string): void {
    if (!this.ws || !this.connected) {
      return;
    }

    this.ws.send(JSON.stringify({
      type: 'leave',
      user_id: this.user_id,
      room_id,
      timestamp: Date.now()
    }));
  }

  disconnect(): void {
    if (this.reconnect_timeout) {
      clearTimeout(this.reconnect_timeout);
      this.reconnect_timeout = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

// Example usage
async function main() {
  const client = new ChatClient('user123', 'Alice');

  try {
    await client.connect('ws://localhost:8080');
    client.joinRoom('general');
    client.sendMessage('general', 'Hello, world!');

    setTimeout(() => {
      client.setTyping('general', true);
      setTimeout(() => {
        client.sendMessage('general', 'How is everyone?');
        client.setTyping('general', false);
      }, 2000);
    }, 3000);
  } catch (error) {
    console.error('Failed to connect:', error);
  }
}

main();
```

---

## Part 2: Message Queue (Producer/Consumer)

### Perfect Implementation Requirements

**Producer (Go):**
```go
// mq-producer.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Event type enumeration
type EventType string

const (
	OrderCreated   EventType = "order.created"
	OrderUpdated   EventType = "order.updated"
	OrderCancelled EventType = "order.cancelled"
)

// Event envelope with versioning
type EventEnvelope struct {
	EventID      string                 `json:"event_id"`
	EventType    EventType              `json:"event_type"`
	EventVersion int                    `json:"event_version"`
	Timestamp    int64                  `json:"timestamp"`
	CorrelationID string                `json:"correlation_id,omitempty"`
	Data         map[string]interface{} `json:"data"`
}

// Order domain events
type OrderCreatedData struct {
	OrderID      string  `json:"order_id"`
	CustomerID   string  `json:"customer_id"`
	TotalAmount  float64 `json:"total_amount"`
	Currency     string  `json:"currency"`
	ItemCount    int     `json:"item_count"`
}

type OrderUpdatedData struct {
	OrderID     string                 `json:"order_id"`
	UpdatedAt   int64                  `json:"updated_at"`
	Changes     map[string]interface{} `json:"changes"`
}

type OrderCancelledData struct {
	OrderID       string `json:"order_id"`
	CancelledBy   string `json:"cancelled_by"`
	CancelReason  string `json:"cancel_reason"`
	RefundAmount  float64 `json:"refund_amount"`
}

// Message queue interface (simulated with channel)
type MessageQueue struct {
	events chan EventEnvelope
}

func NewMessageQueue(bufferSize int) *MessageQueue {
	return &MessageQueue{
		events: make(chan EventEnvelope, bufferSize),
	}
}

func (mq *MessageQueue) Publish(event EventEnvelope) error {
	select {
	case mq.events <- event:
		return nil
	default:
		return fmt.Errorf("queue full")
	}
}

func (mq *MessageQueue) Subscribe() <-chan EventEnvelope {
	return mq.events
}

// Event producer
type OrderEventProducer struct {
	queue *MessageQueue
}

func NewOrderEventProducer(queue *MessageQueue) *OrderEventProducer {
	return &OrderEventProducer{queue: queue}
}

func (p *OrderEventProducer) PublishOrderCreated(data OrderCreatedData, correlationID string) error {
	// Validate event data
	if data.OrderID == "" {
		return fmt.Errorf("order_id required")
	}
	if data.CustomerID == "" {
		return fmt.Errorf("customer_id required")
	}
	if data.TotalAmount <= 0 {
		return fmt.Errorf("total_amount must be positive")
	}

	// Convert to map for envelope
	dataMap := map[string]interface{}{
		"order_id":     data.OrderID,
		"customer_id":  data.CustomerID,
		"total_amount": data.TotalAmount,
		"currency":     data.Currency,
		"item_count":   data.ItemCount,
	}

	envelope := EventEnvelope{
		EventID:       generateEventID(),
		EventType:     OrderCreated,
		EventVersion:  1,
		Timestamp:     time.Now().UnixMilli(),
		CorrelationID: correlationID,
		Data:          dataMap,
	}

	return p.queue.Publish(envelope)
}

func (p *OrderEventProducer) PublishOrderUpdated(data OrderUpdatedData, correlationID string) error {
	if data.OrderID == "" {
		return fmt.Errorf("order_id required")
	}
	if data.Changes == nil || len(data.Changes) == 0 {
		return fmt.Errorf("changes required")
	}

	dataMap := map[string]interface{}{
		"order_id":   data.OrderID,
		"updated_at": data.UpdatedAt,
		"changes":    data.Changes,
	}

	envelope := EventEnvelope{
		EventID:       generateEventID(),
		EventType:     OrderUpdated,
		EventVersion:  1,
		Timestamp:     time.Now().UnixMilli(),
		CorrelationID: correlationID,
		Data:          dataMap,
	}

	return p.queue.Publish(envelope)
}

func (p *OrderEventProducer) PublishOrderCancelled(data OrderCancelledData, correlationID string) error {
	if data.OrderID == "" {
		return fmt.Errorf("order_id required")
	}
	if data.CancelledBy == "" {
		return fmt.Errorf("cancelled_by required")
	}

	dataMap := map[string]interface{}{
		"order_id":       data.OrderID,
		"cancelled_by":   data.CancelledBy,
		"cancel_reason":  data.CancelReason,
		"refund_amount":  data.RefundAmount,
	}

	envelope := EventEnvelope{
		EventID:       generateEventID(),
		EventType:     OrderCancelled,
		EventVersion:  1,
		Timestamp:     time.Now().UnixMilli(),
		CorrelationID: correlationID,
		Data:          dataMap,
	}

	return p.queue.Publish(envelope)
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func main() {
	// Create message queue
	mq := NewMessageQueue(100)
	producer := NewOrderEventProducer(mq)

	// Start consumer (in goroutine to simulate separate process)
	go func() {
		for event := range mq.Subscribe() {
			data, _ := json.MarshalIndent(event, "", "  ")
			fmt.Printf("Produced event:\n%s\n\n", data)
		}
	}()

	// Produce events
	correlationID := "corr_123"

	// Event 1: Order created
	err := producer.PublishOrderCreated(OrderCreatedData{
		OrderID:     "order_001",
		CustomerID:  "cust_456",
		TotalAmount: 99.99,
		Currency:    "USD",
		ItemCount:   3,
	}, correlationID)
	if err != nil {
		log.Fatalf("Failed to publish order created: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Event 2: Order updated
	err = producer.PublishOrderUpdated(OrderUpdatedData{
		OrderID:   "order_001",
		UpdatedAt: time.Now().UnixMilli(),
		Changes: map[string]interface{}{
			"shipping_address": map[string]string{
				"street": "123 Main St",
				"city":   "Seattle",
				"state":  "WA",
			},
		},
	}, correlationID)
	if err != nil {
		log.Fatalf("Failed to publish order updated: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Event 3: Order cancelled
	err = producer.PublishOrderCancelled(OrderCancelledData{
		OrderID:      "order_001",
		CancelledBy:  "cust_456",
		CancelReason: "Customer requested cancellation",
		RefundAmount: 99.99,
	}, correlationID)
	if err != nil {
		log.Fatalf("Failed to publish order cancelled: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
}
```

**Consumer (TypeScript):**
```typescript
// mq-consumer.ts
import * as fs from 'fs';

// Event type enumeration
enum EventType {
  ORDER_CREATED = 'order.created',
  ORDER_UPDATED = 'order.updated',
  ORDER_CANCELLED = 'order.cancelled'
}

// Event envelope interface
interface EventEnvelope {
  event_id: string;
  event_type: EventType;
  event_version: number;
  timestamp: number;
  correlation_id?: string;
  data: Record<string, any>;
}

// Offset tracking for idempotency
interface ConsumerOffset {
  last_event_id: string;
  last_timestamp: number;
  processed_count: number;
}

class OrderEventConsumer {
  private processed_events: Set<string> = new Set();
  private offset: ConsumerOffset = {
    last_event_id: '',
    last_timestamp: 0,
    processed_count: 0
  };
  private offset_file: string;

  constructor(offset_file: string) {
    this.offset_file = offset_file;
    this.loadOffset();
  }

  // Load offset from disk for crash recovery
  private loadOffset(): void {
    try {
      if (fs.existsSync(this.offset_file)) {
        const data = fs.readFileSync(this.offset_file, 'utf-8');
        this.offset = JSON.parse(data);
        console.log(`Loaded offset: ${this.offset.processed_count} events processed`);
      }
    } catch (error) {
      console.error('Failed to load offset:', error);
    }
  }

  // Save offset to disk
  private saveOffset(): void {
    try {
      fs.writeFileSync(this.offset_file, JSON.stringify(this.offset, null, 2));
    } catch (error) {
      console.error('Failed to save offset:', error);
    }
  }

  // Idempotent event processing
  async processEvent(event: EventEnvelope): Promise<void> {
    // Check if already processed (idempotency)
    if (this.processed_events.has(event.event_id)) {
      console.log(`Skipping duplicate event: ${event.event_id}`);
      return;
    }

    // Validate event structure
    if (!event.event_type || !event.event_id || event.event_version === undefined) {
      console.error('Invalid event structure:', event);
      return;
    }

    // Validate timestamp ordering (detect out-of-order delivery)
    if (event.timestamp < this.offset.last_timestamp) {
      console.warn(`Out-of-order event detected: ${event.event_id} (${event.timestamp} < ${this.offset.last_timestamp})`);
    }

    // Exhaustive event type handling
    switch (event.event_type) {
      case EventType.ORDER_CREATED:
        await this.handleOrderCreated(event);
        break;
      case EventType.ORDER_UPDATED:
        await this.handleOrderUpdated(event);
        break;
      case EventType.ORDER_CANCELLED:
        await this.handleOrderCancelled(event);
        break;
      default:
        // Exhaustiveness check
        const _exhaustive: never = event.event_type;
        console.error(`Unknown event type: ${event.event_type}`);
        return;
    }

    // Mark as processed
    this.processed_events.add(event.event_id);
    this.offset.last_event_id = event.event_id;
    this.offset.last_timestamp = event.timestamp;
    this.offset.processed_count++;

    // Persist offset (for crash recovery)
    this.saveOffset();
  }

  private async handleOrderCreated(event: EventEnvelope): Promise<void> {
    const { order_id, customer_id, total_amount, currency, item_count } = event.data;

    // Validate required fields
    if (!order_id || !customer_id || total_amount === undefined) {
      console.error('Missing required fields in order.created event:', event);
      return;
    }

    console.log(`[ORDER CREATED] Order ${order_id} for customer ${customer_id}`);
    console.log(`  Amount: ${total_amount} ${currency}`);
    console.log(`  Items: ${item_count}`);
    console.log(`  Correlation: ${event.correlation_id}`);

    // Simulate business logic
    await this.createOrderInDatabase(event.data);
  }

  private async handleOrderUpdated(event: EventEnvelope): Promise<void> {
    const { order_id, updated_at, changes } = event.data;

    if (!order_id || !changes) {
      console.error('Missing required fields in order.updated event:', event);
      return;
    }

    console.log(`[ORDER UPDATED] Order ${order_id}`);
    console.log(`  Changes:`, changes);
    console.log(`  Updated at: ${new Date(updated_at).toISOString()}`);

    // Simulate business logic
    await this.updateOrderInDatabase(event.data);
  }

  private async handleOrderCancelled(event: EventEnvelope): Promise<void> {
    const { order_id, cancelled_by, cancel_reason, refund_amount } = event.data;

    if (!order_id || !cancelled_by) {
      console.error('Missing required fields in order.cancelled event:', event);
      return;
    }

    console.log(`[ORDER CANCELLED] Order ${order_id}`);
    console.log(`  Cancelled by: ${cancelled_by}`);
    console.log(`  Reason: ${cancel_reason}`);
    console.log(`  Refund: ${refund_amount}`);

    // Simulate business logic
    await this.cancelOrderInDatabase(event.data);
  }

  // Simulated database operations
  private async createOrderInDatabase(data: Record<string, any>): Promise<void> {
    // Simulate async DB operation
    await new Promise(resolve => setTimeout(resolve, 50));
    console.log(`  -> Database: Order ${data.order_id} created`);
  }

  private async updateOrderInDatabase(data: Record<string, any>): Promise<void> {
    await new Promise(resolve => setTimeout(resolve, 50));
    console.log(`  -> Database: Order ${data.order_id} updated`);
  }

  private async cancelOrderInDatabase(data: Record<string, any>): Promise<void> {
    await new Promise(resolve => setTimeout(resolve, 50));
    console.log(`  -> Database: Order ${data.order_id} cancelled`);
  }

  // Get consumer statistics
  getStats(): ConsumerOffset {
    return { ...this.offset };
  }
}

// Simulate message queue consumption
async function main() {
  const consumer = new OrderEventConsumer('./consumer-offset.json');

  // Simulated events (in real system, would come from queue)
  const events: EventEnvelope[] = [
    {
      event_id: 'evt_1',
      event_type: EventType.ORDER_CREATED,
      event_version: 1,
      timestamp: Date.now(),
      correlation_id: 'corr_123',
      data: {
        order_id: 'order_001',
        customer_id: 'cust_456',
        total_amount: 99.99,
        currency: 'USD',
        item_count: 3
      }
    },
    {
      event_id: 'evt_2',
      event_type: EventType.ORDER_UPDATED,
      event_version: 1,
      timestamp: Date.now() + 100,
      correlation_id: 'corr_123',
      data: {
        order_id: 'order_001',
        updated_at: Date.now(),
        changes: {
          shipping_address: {
            street: '123 Main St',
            city: 'Seattle',
            state: 'WA'
          }
        }
      }
    },
    {
      event_id: 'evt_3',
      event_type: EventType.ORDER_CANCELLED,
      event_version: 1,
      timestamp: Date.now() + 200,
      correlation_id: 'corr_123',
      data: {
        order_id: 'order_001',
        cancelled_by: 'cust_456',
        cancel_reason: 'Customer requested cancellation',
        refund_amount: 99.99
      }
    },
    // Duplicate event (should be skipped)
    {
      event_id: 'evt_1',
      event_type: EventType.ORDER_CREATED,
      event_version: 1,
      timestamp: Date.now() + 300,
      correlation_id: 'corr_123',
      data: {
        order_id: 'order_001',
        customer_id: 'cust_456',
        total_amount: 99.99,
        currency: 'USD',
        item_count: 3
      }
    }
  ];

  // Process events
  for (const event of events) {
    await consumer.processEvent(event);
    console.log('---');
  }

  // Print statistics
  const stats = consumer.getStats();
  console.log('\nConsumer Statistics:');
  console.log(`  Total processed: ${stats.processed_count}`);
  console.log(`  Last event ID: ${stats.last_event_id}`);
  console.log(`  Last timestamp: ${new Date(stats.last_timestamp).toISOString()}`);
}

main();
```

---

## Broken Implementations (B01-B15)

### B01: No WebSocket Error Handler

**Bug:** Missing `ws.on('error')` handler in server

**Impact:** Unhandled errors crash the server

**Location:** `ws-chat-server.ts` - handleConnection method

```typescript
// BROKEN: No error handler
private handleConnection(ws: WebSocket): void {
  const client: Client = {
    ws,
    user_id: '',
    username: '',
    rooms: new Set(),
    authenticated: false
  };
  this.clients.set(ws, client);

  ws.on('message', (data: Buffer) => this.handleMessage(ws, data));
  // MISSING: ws.on('error', ...)
  ws.on('close', () => this.handleClose(ws));
}
```

**Test Case:**
- Trigger connection error (e.g., send malformed frame)
- Verify server crashes with uncaught exception
- Perfect implementation should log error and send ERROR message to client

---

### B02: No Message Type Validation

**Bug:** No validation that message.type is a valid MessageType enum value

**Impact:** Invalid message types not caught, switch statement default never reached

**Location:** `ws-chat-server.ts` - handleMessage method

```typescript
// BROKEN: No type validation
private handleMessage(ws: WebSocket, data: Buffer): void {
  let message: IncomingMessage;
  try {
    message = JSON.parse(data.toString());
  } catch {
    this.sendError(ws, 'INVALID_JSON', 'Failed to parse message');
    return;
  }

  // MISSING: Check if message.type is valid MessageType
  // message.type could be "invalid_type" and won't be caught

  switch (message.type) {
    case MessageType.JOIN:
      this.handleJoin(ws, message as JoinMessage);
      break;
    // ...
  }
}
```

**Test Case:**
- Send message with `type: "invalid_type"`
- Verify no error is returned (bug - should return ERROR message)
- Perfect implementation validates against MessageType enum

---

### B03: No Client Authentication Check

**Bug:** Missing authentication check before allowing MESSAGE/LEAVE/TYPING

**Impact:** Unauthenticated clients can send messages

**Location:** `ws-chat-server.ts` - handleMessage method

```typescript
// BROKEN: No auth check
private handleMessage(ws: WebSocket, data: Buffer): void {
  // ... parse message ...

  // MISSING: Authentication check
  // if (message.type !== MessageType.JOIN && !client.authenticated) {
  //   this.sendError(ws, 'UNAUTHORIZED', 'Must join before sending messages');
  //   return;
  // }

  switch (message.type) {
    case MessageType.MESSAGE:
      this.handleChatMessage(ws, message as ChatMessage);
      break;
    // ...
  }
}
```

**Test Case:**
- Connect client but don't send JOIN message
- Send MESSAGE directly
- Verify message is broadcasted (bug - should be rejected)

---

### B04: No Room Membership Validation

**Bug:** handleChatMessage doesn't check if client is in the room

**Impact:** Users can send messages to rooms they haven't joined

**Location:** `ws-chat-server.ts` - handleChatMessage method

```typescript
// BROKEN: No room membership check
private handleChatMessage(ws: WebSocket, message: ChatMessage): void {
  const client = this.clients.get(ws);
  if (!client) return;

  // MISSING: Room membership check
  // if (!client.rooms.has(message.room_id)) {
  //   this.sendError(ws, 'NOT_IN_ROOM', `Not a member of room ${message.room_id}`);
  //   return;
  // }

  this.broadcastToRoom(message.room_id, {
    type: MessageType.MESSAGE,
    user_id: client.user_id,
    content: message.content,
    room_id: message.room_id,
    timestamp: Date.now()
  });
}
```

**Test Case:**
- Client joins "room1"
- Client sends message to "room2" (without joining)
- Verify message is broadcasted (bug - should be rejected)

---

### B05: Inconsistent Message Type String

**Bug:** Client uses "msg" instead of "message" for MESSAGE type

**Impact:** Messages not handled correctly due to type mismatch

**Location:** `ws-chat-client.ts` - sendMessage method

```typescript
// BROKEN: Wrong type string
sendMessage(room_id: string, content: string): void {
  this.ws.send(JSON.stringify({
    type: 'msg',  // WRONG: Should be 'message'
    user_id: this.user_id,
    room_id,
    content,
    timestamp: Date.now()
  }));
}
```

**Test Case:**
- Client sends message with type "msg"
- Server receives it and falls through to default case
- Verify ERROR message returned (type validation working)

---

### B06: No Empty Message Content Validation

**Bug:** handleChatMessage doesn't validate content is non-empty

**Impact:** Empty messages broadcasted to room

**Location:** `ws-chat-server.ts` - handleChatMessage method

```typescript
// BROKEN: No content validation
private handleChatMessage(ws: WebSocket, message: ChatMessage): void {
  const client = this.clients.get(ws);
  if (!client) return;

  // MISSING: Content validation
  // if (!message.content || message.content.trim().length === 0) {
  //   this.sendError(ws, 'EMPTY_MESSAGE', 'Message content cannot be empty');
  //   return;
  // }

  this.broadcastToRoom(message.room_id, {
    type: MessageType.MESSAGE,
    user_id: client.user_id,
    content: message.content,
    room_id: message.room_id,
    timestamp: Date.now()
  });
}
```

**Test Case:**
- Send message with empty string content
- Verify empty message is broadcasted (bug - should be rejected)

---

### B07: Client Missing Reconnect Logic

**Bug:** No reconnection attempt on disconnect

**Impact:** Client stays disconnected after connection loss

**Location:** `ws-chat-client.ts` - close event handler

```typescript
// BROKEN: No reconnect
this.ws.on('close', () => {
  console.log('Disconnected from server');
  this.connected = false;
  // MISSING: Reconnection logic
  // this.attemptReconnect(url);
});
```

**Test Case:**
- Disconnect client (kill server or close connection)
- Wait 10 seconds
- Verify client doesn't reconnect (bug - should attempt reconnect)

---

### B08: Consumer Handles Only 1/3 Event Types

**Bug:** Consumer only implements handleOrderCreated, missing other 2 handlers

**Impact:** 2/3 of events are silently ignored

**Location:** `mq-consumer.ts` - processEvent method

```typescript
// BROKEN: Incomplete event handling
switch (event.event_type) {
  case EventType.ORDER_CREATED:
    await this.handleOrderCreated(event);
    break;
  // MISSING: ORDER_UPDATED and ORDER_CANCELLED cases
  default:
    console.error(`Unknown event type: ${event.event_type}`);
    return;
}
```

**Test Case:**
- Publish order.updated event
- Verify it falls through to default case and is logged as "unknown"
- Perfect implementation handles all 3 event types

---

### B09: No Event Version Handling

**Bug:** Consumer doesn't check event_version field

**Impact:** Can't handle schema changes or backward compatibility

**Location:** `mq-consumer.ts` - processEvent and handler methods

```typescript
// BROKEN: No version check
async processEvent(event: EventEnvelope): Promise<void> {
  // MISSING: Version check
  // if (event.event_version > MAX_SUPPORTED_VERSION) {
  //   console.error(`Unsupported event version: ${event.event_version}`);
  //   return;
  // }

  switch (event.event_type) {
    // ...
  }
}
```

**Test Case:**
- Publish event with event_version: 999
- Verify consumer processes it anyway (bug - should reject or handle gracefully)

---

### B10: No Required Field Validation in Handlers

**Bug:** handleOrderCreated doesn't validate required fields

**Impact:** Incomplete data processed, causing downstream errors

**Location:** `mq-consumer.ts` - handleOrderCreated method

```typescript
// BROKEN: No field validation
private async handleOrderCreated(event: EventEnvelope): Promise<void> {
  const { order_id, customer_id, total_amount } = event.data;

  // MISSING: Required field validation
  // if (!order_id || !customer_id || total_amount === undefined) {
  //   console.error('Missing required fields in order.created event:', event);
  //   return;
  // }

  console.log(`[ORDER CREATED] Order ${order_id}`);
  await this.createOrderInDatabase(event.data);
}
```

**Test Case:**
- Publish order.created event missing customer_id
- Verify it's processed anyway (bug - should validate and reject)

---

### B11: No Producer Event Validation

**Bug:** Producer doesn't validate data before publishing

**Impact:** Invalid events published to queue

**Location:** `mq-producer.go` - PublishOrderCreated method

```go
// BROKEN: No validation
func (p *OrderEventProducer) PublishOrderCreated(data OrderCreatedData, correlationID string) error {
	// MISSING: Data validation
	// if data.OrderID == "" {
	// 	return fmt.Errorf("order_id required")
	// }
	// if data.TotalAmount <= 0 {
	// 	return fmt.Errorf("total_amount must be positive")
	// }

	envelope := EventEnvelope{
		EventID:       generateEventID(),
		EventType:     OrderCreated,
		EventVersion:  1,
		Timestamp:     time.Now().UnixMilli(),
		CorrelationID: correlationID,
		Data:          convertToMap(data),
	}

	return p.queue.Publish(envelope)
}
```

**Test Case:**
- Call PublishOrderCreated with empty OrderID
- Verify event is published (bug - should return validation error)

---

### B12: Queue Full Error Not Handled

**Bug:** Producer doesn't handle queue full condition

**Impact:** Events silently dropped when queue is full

**Location:** `mq-producer.go` - MessageQueue.Publish method

```go
// BROKEN: Silent failure
func (mq *MessageQueue) Publish(event EventEnvelope) error {
	select {
	case mq.events <- event:
		return nil
	default:
		// Returns error but callers don't check it
		return fmt.Errorf("queue full")
	}
}

// BROKEN: Caller doesn't check error
err := producer.PublishOrderCreated(data, correlationID)
// MISSING: if err != nil { ... }
```

**Test Case:**
- Fill queue to capacity (100 events)
- Publish 101st event
- Verify error is returned but not logged (silent failure)

---

### B13: No WebSocket Authentication

**Bug:** Server accepts all connections without auth token validation

**Impact:** Anyone can connect and spam the server

**Location:** `ws-chat-server.ts` - handleConnection method

```typescript
// BROKEN: No auth token check
private handleConnection(ws: WebSocket): void {
  // MISSING: Auth token validation
  // const url = new URL(req.url, 'http://localhost');
  // const token = url.searchParams.get('token');
  // if (!isValidToken(token)) {
  //   ws.close(1008, 'Unauthorized');
  //   return;
  // }

  const client: Client = {
    ws,
    user_id: '',
    username: '',
    rooms: new Set(),
    authenticated: false
  };
  this.clients.set(ws, client);
  // ...
}
```

**Test Case:**
- Connect without auth token
- Verify connection accepted (bug - should be rejected)

---

### B14: No Consumer Offset Tracking

**Bug:** Consumer doesn't save/load offset for crash recovery

**Impact:** Events reprocessed after consumer restart

**Location:** `mq-consumer.ts` - constructor and processEvent

```typescript
// BROKEN: No offset persistence
class OrderEventConsumer {
  private processed_events: Set<string> = new Set();
  // MISSING: Offset file loading/saving

  constructor() {
    // MISSING: this.loadOffset();
  }

  async processEvent(event: EventEnvelope): Promise<void> {
    // ... process event ...

    this.processed_events.add(event.event_id);
    // MISSING: this.saveOffset();
  }
}
```

**Test Case:**
- Process 10 events
- Restart consumer
- Verify all 10 events are reprocessed (bug - should resume from offset)

---

### B15: No Idempotency Check

**Bug:** Consumer doesn't check if event was already processed

**Impact:** Duplicate events processed multiple times

**Location:** `mq-consumer.ts` - processEvent method

```typescript
// BROKEN: No idempotency
async processEvent(event: EventEnvelope): Promise<void> {
  // MISSING: Duplicate check
  // if (this.processed_events.has(event.event_id)) {
  //   console.log(`Skipping duplicate event: ${event.event_id}`);
  //   return;
  // }

  switch (event.event_type) {
    case EventType.ORDER_CREATED:
      await this.handleOrderCreated(event);
      break;
    // ...
  }

  this.processed_events.add(event.event_id);
}
```

**Test Case:**
- Process event with event_id "evt_1"
- Process same event again
- Verify event is processed twice (bug - should be skipped on second attempt)

---

## Test Execution Plan

### Phase 1: WebSocket Chat (B01-B07)

**Test Setup:**
1. Start WebSocket server on ws://localhost:8080
2. Create 2 clients: Alice and Bob
3. Create test room: "general"

**Test Cases:**

**TC-WS-01: Error Handler Missing (B01)**
- Start server with B01 broken implementation
- Connect client and trigger malformed frame
- Expected: Server crashes with uncaught error
- Perfect: Server logs error, sends ERROR message, continues running

**TC-WS-02: Invalid Message Type (B02)**
- Connect Alice
- Send message: `{"type": "invalid_type", "user_id": "alice"}`
- Expected: Message accepted (bug)
- Perfect: Server returns ERROR message with "Unknown message type"

**TC-WS-03: No Authentication (B03)**
- Connect Alice (no JOIN sent)
- Send MESSAGE directly: `{"type": "message", "user_id": "alice", "room_id": "general", "content": "Hello"}`
- Expected: Message broadcasted (bug)
- Perfect: Server returns UNAUTHORIZED error

**TC-WS-04: Room Membership Bypass (B04)**
- Alice joins "room1"
- Alice sends message to "room2" (never joined)
- Expected: Message broadcasted to room2 (bug)
- Perfect: Server returns NOT_IN_ROOM error

**TC-WS-05: Type String Mismatch (B05)**
- Alice joins "general"
- Client sends message with type "msg" instead of "message"
- Expected: Message not handled, default case triggered
- Perfect: Server returns error for unknown type

**TC-WS-06: Empty Content (B06)**
- Alice joins "general"
- Alice sends message with content: ""
- Expected: Empty message broadcasted (bug)
- Perfect: Server returns EMPTY_MESSAGE error

**TC-WS-07: No Reconnect (B07)**
- Alice connects and joins "general"
- Kill server
- Restart server after 5 seconds
- Expected: Alice stays disconnected (bug)
- Perfect: Alice auto-reconnects within 5-10 seconds

**TC-WS-08: Typing Indicator**
- Alice and Bob both join "general"
- Alice sends typing=true
- Expected: Bob receives typing notification
- Alice sends typing=false
- Expected: Bob receives stop-typing notification

**TC-WS-09: Multi-Room**
- Alice joins "room1" and "room2"
- Bob joins "room2"
- Alice sends message to "room1"
- Expected: Bob doesn't receive it (not in room1)
- Alice sends message to "room2"
- Expected: Bob receives it

**TC-WS-10: Leave Room**
- Alice joins "general"
- Bob joins "general"
- Alice leaves "general"
- Expected: Bob receives LEAVE notification
- Alice sends message to "general"
- Expected: Server returns NOT_IN_ROOM error

---

### Phase 2: Message Queue (B08-B15)

**Test Setup:**
1. Start Go producer
2. Start TypeScript consumer with clean offset file
3. Create test correlation ID: "test_corr_123"

**Test Cases:**

**TC-MQ-01: Incomplete Event Handling (B08)**
- Producer publishes order.created → Consumer processes
- Producer publishes order.updated → Consumer logs "unknown event type" (bug)
- Producer publishes order.cancelled → Consumer logs "unknown event type" (bug)
- Expected: 2/3 events not handled
- Perfect: All 3 events processed with specific handlers

**TC-MQ-02: Version Not Checked (B09)**
- Producer publishes order.created with event_version: 999
- Expected: Consumer processes it (bug)
- Perfect: Consumer rejects with "Unsupported event version"

**TC-MQ-03: Missing Required Fields (B10)**
- Producer publishes order.created missing customer_id
- Expected: Consumer processes with customer_id=undefined, DB operation fails (bug)
- Perfect: Consumer validates and logs error, skips DB operation

**TC-MQ-04: Producer Validation (B11)**
- Call PublishOrderCreated with empty OrderID
- Expected: Event published anyway (bug)
- Perfect: Returns validation error

**TC-MQ-05: Queue Full (B12)**
- Fill queue with 100 events
- Publish 101st event
- Expected: Error returned but not logged (silent failure)
- Perfect: Producer logs error, implements retry or dead-letter queue

**TC-MQ-06: Offset Persistence (B14)**
- Consumer processes 10 events
- Kill consumer process
- Restart consumer
- Expected: All 10 events reprocessed (bug)
- Perfect: Consumer resumes from offset, skips 10 events

**TC-MQ-07: Idempotency (B15)**
- Process event "evt_001"
- Process same event "evt_001" again
- Expected: Event processed twice (bug)
- Perfect: Second attempt skipped with "duplicate event" log

**TC-MQ-08: Correlation ID Tracking**
- Publish 3 events with same correlation_id
- Expected: All 3 events logged with correlation_id
- Verify correlation ID appears in all handler logs

**TC-MQ-09: Event Ordering**
- Publish events with timestamps: [1000, 1001, 1002]
- Deliver out of order: [1001, 1000, 1002]
- Expected: Consumer logs warning for out-of-order event (timestamp 1000 < last 1001)

**TC-MQ-10: Crash Recovery**
- Consumer processes events 1-5
- Crash consumer before saveOffset() for event 5
- Restart consumer
- Expected: Event 5 reprocessed (idempotency prevents duplicate DB write)

---

## Expected Results Summary

### Perfect Implementation (All Tests Pass)

**WebSocket:**
- ✅ All 5 message types handled (join/message/leave/typing/error)
- ✅ Error handlers prevent crashes
- ✅ Authentication required before messaging
- ✅ Room membership enforced
- ✅ Empty content rejected
- ✅ Auto-reconnect on disconnect
- ✅ Typing indicators work
- ✅ Multi-room support
- ✅ Clean leave notifications

**Message Queue:**
- ✅ All 3 event types handled (order.created/updated/cancelled)
- ✅ Event version validation
- ✅ Required field validation
- ✅ Producer data validation
- ✅ Queue full error handling
- ✅ Offset persistence for crash recovery
- ✅ Idempotent event processing
- ✅ Correlation ID tracking
- ✅ Out-of-order event detection

### Broken Implementations (Specific Failures)

**B01:** Server crashes on connection error
**B02:** Invalid message types accepted
**B03:** Unauthenticated messaging allowed
**B04:** Room membership not enforced
**B05:** Type string mismatch causes failures
**B06:** Empty messages allowed
**B07:** No auto-reconnect
**B08:** 2/3 event types not handled
**B09:** Any event version accepted
**B10:** Missing fields not validated
**B11:** Invalid data published
**B12:** Silent queue full failures
**B13:** No WebSocket auth
**B14:** Events reprocessed after restart
**B15:** Duplicate events processed multiple times

---

## Linting Focus Areas

### WebSocket Server/Client

1. **Exhaustive switch statements** - All MessageType enum values must have cases
2. **Error handling** - All event listeners must have error handlers
3. **Authentication** - Connection and message-level auth checks
4. **Validation** - Message structure, required fields, empty content
5. **Type safety** - No `any`, strict enum usage
6. **Resource cleanup** - Remove from rooms/clients on disconnect

### Message Queue

1. **Exhaustive event handling** - All EventType enum values must have handlers
2. **Schema versioning** - event_version field must be validated
3. **Idempotency** - Duplicate event_id must be detected
4. **Offset tracking** - Consumer must persist offset to disk
5. **Field validation** - Required fields must be checked in all handlers
6. **Error propagation** - Producer must check Publish() errors

---

## File Structure

```
61-event-driven/
├── perfect/
│   ├── ws-chat-server.ts          (WebSocket server with all handlers)
│   ├── ws-chat-client.ts          (WebSocket client with reconnect)
│   ├── mq-producer.go             (Go message queue producer)
│   └── mq-consumer.ts             (TypeScript consumer with idempotency)
├── broken/
│   ├── B01-no-ws-error.ts         (Missing ws.on('error'))
│   ├── B02-no-type-validation.ts  (No MessageType validation)
│   ├── B03-no-auth.ts             (No authentication check)
│   ├── B04-no-room-check.ts       (Room membership not enforced)
│   ├── B05-type-mismatch.ts       (Type "msg" vs "message")
│   ├── B06-empty-content.ts       (Empty message allowed)
│   ├── B07-no-reconnect.ts        (No auto-reconnect)
│   ├── B08-incomplete-events.ts   (Only 1/3 event types handled)
│   ├── B09-no-version-check.ts    (Version field ignored)
│   ├── B10-no-field-validation.ts (Required fields not checked)
│   ├── B11-no-producer-validation.go (Producer doesn't validate)
│   ├── B12-queue-full.go          (Queue full error not handled)
│   ├── B13-no-ws-auth.ts          (No WebSocket connection auth)
│   ├── B14-no-offset.ts           (No offset tracking)
│   └── B15-no-idempotency.ts      (Duplicate events processed)
└── tests/
    ├── websocket-test-suite.ts    (TC-WS-01 through TC-WS-10)
    └── message-queue-test-suite.ts (TC-MQ-01 through TC-MQ-10)
```

---

## Estimated Complexity

**Lines of Code:**
- Perfect implementations: ~800 lines total
  - ws-chat-server.ts: ~300 lines
  - ws-chat-client.ts: ~150 lines
  - mq-producer.go: ~200 lines
  - mq-consumer.ts: ~250 lines
- Broken implementations: ~15 variants, ~50 lines modified per variant = ~750 lines
- Test suite: ~400 lines

**Total: ~1950 lines**

**Implementation Time:** ~8-10 hours
**Testing Time:** ~4-6 hours

---

## Success Criteria

1. Perfect implementation passes all 20 test cases (10 WebSocket + 10 MQ)
2. Each broken implementation fails exactly 1 specific test case
3. Linter detects all 15 bugs with specific error messages:
   - "Missing error handler for WebSocket connection"
   - "Message type not validated against MessageType enum"
   - "Missing authentication check before message handling"
   - "Room membership not validated"
   - "Message type string mismatch: 'msg' vs 'message'"
   - "Empty message content not validated"
   - "WebSocket client missing reconnect logic"
   - "Switch statement not exhaustive: missing EventType cases"
   - "Event version field not validated"
   - "Required field validation missing in event handler"
   - "Producer publishing event without data validation"
   - "Queue Publish() error not checked"
   - "WebSocket connection missing authentication"
   - "Consumer offset not persisted to disk"
   - "Duplicate event_id not detected (no idempotency)"
4. All code adheres to TypeScript strict mode and Go best practices
5. No `any` types in TypeScript code
6. All event/message enums exhaustively handled in switch statements

