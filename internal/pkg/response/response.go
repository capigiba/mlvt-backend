package response

import "mlvt/internal/entity"

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// StatusResponse represents the response for GetVideoStatus
type StatusResponse struct {
	Status entity.VideoStatus `json:"status"`
}

// MessageResponse represents a message response
type MessageResponse struct {
	Message string `json:"message"`
}

// TokenResponse represents the response containing a token
type TokenResponse struct {
	Token  string `json:"token"`
	UserID uint64 `json:"user_id"`
}

// AvatarDownloadURLResponse represents the response containing avatar download URL
type AvatarDownloadURLResponse struct {
	AvatarDownloadURL string `json:"avatar_download_url"`
}

// AvatarUploadURLResponse represents the response containing avatar upload URL
type AvatarUploadURLResponse struct {
	AvatarUploadURL string `json:"avatar_upload_url"`
}

// UserResponse represents a single user response
type UserResponse struct {
	User entity.User `json:"user"`
}

// UsersResponse represents multiple users response
type UsersResponse struct {
	Users []entity.User `json:"users"`
}

// UploadURLResponse represents the response containing an upload URL
type UploadURLResponse struct {
	UploadURL string `json:"upload_url"`
}

// DownloadURLResponse represents the response containing a download URL
type DownloadURLResponse struct {
	DownloadURL string `json:"download_url"`
}

// TranscriptionResponse represents the response containing a transcription and its download URL
type TranscriptionResponse struct {
	Transcription entity.Transcription `json:"transcription"`
	DownloadURL   string               `json:"download_url"`
}

// TranscriptionsResponse represents the response containing a list of transcriptions
type TranscriptionsResponse struct {
	Transcriptions []entity.Transcription `json:"transcriptions"`
}

// AudioResponse represents the response containing an audio and its download URL
type AudioResponse struct {
	Audio       entity.Audio `json:"audio"`
	DownloadURL string       `json:"download_url"`
}

// AudiosResponse represents the response containing a list of audios
type AudiosResponse struct {
	Audios []entity.Audio `json:"audios"`
}
