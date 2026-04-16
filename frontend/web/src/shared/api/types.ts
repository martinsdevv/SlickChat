export type ApiErrorShape = {
  message: string;
  status: number;
  details?: unknown;
};

export class ApiError extends Error {
  status: number;
  details?: unknown;

  constructor({ message, status, details }: ApiErrorShape) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.details = details;
  }
}

export type RegisterRequest = {
  username: string;
  password: string;
};

export type RegisterResponse = {
  handle: string;
  recovery_key: string;
  created_at: string;
};

export type LoginRequest = {
  handle: string;
  password: string;
};

export type LoginResponse = {
  token: string;
  user_id: string;
  handle: string;
  expires_at: string;
};

export type MeResponse = {
  user_id: string;
  handle: string;
  created_at: string;
};

export type Room = {
  room_id: string;
  name: string;
  description: string;
  type: "PUBLIC" | "PRIVATE" | "DIRECT" | "TEMPORARY";
  owner_id?: string;
  ttl: number;
  paranoid_mode: boolean;
  zero_logging: boolean;
  created_at: string;
  expires_at?: string;
};

export type CreateRoomRequest = {
  name: string;
  description?: string;
  type: Room["type"];
  ttl: number;
  zero_logging: boolean;
};

export type RoomMember = {
  user_id: string;
  handle: string;
  role: "ADMIN" | "MODERATOR" | "MEMBER";
};

export type WSTicketResponse = {
  ticket: string;
};

export type MessageHistoryItem = {
  id: string;
  sender_id?: string;
  content: string;
  type: string;
  created_at: string;
  expires_at?: string | null;
};

export type WsOutEvent =
  | "message.received"
  | "message.delivered"
  | "message.read"
  | "message.deleted"
  | "message.expired"
  | "message_ack"
  | "error"
  | "session_expired";

export type WsInboundPayload = {
  send_message: {
    room_id: string;
    content: string;
  };
  delete_message: {
    room_id: string;
    message_id: string;
  };
  message_delivered: {
    room_id: string;
    message_id: string;
  };
  message_read: {
    room_id: string;
    message_id: string;
  };
};
