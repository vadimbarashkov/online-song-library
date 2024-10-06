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
//
//	@Description	Represents the structure of a song entity for API responses.
//	@Tags			songs
type songSchema struct {
	ID         uuid.UUID        `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	GroupName  string           `json:"groupName" example:"The Beatles"`
	Name       string           `json:"name" example:"Hey Jude"`
	SongDetail songDetailSchema `json:"songDetail"`
	CreatedAt  time.Time        `json:"created_at" example:"2024-10-05T14:48:00Z"`
	UpdatedAt  time.Time        `json:"updated_at" example:"2024-10-06T09:12:00Z"`
}

// songDetailSchema represents detailed information about a song.
// It includes the release date, text, and a link to the song.
//
//	@Description	Represents detailed information about a song.
//	@Tags			songs
type songDetailSchema struct {
	ReleaseDate string `json:"releaseDate" example:"02.01.1968"`
	Text        string `json:"text" example:"Hey Jude, don't make it bad..."`
	Link        string `json:"link" example:"https://example.com/heyjude"`
}

// songWithVersesSchema is a structure used for responses containing a song and its verses.
//
//	@Description	Represents a song and its verses for API responses.
//	@Tags			songs
type songWithVersesSchema struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174001"`
	GroupName string    `json:"groupName" example:"Queen"`
	Name      string    `json:"name" example:"Bohemian Rhapsody"`
	Verses    []string  `json:"verses" example:"Is this the real life?,Is this just fantasy?"`
	CreatedAt time.Time `json:"created_at" example:"2024-10-05T14:48:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-10-06T09:12:00Z"`
}

// paginationSchema represents pagination metadata for API responses.
//
//	@Description	Represents pagination metadata for API responses.
//	@Tags			pagination
type paginationSchema struct {
	Offset uint64 `json:"offset" example:"0"`
	Limit  uint64 `json:"limit" example:"10"`
	Items  uint64 `json:"items" example:"2"`
	Total  uint64 `json:"total" example:"100"`
}

// addSongRequest defines the expected structure for requests to add a new song.
//
//	@Description	Defines the expected structure for requests to add a new song.
//	@Tags			songs
type addSongRequest struct {
	Group string `json:"group" validate:"required" example:"The Rolling Stones"`
	Song  string `json:"song" validate:"required" example:"Paint It Black"`
}

// updateSongRequest defines the expected structure for requests to update an existing song.
//
//	@Description	Defines the expected structure for requests to update an existing song.
//	@Tags			songs
type updateSongRequest struct {
	GroupName   string `json:"groupName" example:"Led Zeppelin"`
	Name        string `json:"name" example:"Stairway to Heaven"`
	ReleaseDate string `json:"releaseDate" validate:"omitempty,releaseDate" example:"08.11.1971"`
	Text        string `json:"text" example:"There's a lady who's sure..."`
	Link        string `json:"link" validate:"url,omitempty" example:"https://example.com/stairway"`
}

// songsResponse represents the structure of the response for fetching multiple songs.
//
//	@Description	Represents the structure of the response for fetching multiple songs.
//	@Tags			songs
type songsResponse struct {
	Songs      []songSchema     `json:"songs"`
	Pagination paginationSchema `json:"pagination"`
}

// songWithVersesResponse represents the structure of the response for fetching a song with its verses.
//
//	@Description	Represents the structure of the response for fetching a song with its verses.
//	@Tags			songs
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

	addStringFilter := func(param string, field entity.SongFilterField) {
		if param != "" {
			filters = append(filters, entity.SongFilter{
				Field: field,
				Value: param,
			})
		}
	}

	addIntFilter := func(param string, field entity.SongFilterField) {
		if param != "" {
			value, err := strconv.Atoi(param)
			if err == nil {
				filters = append(filters, entity.SongFilter{
					Field: field,
					Value: value,
				})
			}
		}
	}

	addDateFilter := func(param string, field entity.SongFilterField) {
		if param != "" {
			value, err := time.Parse("02.01.2006", param)
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
	addIntFilter(query.Get("releaseYear"), entity.SongReleaseYearFilterField)
	addDateFilter(query.Get("releaseDate"), entity.SongReleaseDateFilterField)
	addDateFilter(query.Get("releaseDateAfter"), entity.SongReleaseDateAfterFilterField)
	addDateFilter(query.Get("releaseDateBefore"), entity.SongReleaseDateBeforeFilterField)
	addStringFilter(query.Get("text"), entity.SongTextFilterField)

	return filters
}

const statusError = "error"

// errorResponse represents the structure of error responses from the API.
//
//	@Description	Represents the structure of error responses from the API.
//	@Tags			errors
type errorResponse struct {
	Status  string   `json:"status" example:"error"`
	Message string   `json:"message" example:"invalid request body"`
	Details []string `json:"details,omitempty" example:"Group name is required,Song name is required"`
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
