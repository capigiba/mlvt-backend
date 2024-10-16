package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mlvt/internal/entity"
	"mlvt/internal/infra/env"
	"mlvt/internal/pkg/response"
	"mlvt/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupRouter initializes the Gin router with the VideoController routes
func setupRouter(controller *VideoController) *gin.Engine {
	// Set Gin to Test Mode to reduce unnecessary logs
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Register routes
	router.GET("/videos/:video_id/status", controller.GetVideoStatus)
	router.PUT("/videos/:video_id/status", controller.UpdateVideoStatus)
	router.POST("/videos", controller.AddVideo)
	router.POST("/videos/generate-upload-url/video", controller.GenerateUploadURLForVideo)
	router.POST("/videos/generate-upload-url/image", controller.GenerateUploadURLForImage)
	router.GET("/videos/:video_id/download-url/video", controller.GenerateDownloadURLForVideo)
	router.GET("/videos/:video_id/download-url/image", controller.GenerateDownloadURLForImage)
	router.GET("/videos/:video_id", controller.GetVideoByID)
	router.DELETE("/videos/:video_id", controller.DeleteVideo)
	router.GET("/videos/user/:user_id", controller.ListVideosByUserID)

	return router
}

func TestGetVideoStatus(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	t.Run("Success", func(t *testing.T) {
		videoID := uint64(1)
		status := entity.StatusSuccess

		mockService.On("GetVideoStatus", videoID).Return(status, nil)

		req, _ := http.NewRequest("GET", "/videos/1/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.StatusResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, status, resp.Status)

		mockService.AssertCalled(t, "GetVideoStatus", videoID)
	})

	t.Run("Invalid Video ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/videos/abc/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid video ID", resp.Error)
	})

	t.Run("Video Not Found", func(t *testing.T) {
		videoID := uint64(2)
		errMsg := "video with ID 2 does not exist"
		mockService.On("GetVideoStatus", videoID).Return(entity.VideoStatus(""), errors.New(errMsg))

		req, _ := http.NewRequest("GET", "/videos/2/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "video not found", resp.Error)

		mockService.AssertCalled(t, "GetVideoStatus", videoID)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		videoID := uint64(3)
		errMsg := "database connection failed"
		mockService.On("GetVideoStatus", videoID).Return(entity.VideoStatus(""), errors.New(errMsg))

		req, _ := http.NewRequest("GET", "/videos/3/status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal server error", resp.Error)

		mockService.AssertCalled(t, "GetVideoStatus", videoID)
	})
}

