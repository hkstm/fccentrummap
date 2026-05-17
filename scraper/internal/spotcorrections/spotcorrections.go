package spotcorrections

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hkstm/fccentrummap/internal/geocoder"
	"github.com/hkstm/fccentrummap/internal/models"
)

var (
	ErrSpotNotFound         = errors.New("spot id not found")
	ErrTimestampURLRequired = errors.New("a full timestamped youtube.com or youtu.be URL is required")
)

type Repository interface {
	GetSpotCorrectionTarget(spotID string) (*models.SpotCorrectionTarget, error)
	UpsertSpotCorrection(c models.SpotCorrection) error
}

type PlaceIDLookup interface {
	LookupPlaceIDCoordinates(ctx context.Context, placeID string) (*geocoder.Coordinates, error)
}

type PlaceInputResolver interface {
	ResolvePlaceInput(ctx context.Context, input string) (*geocoder.Coordinates, error)
}

type CorrectionInput struct {
	SpotID                string
	SpotName              *string
	PlaceID               *string
	TimestampedYouTubeURL *string
	Hide                  bool
}

type Service struct {
	repo        Repository
	placeLookup PlaceIDLookup
}

func NewService(repo Repository, placeLookup PlaceIDLookup) *Service {
	return &Service{repo: repo, placeLookup: placeLookup}
}

func (s *Service) SaveCorrection(ctx context.Context, input CorrectionInput) (*models.SpotCorrectionTarget, error) {
	spotID := strings.TrimSpace(input.SpotID)
	if spotID == "" {
		return nil, fmt.Errorf("spot id is required")
	}
	target, err := s.repo.GetSpotCorrectionTarget(spotID)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, fmt.Errorf("%w: %s", ErrSpotNotFound, spotID)
	}

	correction := models.SpotCorrection{SpotID: spotID}
	if target.Correction != nil {
		correction = *target.Correction
	}
	correction.SpotID = spotID

	if input.SpotName != nil {
		if v := strings.TrimSpace(*input.SpotName); v != "" {
			correction.SpotName = &v
		}
	}

	if input.PlaceID != nil {
		if v := strings.TrimSpace(*input.PlaceID); v != "" {
			if s.placeLookup == nil {
				return nil, fmt.Errorf("cannot resolve corrected place input %q: place lookup is not configured", v)
			}
			coords, err := resolvePlaceCorrectionInput(ctx, s.placeLookup, v)
			if err != nil {
				return nil, fmt.Errorf("cannot resolve corrected place input %q for spotId %s: %w", v, spotID, err)
			}
			resolvedPlaceID := strings.TrimSpace(coords.PlaceID)
			if resolvedPlaceID == "" {
				resolvedPlaceID = v
			}
			correction.PlaceID = &resolvedPlaceID
			correction.Latitude = &coords.Latitude
			correction.Longitude = &coords.Longitude
		}
	}

	if input.Hide {
		correction.Hidden = true
	}

	if input.TimestampedYouTubeURL != nil {
		if v := strings.TrimSpace(*input.TimestampedYouTubeURL); v != "" {
			seconds, err := ParseTimestampedYouTubeURL(v)
			if err != nil {
				return nil, err
			}
			correction.YouTubeTimestampSeconds = &seconds
		}
	}

	if err := s.repo.UpsertSpotCorrection(correction); err != nil {
		return nil, err
	}
	return s.repo.GetSpotCorrectionTarget(spotID)
}

func resolvePlaceCorrectionInput(ctx context.Context, lookup PlaceIDLookup, url string) (*geocoder.Coordinates, error) {
	// 1. Automatically parse and trim the token out of the URL
	token := ExtractTokenFromURL(url)
	if token == "" {
		return nil, fmt.Errorf("error: Could not find a valid location token" +
			" inside the provided URL")
	}

	// 2. Decode the token purely offline
	placeID, err := DecodeUrlTokenToPlaceID(token)
	if err != nil {
		return nil, fmt.Errorf("Decoding error: %v\n", err)
	}

	return lookup.LookupPlaceIDCoordinates(ctx, placeID)
}

