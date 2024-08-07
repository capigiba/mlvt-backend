package models

import "time"

type Video struct {
	ID         string    `json:"id,omitempty"`      // ID of the video
	Duration   int64     `json:"duration"`          // duration of video
	FilePath   string    `json:"file_path"`         // type of video eg mp4, mp3, etc
	UploadedAt time.Time `json:"uploaded_at"`       // timestamp of video
	Size       int64     `json:"size"`              // size of video
	UserID     string    `json:"user_id,omitempty"` //
	Type       string    `json:"type,omitempty"`    //
}
