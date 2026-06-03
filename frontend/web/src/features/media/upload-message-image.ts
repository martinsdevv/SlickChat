import { apiRequest } from "../../shared/api/http-client";
import { ApiError } from "../../shared/api/types";
import type { MediaUploadCompleteResponse, MediaUploadRequestResponse } from "../../shared/api/types";
import { getReliableFileSize, resolveImageContentType } from "./image-file";
import { putUploadedFile } from "./put-uploaded-file";
import { validateMessageImageFile } from "./upload-room-media";

export async function uploadMessageImage(
  token: string,
  roomId: string,
  messageId: string,
  file: File,
): Promise<MediaUploadCompleteResponse> {
  const validationError = await validateMessageImageFile(file);
  if (validationError) {
    throw new ApiError({ message: validationError, status: 400 });
  }

  const contentType = resolveImageContentType(file)!;
  const sizeBytes = await getReliableFileSize(file);

  const request = await apiRequest<MediaUploadRequestResponse>("/media/upload-request", {
    method: "POST",
    token,
    body: {
      purpose: "message_image",
      room_id: roomId,
      message_id: messageId,
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
      purpose: "message_image",
      room_id: roomId,
      upload_id: request.upload_id,
      object_key: request.object_key,
    },
  });
}