// DecodeUrlTokenToPlaceID takes the hex token from the URL and converts it to a Place ID offline.
func DecodeUrlTokenToPlaceID(token string) (string, error) {
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid token format")
	}

	hex1 := strings.TrimPrefix(parts[0], "0x")
	hex2 := strings.TrimPrefix(parts[1], "0x")

	var machineID, featureID uint64
	if _, err := fmt.Sscanf(hex1, "%x", &machineID); err != nil {
		return "", err
	}
	if _, err := fmt.Sscanf(hex2, "%x", &featureID); err != nil {
		return "", err
	}

	// A valid Place ID is exactly 20 bytes long
	buf := make([]byte, 20)

	// --- Outer Wrapper ---
	buf[0] = 0x0A // Field 1, Wire Type 2 (Length-delimited)
	buf[1] = 0x12 // Length: 18 bytes

	// --- Inner Payload ---
	buf[2] = 0x09 // Field 1, Wire Type 1 (Fixed64)
	binary.LittleEndian.PutUint64(buf[3:11], machineID)

	buf[11] = 0x11 // Field 2, Wire Type 1 (Fixed64)
	binary.LittleEndian.PutUint64(buf[12:20], featureID)

	// Encode using URL-Safe Base64 with strictly NO padding characters ('=')
	// The prefix "ChI" naturally emerges from encoding the 0x0A, 0x12, 0x09 bytes.
	placeID := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(buf)

	return placeID, nil
}

func ExtractTokenFromURL(url string) string {
	re := regexp.MustCompile(`1s(0x[0-9a-fA-F]+:0x[0-9a-fA-F]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func ParseTimestampedYouTubeURL(raw string) (int64, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return 0, ErrTimestampURLRequired
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return 0, ErrTimestampURLRequired
	}
	host := strings.ToLower(u.Host)
	switch host {
	case "youtube.com", "www.youtube.com", "m.youtube.com":
		if strings.TrimSpace(u.Query().Get("v")) == "" && !strings.HasPrefix(u.Path, "/embed/") && !strings.HasPrefix(u.Path, "/shorts/") {
			return 0, ErrTimestampURLRequired
		}
	case "youtu.be", "www.youtu.be":
		if strings.Trim(strings.TrimSpace(u.Path), "/") == "" {
			return 0, ErrTimestampURLRequired
		}
	default:
		return 0, ErrTimestampURLRequired
	}

	for _, key := range []string{"t", "start", "time_continue"} {
		value := strings.TrimSpace(u.Query().Get(key))
		if value == "" {
			continue
		}
		seconds, err := parseYouTubeTimestamp(value)
		if err != nil || seconds < 0 {
			return 0, ErrTimestampURLRequired
		}
		return seconds, nil
	}
	return 0, ErrTimestampURLRequired
}

var timestampPattern = regexp.MustCompile(`(?i)^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s?)?$`)

func parseYouTubeTimestamp(value string) (int64, error) {
	v := strings.TrimSpace(strings.TrimPrefix(value, "#"))
	if v == "" {
		return 0, ErrTimestampURLRequired
	}
	if n, err := strconv.ParseInt(v, 10, 64); err == nil {
		return n, nil
	}
	matches := timestampPattern.FindStringSubmatch(v)
	if matches == nil {
		return 0, ErrTimestampURLRequired
	}
	var total int64
	multipliers := []int64{3600, 60, 1}
	matchedPart := false
	for i, multiplier := range multipliers {
		part := matches[i+1]
		if part == "" {
			continue
		}
		matchedPart = true
		n, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return 0, ErrTimestampURLRequired
		}
		total += n * multiplier
	}
	if !matchedPart {
		return 0, ErrTimestampURLRequired
	}
	return total, nil
}
