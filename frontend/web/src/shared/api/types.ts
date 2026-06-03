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
  avatar_object_key?: string;
  expires_at: string;
};

export type MeResponse = {
  user_id: string;
  handle: string;
  avatar_object_key?: string;
  created_at: string;
};

export type RoomUnreadItem = {
  room_id: string;
  count: number;
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
  avatar_object_key?: string;
  banner_object_key?: string;
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
  avatar_object_key?: string;
};

export type RoomMediaPurpose = "room_avatar" | "room_banner" | "message_image" | "user_avatar";

export type MediaUploadRequestResponse = {
  upload_id: string;
  object_key: string;
  upload_url: string;
  upload_via_api?: boolean;
  expires_in_seconds: number;
};

export type MediaUploadCompleteResponse = {
  object_key: string;
  read_url?: string;
};

export type WSTicketResponse = {
  ticket: string;
};

export type MessageHistoryItem = {
  id: string;
  sender_id?: string;
  content: string;
  caption?: string;
  type: string;
  attachment_object_key?: string;
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
    message_id?: string;
    message_type?: string;
    object_key?: string;
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