func TestUpdateVideoStatus(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	t.Run("Success", func(t *testing.T) {
		videoID := uint64(1)
		newStatus := entity.StatusProcessing

		mockService.On("UpdateVideoStatus", videoID, newStatus).Return(nil)

		reqBody := UpdateVideoStatusRequest{
			Status: newStatus,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/videos/1/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.MessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "status updated successfully", resp.Message)

		mockService.AssertCalled(t, "UpdateVideoStatus", videoID, newStatus)
	})

	t.Run("Invalid Video ID", func(t *testing.T) {
		reqBody := UpdateVideoStatusRequest{
			Status: entity.StatusProcessing,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/videos/abc/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid video ID", resp.Error)
	})

	t.Run("Invalid Input", func(t *testing.T) {
		// Missing 'status' field
		body := []byte(`{}`)
		req, _ := http.NewRequest("PUT", "/videos/1/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid input", resp.Error)
	})

	t.Run("Video Not Found", func(t *testing.T) {
		videoID := uint64(2)
		newStatus := entity.StatusFailed
		errMsg := "no video found with id 2"
		mockService.On("UpdateVideoStatus", videoID, newStatus).Return(errors.New(errMsg))

		reqBody := UpdateVideoStatusRequest{
			Status: newStatus,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/videos/2/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "video not found", resp.Error)

		mockService.AssertCalled(t, "UpdateVideoStatus", videoID, newStatus)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		videoID := uint64(3)
		newStatus := entity.StatusSuccess
		errMsg := "database update failed"
		mockService.On("UpdateVideoStatus", videoID, newStatus).Return(errors.New(errMsg))

		reqBody := UpdateVideoStatusRequest{
			Status: newStatus,
		}
		body, _ := json.Marshal(reqBody)
		req, _ := http.NewRequest("PUT", "/videos/3/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal server error", resp.Error)

		mockService.AssertCalled(t, "UpdateVideoStatus", videoID, newStatus)
	})
}

func TestAddVideo(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	t.Run("Success", func(t *testing.T) {
		video := entity.Video{
			ID:          1,
			Title:       "Test Video",
			Duration:    120,
			Description: "A test video",
			FileName:    "test.mp4",
			Folder:      "videos",
			Image:       "test.jpg",
			Status:      entity.StatusRaw,
			UserID:      1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Use mock.Anything to ignore the actual Video instance
		mockService.On("CreateVideo", mock.AnythingOfType("*entity.Video")).Return(nil)

		body, _ := json.Marshal(video)
		req, _ := http.NewRequest("POST", "/videos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.MessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "Video added successfully", resp.Message)

		mockService.AssertCalled(t, "CreateVideo", mock.AnythingOfType("*entity.Video"))
	})
}

func TestGenerateUploadURLForVideo(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	// Mock environment variable
	originalFolder := env.EnvConfig.VideosFolder
	env.EnvConfig.VideosFolder = "test_videos"
	defer func() {
		env.EnvConfig.VideosFolder = originalFolder
	}()

	t.Run("Success", func(t *testing.T) {
		fileName := "video.mp4"
		fileType := "video/mp4"
		uploadURL := "https://s3.amazonaws.com/test_videos/video.mp4?signature=abc"

		mockService.On("GeneratePresignedUploadURLForVideo", "test_videos", fileName, fileType).Return(uploadURL, nil)

		req, _ := http.NewRequest("POST", "/videos/generate-upload-url/video?file_name=video.mp4&file_type=video/mp4", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uploadURL, resp["upload_url"])

		mockService.AssertCalled(t, "GeneratePresignedUploadURLForVideo", "test_videos", fileName, fileType)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		fileName := "video2.mp4"
		fileType := "video/mp4"
		errMsg := "S3 service unavailable"

		mockService.On("GeneratePresignedUploadURLForVideo", "test_videos", fileName, fileType).Return("", errors.New(errMsg))

		req, _ := http.NewRequest("POST", "/videos/generate-upload-url/video?file_name=video2.mp4&file_type=video/mp4", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, errMsg, resp.Error)

		mockService.AssertCalled(t, "GeneratePresignedUploadURLForVideo", "test_videos", fileName, fileType)
	})
}

func TestGenerateUploadURLForImage(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	// Mock environment variable
	originalFolder := env.EnvConfig.VideoFramesFolder
	env.EnvConfig.VideoFramesFolder = "test_frames"
	defer func() {
		env.EnvConfig.VideoFramesFolder = originalFolder
	}()

	t.Run("Success", func(t *testing.T) {
		fileName := "image.jpg"
		fileType := "image/jpeg"
		uploadURL := "https://s3.amazonaws.com/test_frames/image.jpg?signature=xyz"

		mockService.On("GeneratePresignedUploadURLForImage", "test_frames", fileName, fileType).Return(uploadURL, nil)

		req, _ := http.NewRequest("POST", "/videos/generate-upload-url/image?file_name=image.jpg&file_type=image/jpeg", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uploadURL, resp["upload_url"])

		mockService.AssertCalled(t, "GeneratePresignedUploadURLForImage", "test_frames", fileName, fileType)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		fileName := "image2.jpg"
		fileType := "image/jpeg"
		errMsg := "S3 service timeout"

		mockService.On("GeneratePresignedUploadURLForImage", "test_frames", fileName, fileType).Return("", errors.New(errMsg))

		req, _ := http.NewRequest("POST", "/videos/generate-upload-url/image?file_name=image2.jpg&file_type=image/jpeg", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, errMsg, resp.Error)

		mockService.AssertCalled(t, "GeneratePresignedUploadURLForImage", "test_frames", fileName, fileType)
	})
}

func TestGenerateDownloadURLForVideo(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	t.Run("Success", func(t *testing.T) {
		videoID := uint64(1)
		downloadURL := "https://s3.amazonaws.com/videos/video.mp4?signature=download"

		mockService.On("GeneratePresignedDownloadURLForVideo", videoID).Return(downloadURL, nil)

		req, _ := http.NewRequest("GET", "/videos/1/download-url/video", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, downloadURL, resp["video_download_url"])

		mockService.AssertCalled(t, "GeneratePresignedDownloadURLForVideo", videoID)
	})

	t.Run("Invalid Video ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/videos/abc/download-url/video", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid video ID", resp.Error)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		videoID := uint64(2)
		errMsg := "Failed to generate download URL"

		mockService.On("GeneratePresignedDownloadURLForVideo", videoID).Return("", errors.New(errMsg))

		req, _ := http.NewRequest("GET", "/videos/2/download-url/video", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, errMsg, resp.Error)

		mockService.AssertCalled(t, "GeneratePresignedDownloadURLForVideo", videoID)
	})
}

func TestGenerateDownloadURLForImage(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	t.Run("Success", func(t *testing.T) {
		videoID := uint64(1)
		downloadURL := "https://s3.amazonaws.com/images/image.jpg?signature=download"

		mockService.On("GeneratePresignedDownloadURLForImage", videoID).Return(downloadURL, nil)

		req, _ := http.NewRequest("GET", "/videos/1/download-url/image", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, downloadURL, resp["image_download_url"])

		mockService.AssertCalled(t, "GeneratePresignedDownloadURLForImage", videoID)
	})

	t.Run("Invalid Video ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/videos/abc/download-url/image", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid video ID", resp.Error)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		videoID := uint64(2)
		errMsg := "Failed to generate image download URL"

		mockService.On("GeneratePresignedDownloadURLForImage", videoID).Return("", errors.New(errMsg))

		req, _ := http.NewRequest("GET", "/videos/2/download-url/image", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, errMsg, resp.Error)

		mockService.AssertCalled(t, "GeneratePresignedDownloadURLForImage", videoID)
	})
}

func TestGetVideoByID(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	t.Run("Success", func(t *testing.T) {
		videoID := uint64(1)
		fixedTime := time.Date(2024, time.October, 14, 21, 40, 27, 24360000, time.Local)

		expectedVideo := &entity.Video{
			ID:          videoID,
			Title:       "Test Video",
			Duration:    120,
			Description: "A test video",
			FileName:    "test.mp4",
			Folder:      "videos",
			Image:       "test.jpg",
			Status:      entity.StatusSuccess,
			UserID:      1,
			CreatedAt:   fixedTime,
			UpdatedAt:   fixedTime,
		}
		videoURL := "https://s3.amazonaws.com/videos/test.mp4?signature=download"
		imageURL := "https://s3.amazonaws.com/images/test.jpg?signature=download"

		// Set up mock expectation
		mockService.On("GetVideoByID", videoID).Return(expectedVideo, videoURL, imageURL, nil)

		req, _ := http.NewRequest("GET", "/videos/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp struct {
			Video    entity.Video `json:"video"`
			VideoURL string       `json:"video_url"`
			ImageURL string       `json:"image_url"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)

		// Compare Video fields individually
		assert.Equal(t, expectedVideo.ID, resp.Video.ID, "Video ID should match")
		assert.Equal(t, expectedVideo.Title, resp.Video.Title, "Video Title should match")
		assert.Equal(t, expectedVideo.Duration, resp.Video.Duration, "Video Duration should match")
		assert.Equal(t, expectedVideo.Description, resp.Video.Description, "Video Description should match")
		assert.Equal(t, expectedVideo.FileName, resp.Video.FileName, "Video FileName should match")
		assert.Equal(t, expectedVideo.Folder, resp.Video.Folder, "Video Folder should match")
		assert.Equal(t, expectedVideo.Image, resp.Video.Image, "Video Image should match")
		assert.Equal(t, expectedVideo.Status, resp.Video.Status, "Video Status should match")
		assert.Equal(t, expectedVideo.UserID, resp.Video.UserID, "Video UserID should match")
		assert.True(t, expectedVideo.CreatedAt.Equal(resp.Video.CreatedAt), "Video CreatedAt should match")
		assert.True(t, expectedVideo.UpdatedAt.Equal(resp.Video.UpdatedAt), "Video UpdatedAt should match")

		// Compare URLs
		assert.Equal(t, videoURL, resp.VideoURL, "VideoURL should match")
		assert.Equal(t, imageURL, resp.ImageURL, "ImageURL should match")

		mockService.AssertCalled(t, "GetVideoByID", videoID)
	})

	t.Run("Invalid Video ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/videos/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid video ID", resp.Error)
	})

	t.Run("Video Not Found", func(t *testing.T) {
		videoID := uint64(2)
		mockService.On("GetVideoByID", videoID).Return((*entity.Video)(nil), "", "", errors.New("video not found"))

		req, _ := http.NewRequest("GET", "/videos/2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "video not found", resp.Error)

		mockService.AssertCalled(t, "GetVideoByID", videoID)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		videoID := uint64(3)
		mockService.On("GetVideoByID", videoID).Return((*entity.Video)(nil), "", "", errors.New("database error"))

		req, _ := http.NewRequest("GET", "/videos/3", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal server error", resp.Error)

		mockService.AssertCalled(t, "GetVideoByID", videoID)
	})
}

func TestDeleteVideo(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	t.Run("Success", func(t *testing.T) {
		videoID := uint64(1)

		mockService.On("DeleteVideo", videoID).Return(nil)

		req, _ := http.NewRequest("DELETE", "/videos/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.MessageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Video deleted successfully", resp.Message)

		mockService.AssertCalled(t, "DeleteVideo", videoID)
	})

	t.Run("Invalid Video ID", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/videos/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid video ID", resp.Error)
	})

	t.Run("Video Not Found", func(t *testing.T) {
		videoID := uint64(2)
		mockService.On("DeleteVideo", videoID).Return(errors.New("video not found"))

		req, _ := http.NewRequest("DELETE", "/videos/2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "video not found", resp.Error)

		mockService.AssertCalled(t, "DeleteVideo", videoID)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		videoID := uint64(3)
		mockService.On("DeleteVideo", videoID).Return(errors.New("database deletion failed"))

		req, _ := http.NewRequest("DELETE", "/videos/3", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal server error", resp.Error)

		mockService.AssertCalled(t, "DeleteVideo", videoID)
	})
}

func TestListVideosByUserID(t *testing.T) {
	mockService := new(service.MockVideoService)
	controller := NewVideoController(mockService)
	router := setupRouter(controller)

	t.Run("Success", func(t *testing.T) {
		userID := uint64(1)
		fixedTime := time.Date(2024, time.October, 14, 21, 35, 25, 616671000, time.Local)

		videos := []entity.Video{
			{
				ID:          1,
				Title:       "Video 1",
				Duration:    100,
				Description: "First video",
				FileName:    "video1.mp4",
				Folder:      "videos",
				Image:       "video1.jpg",
				Status:      entity.StatusRaw,
				UserID:      userID,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
			{
				ID:          2,
				Title:       "Video 2",
				Duration:    200,
				Description: "Second video",
				FileName:    "video2.mp4",
				Folder:      "videos",
				Image:       "video2.jpg",
				Status:      entity.StatusProcessing,
				UserID:      userID,
				CreatedAt:   fixedTime,
				UpdatedAt:   fixedTime,
			},
		}
		frames := []entity.Frame{
			{
				VideoID: 1,
				Link:    "https://s3.amazonaws.com/images/video1_frame1.jpg",
			},
			{
				VideoID: 2,
				Link:    "https://s3.amazonaws.com/images/video2_frame1.jpg",
			},
		}

		// Set up mock expectation
		mockService.On("ListVideosByUserID", userID).Return(videos, frames, nil)

		req, _ := http.NewRequest("GET", "/videos/user/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp struct {
			Videos []entity.Video `json:"videos"`
			Frames []entity.Frame `json:"frames"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, videos, resp.Videos)
		assert.Equal(t, frames, resp.Frames)

		mockService.AssertCalled(t, "ListVideosByUserID", userID)
	})

	t.Run("Invalid User ID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/videos/user/abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "invalid user ID", resp.Error)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		userID := uint64(2)

		// Set up mock to return empty slices and an error
		mockService.On("ListVideosByUserID", userID).Return([]entity.Video{}, []entity.Frame{}, errors.New("database query failed"))

		req, _ := http.NewRequest("GET", "/videos/user/2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp response.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "internal server error", resp.Error)

		mockService.AssertCalled(t, "ListVideosByUserID", userID)
	})
}
