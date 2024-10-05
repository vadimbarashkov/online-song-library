package http

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

// songSchema represents the structure of a song entity for API responses.
// It includes metadata about the song along with detailed information.
type songSchema struct {
	ID         uuid.UUID        `json:"id"`
	GroupName  string           `json:"groupName"`
	Name       string           `json:"name"`
	SongDetail songDetailSchema `json:"songDetail"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// songDetailSchema represents detailed information about a song.
// It includes the release date, text, and a link to the song.
type songDetailSchema struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

// songWithVersesSchema is a structure used for responses containing a song and its verses.
type songWithVersesSchema struct {
	ID        uuid.UUID `json:"id"`
	GroupName string    `json:"groupName"`
	Name      string    `json:"name"`
	Verses    []string  `json:"verses"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// paginationSchema represents pagination metadata for API responses.
type paginationSchema struct {
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
	Items  uint64 `json:"items"`
	Total  uint64 `json:"total"`
}

// addSongRequest defines the expected structure for requests to add a new song.
type addSongRequest struct {
	Group string `json:"group" validate:"required"`
	Song  string `json:"song" validate:"required"`
}

// updateSongRequest defines the expected structure for requests to update an existing song.
type updateSongRequest struct {
	GroupName   string `json:"groupName"`
	Name        string `json:"name"`
	ReleaseDate string `json:"releaseDate" validate:"omitempty,releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link" validate:"url,omitempty"`
}

// songsResponse represents the structure of the response for fetching multiple songs.
type songsResponse struct {
	Songs      []songSchema     `json:"songs"`
	Pagination paginationSchema `json:"pagination"`
}

// songWithVersesResponse represents the structure of the response for fetching a song with its verses.
type songWithVersesResponse struct {
	Song       songWithVersesSchema `json:"song"`
	Pagination paginationSchema     `json:"pagination"`
}

// parsePagination extracts pagination parameters from the HTTP request query.
func parsePagination(r *http.Request) entity.Pagination {
	getUintQueryParam := func(key string, defaultValue uint64) uint64 {
		param := r.URL.Query().Get(key)
		if param != "" {
			value, err := strconv.ParseUint(param, 10, 64)
			if err == nil {
				return value
			}
		}

		return defaultValue
	}

	pagination := entity.Pagination{
		Offset: getUintQueryParam("offset", entity.DefaultOffset),
		Limit:  getUintQueryParam("limit", entity.DefaultLimit),
	}

	return pagination
}

// parseSongFilters extracts song filter criteria from the HTTP request query.
func parseSongFilters(r *http.Request) []entity.SongFilter {
	var filters []entity.SongFilter

	addStringFilter := func(paramValue string, field entity.SongFilterField) {
		if paramValue != "" {
			filters = append(filters, entity.SongFilter{
				Field: field,
				Value: paramValue,
			})
		}
	}

	addDateFilter := func(paramValue string, field entity.SongFilterField) {
		if paramValue != "" {
			value, err := time.Parse("02.01.2006", paramValue)
			if err == nil {
				filters = append(filters, entity.SongFilter{
					Field: field,
					Value: value,
				})
			}
		}
	}

	query := r.URL.Query()

	addStringFilter(query.Get("groupName"), entity.SongGroupNameFilterField)
	addStringFilter(query.Get("name"), entity.SongNameFilterField)
	addDateFilter(query.Get("releaseYear"), entity.SongReleaseYearFilterField)
	addDateFilter(query.Get("releaseDate"), entity.SongReleaseDateFilterField)
	addDateFilter(query.Get("releaseDateAfter"), entity.SongReleaseDateAfterFilterField)
	addDateFilter(query.Get("releaseDateBefore"), entity.SongReleaseDateBeforeFilterField)
	addStringFilter(query.Get("text"), entity.SongTextFilterField)

	return filters
}

const statusError = "error"

// errorResponse represents the structure of error responses from the API.
type errorResponse struct {
	Status  string   `json:"status"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

// Predefined error responses for common scenarios
var (
	emptyRequestBodyResp = errorResponse{
		Status:  statusError,
		Message: "empty request body",
	}

	invalidRequestBodyResp = errorResponse{
		Status:  statusError,
		Message: "invalid request body",
	}

	invalidSongIDParamResp = errorResponse{
		Status:  statusError,
		Message: "invalid song id param",
	}

	songNotFoundErrResp = errorResponse{
		Status:  statusError,
		Message: "song not found",
	}

	serverErrResp = errorResponse{
		Status:  statusError,
		Message: "server error occurred",
	}
)

// messageForValidateTag returns a user-friendly message for validation errors based on the tag.
func messageForValidateTag(tag string) string {
	switch tag {
	case "required":
		return "required field"
	case "releaseDate":
		return "invalid format, must be like '02.01.2006'"
	case "url":
		return "invalid url"
	default:
		return "invalid value"
	}
}

// getValidationErrorDetails extracts detailed error messages from validation errors.
func getValidationErrorDetails(err error) []string {
	var details []string

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			field := e.Field()
			msg := messageForValidateTag(e.Tag())

			details = append(details, fmt.Sprintf("%s: %s", field, msg))
		}
	}

	return details
}

// validationError creates an errorResponse for validation errors.
func validationError(err error) errorResponse {
	return errorResponse{
		Status:  statusError,
		Message: "validation error",
		Details: getValidationErrorDetails(err),
	}
}
