package http

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vadimbarashkov/online-song-library/internal/entity"
)

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name           string
		values         url.Values
		wantPagination entity.Pagination
	}{
		{
			name: "not-empty offset and limit",
			values: url.Values{
				"offset": []string{"40"},
				"limit":  []string{"10"},
			},
			wantPagination: entity.Pagination{
				Offset: 40,
				Limit:  10,
			},
		},
		{
			name: "empty offset andl limit",
			wantPagination: entity.Pagination{
				Offset: entity.DefaultOffset,
				Limit:  entity.DefaultLimit,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &http.Request{
				URL: &url.URL{
					RawQuery: tt.values.Encode(),
				},
			}

			pagination := parsePagination(r)

			assert.Equal(t, tt.wantPagination.Offset, pagination.Offset)
			assert.Equal(t, tt.wantPagination.Limit, pagination.Limit)
		})
	}
}

func TestParseSongFilters(t *testing.T) {
	tests := []struct {
		name            string
		values          url.Values
		expectedFilters []entity.SongFilter
	}{
		{
			name:            "no filters",
			values:          url.Values{},
			expectedFilters: []entity.SongFilter{},
		},
		{
			name: "multiple filters",
			values: url.Values{
				"groupName":         []string{"Test Group"},
				"name":              []string{"Test Song"},
				"releaseYear":       []string{"02.01.2018"},
				"releaseDate":       []string{"01.01.2019"},
				"releaseDateAfter":  []string{"01.01.2015"},
				"releaseDateBefore": []string{"01.01.2021"},
				"text":              []string{"Test Text"},
			},
			expectedFilters: []entity.SongFilter{
				{Field: entity.SongGroupNameFilterField, Value: "Test Group"},
				{Field: entity.SongNameFilterField, Value: "Test Song"},
				{Field: entity.SongReleaseYearFilterField, Value: parseDate("02.01.2018")},
				{Field: entity.SongReleaseDateFilterField, Value: parseDate("01.01.2019")},
				{Field: entity.SongReleaseDateAfterFilterField, Value: parseDate("01.01.2015")},
				{Field: entity.SongReleaseDateBeforeFilterField, Value: parseDate("01.01.2021")},
				{Field: entity.SongTextFilterField, Value: "Test Text"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				URL: &url.URL{
					RawQuery: tt.values.Encode(),
				},
			}

			filters := parseSongFilters(req)

			assert.Len(t, filters, len(tt.expectedFilters))

			for i, filter := range filters {
				assert.Equal(t, tt.expectedFilters[i].Field, filter.Field)
				assert.Equal(t, tt.expectedFilters[i].Value, filter.Value)

			}
		})
	}
}

func parseDate(dateStr string) time.Time {
	date, _ := time.Parse("02.01.2006", dateStr)
	return date
}
