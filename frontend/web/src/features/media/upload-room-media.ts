import { apiRequest } from "../../shared/api/http-client";
import { ApiError } from "../../shared/api/types";
import type {
  MediaUploadCompleteResponse,
  MediaUploadRequestResponse,
  RoomMediaPurpose,
} from "../../shared/api/types";
import {
  getReliableFileSize,
  MEDIA_LIMITS,
  resolveImageContentType,
  validateImageFile,
} from "./image-file";
import { putUploadedFile } from "./put-uploaded-file";

export async function validateRoomImageFile(
  file: File,
  purpose: RoomMediaPurpose,
): Promise<string | null> {
  if (purpose === "room_avatar") {
    return validateImageFile(file, MEDIA_LIMITS.roomAvatar, "5 MB");
  }
  return validateImageFile(file, MEDIA_LIMITS.roomBanner, "8 MB");
}

export async function validateMessageImageFile(file: File): Promise<string | null> {
  return validateImageFile(file, MEDIA_LIMITS.messageImage, "15 MB");
}

export async function uploadRoomMedia(
  token: string,
  roomId: string,
  purpose: RoomMediaPurpose,
  file: File,
): Promise<MediaUploadCompleteResponse> {
  const validationError = await validateRoomImageFile(file, purpose);
  if (validationError) {
    throw new ApiError({ message: validationError, status: 400 });
  }

  const contentType = resolveImageContentType(file)!;
  const sizeBytes = await getReliableFileSize(file);

  const request = await apiRequest<MediaUploadRequestResponse>("/media/upload-request", {
    method: "POST",
    token,
    body: {
      purpose,
      room_id: roomId,
      content_type: contentType,
      size_bytes: sizeBytes,
    },
  });

  await putUploadedFile(
    request.upload_url,
    token,
    file,
    contentType,
    request.object_key,
    request.upload_via_api,
  );

  return apiRequest<MediaUploadCompleteResponse>("/media/upload-complete", {
    method: "POST",
    token,
    body: {
      purpose,
      room_id: roomId,
      upload_id: request.upload_id,
      object_key: request.object_key,
    },
  });
}
