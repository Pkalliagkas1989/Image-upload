package handlers

import (
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"forum/middleware"
	"forum/models"
	"forum/repository"
	"forum/utils"
)

const uploadBaseDir = "../ui/static/uploads/images"

type ImageHandler struct {
	ImageRepo *repository.ImageRepository
}

func NewImageHandler(repo *repository.ImageRepository) *ImageHandler {
	return &ImageHandler{ImageRepo: repo}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(21 << 20); err != nil {
		utils.ErrorResponse(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	postID := r.FormValue("post_id")
	if postID == "" {
		utils.ErrorResponse(w, "Post ID required", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		utils.ErrorResponse(w, "Image file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > 20<<20 {
		utils.ErrorResponse(w, "Image exceeds 20 MB limit", http.StatusBadRequest)
		return
	}

	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	file.Seek(0, 0)

	var img image.Image
	var ext string
	switch contentType {
	case "image/jpeg":
		img, err = jpeg.Decode(file)
		ext = ".jpg"
	case "image/png":
		img, err = png.Decode(file)
		ext = ".png"
	case "image/gif":
		img, err = gif.Decode(file)
		ext = ".gif"
	default:
		utils.ErrorResponse(w, "Unsupported image type", http.StatusBadRequest)
		return
	}
	if err != nil {
		utils.ErrorResponse(w, "Failed to decode image", http.StatusBadRequest)
		return
	}

	dateStr := time.Now().Format("2006-01-02")
	baseDir := filepath.Join(uploadBaseDir, user.ID, dateStr)
	thumbDir := filepath.Join(baseDir, "thumbnails")
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		utils.ErrorResponse(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	uuid := utils.GenerateUUID()
	fileName := uuid + ext
	filePath := filepath.Join(baseDir, fileName)
	out, err := os.Create(filePath)
	if err != nil {
		utils.ErrorResponse(w, "Failed to save image", http.StatusInternalServerError)
		return
	}
	if err := encodeImage(out, img, contentType); err != nil {
		out.Close()
		utils.ErrorResponse(w, "Failed to save image", http.StatusInternalServerError)
		return
	}
	out.Close()

	thumbImg := createThumbnail(img)
	thumbPath := filepath.Join(thumbDir, fileName)
	outT, err := os.Create(thumbPath)
	if err != nil {
		utils.ErrorResponse(w, "Failed to save thumbnail", http.StatusInternalServerError)
		return
	}
	if err := encodeImage(outT, thumbImg, contentType); err != nil {
		outT.Close()
		utils.ErrorResponse(w, "Failed to save thumbnail", http.StatusInternalServerError)
		return
	}
	outT.Close()

	relPath := filepath.ToSlash(strings.TrimPrefix(filePath, "../ui/static/"))
	relThumb := filepath.ToSlash(strings.TrimPrefix(thumbPath, "../ui/static/"))
	imgModel := models.Image{
		PostID:        postID,
		UserID:        user.ID,
		FilePath:      relPath,
		ThumbnailPath: relThumb,
	}

	created, err := h.ImageRepo.Create(imgModel)
	if err != nil {
		utils.ErrorResponse(w, "Failed to save image", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, created, http.StatusCreated)
}

func encodeImage(w *os.File, img image.Image, contentType string) error {
	switch contentType {
	case "image/jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
	case "image/png":
		return png.Encode(w, img)
	case "image/gif":
		return gif.Encode(w, img, nil)
	}
	return nil
}

func createThumbnail(src image.Image) image.Image {
	const size = 150
	sw := src.Bounds().Dx()
	sh := src.Bounds().Dy()
	scale := float64(size) / float64(sw)
	if sh > sw {
		scale = float64(size) / float64(sh)
	}
	nw := int(float64(sw) * scale)
	nh := int(float64(sh) * scale)
	resized := resizeImage(src, nw, nh)
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	drawBackground(dst, color.Black)
	offX := (size - nw) / 2
	offY := (size - nh) / 2
	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			dst.Set(offX+x, offY+y, resized.At(x, y))
		}
	}
	return dst
}

func drawBackground(img *image.RGBA, c color.Color) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

func resizeImage(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sb := src.Bounds()
	sw := sb.Dx()
	sh := sb.Dy()
	for y := 0; y < h; y++ {
		sy := sb.Min.Y + int(float64(y)*float64(sh)/float64(h))
		for x := 0; x < w; x++ {
			sx := sb.Min.X + int(float64(x)*float64(sw)/float64(w))
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
