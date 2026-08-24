package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"gogogo/helpers"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// FacebookPixelService wraps Meta Conversions API event sending.
type FacebookPixelService struct {
	httpClient *http.Client
}

type fbUserData struct {
	ExternalID      []string `json:"external_id,omitempty"`
	ClientIPAddress string   `json:"client_ip_address,omitempty"`
	ClientUserAgent string   `json:"client_user_agent,omitempty"`
	FBC             string   `json:"fbc,omitempty"`
	FBP             string   `json:"fbp,omitempty"`
}

type fbCustomData struct {
	Currency string  `json:"currency,omitempty"`
	Value    float64 `json:"value,omitempty"`
	OrderID  string  `json:"order_id,omitempty"`
}

type fbEvent struct {
	EventName      string        `json:"event_name"`
	EventTime      int64         `json:"event_time"`
	EventID        string        `json:"event_id,omitempty"`
	ActionSource   string        `json:"action_source"`
	EventSourceURL string        `json:"event_source_url,omitempty"`
	UserData       fbUserData    `json:"user_data"`
	CustomData     *fbCustomData `json:"custom_data,omitempty"`
}

type fbPayload struct {
	Data          []fbEvent `json:"data"`
	TestEventCode string    `json:"test_event_code,omitempty"`
}

type fbResponse struct {
	EventsReceived int           `json:"events_received"`
	Messages       []interface{} `json:"messages"`
	FBTraceID      string        `json:"fbtrace_id"`
}

// EventContext holds optional browser context for better matching.
type EventContext struct {
	ClientIP  string
	UserAgent string
	FBC       string
	FBP       string
	SourceURL string
}

func hashID(id string) string {
	if id == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(id))))
	return hex.EncodeToString(sum[:])
}

func normalizePixelTarget(cfg helpers.FacebookPixelTargetConfig) (helpers.FacebookPixelTargetConfig, bool) {
	cfg.PixelID = strings.TrimSpace(cfg.PixelID)
	cfg.AccessToken = strings.TrimSpace(cfg.AccessToken)
	cfg.TestEventCode = strings.TrimSpace(cfg.TestEventCode)
	cfg.APIVersion = strings.TrimSpace(cfg.APIVersion)
	if cfg.APIVersion == "" {
		cfg.APIVersion = "v18.0"
	}
	return cfg, cfg.PixelID != "" && cfg.AccessToken != ""
}

func (s *FacebookPixelService) getConfigs() []helpers.FacebookPixelTargetConfig {
	holder := helpers.GetCfgInstance()
	if holder == nil || holder.Conf == nil {
		return nil
	}

	cfg := holder.Conf.FacebookPixel
	configs := make([]helpers.FacebookPixelTargetConfig, 0, 1+len(cfg.AdditionalPixels))

	if primary, ok := normalizePixelTarget(helpers.FacebookPixelTargetConfig{
		PixelID:        cfg.PixelID,
		AccessToken:    cfg.AccessToken,
		TestEventCode:  cfg.TestEventCode,
		APIVersion:     cfg.APIVersion,
		InviterUserIDs: cfg.InviterUserIDs,
	}); ok {
		configs = append(configs, primary)
	}

	for _, extra := range cfg.AdditionalPixels {
		if normalized, ok := normalizePixelTarget(extra); ok {
			configs = append(configs, normalized)
		}
	}

	return configs
}

func (s *FacebookPixelService) findConfigByInviterID(inviterUserID uint64) (*helpers.FacebookPixelTargetConfig, bool) {
	configs := s.getConfigs()
	if len(configs) == 0 {
		return nil, false
	}

	for _, cfg := range configs {
		for _, allowedID := range cfg.InviterUserIDs {
			if allowedID == inviterUserID {
				matched := cfg
				return &matched, true
			}
		}
	}

	return nil, false
}

func (s *FacebookPixelService) describeTargets() string {
	configs := s.getConfigs()
	if len(configs) == 0 {
		return "[]"
	}

	parts := make([]string, 0, len(configs))
	for _, cfg := range configs {
		parts = append(parts, fmt.Sprintf("{pixel_id:%s inviter_user_ids:%v}", cfg.PixelID, cfg.InviterUserIDs))
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func (s *FacebookPixelService) sendEvent(cfg helpers.FacebookPixelTargetConfig, event fbEvent) error {
	if cfg.PixelID == "" || cfg.AccessToken == "" {
		return nil
	}

	payload := fbPayload{Data: []fbEvent{event}}
	if cfg.TestEventCode != "" && os.Getenv("dev") == "1" {
		payload.TestEventCode = cfg.TestEventCode
	}

	fmt.Printf(
		"[FacebookPixel] sending event=%s pixel_id=%s has_test_event_code=%t\n",
		event.EventName,
		cfg.PixelID,
		payload.TestEventCode != "",
	)

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	endpoint := fmt.Sprintf(
		"https://graph.facebook.com/%s/%s/events?access_token=%s",
		cfg.APIVersion,
		cfg.PixelID,
		cfg.AccessToken,
	)

	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("CAPI status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result fbResponse
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			fmt.Printf("[FacebookPixel] pixel_id=%s non-json success body=%s\n", cfg.PixelID, string(respBody))
		} else {
			fmt.Printf(
				"[FacebookPixel] pixel_id=%s events_received=%d messages=%v fbtrace_id=%s\n",
				cfg.PixelID,
				result.EventsReceived,
				result.Messages,
				result.FBTraceID,
			)
		}
	}

	return nil
}

func (s *FacebookPixelService) trackAsync(inviterUserID uint64, event fbEvent) {
	cfg, ok := s.findConfigByInviterID(inviterUserID)
	if !ok {
		fmt.Printf(
			"[FacebookPixel] skip %s: no pixel target matched inviter_user_id=%d available_targets=%s\n",
			event.EventName,
			inviterUserID,
			s.describeTargets(),
		)
		return
	}
	fmt.Printf(
		"[FacebookPixel] dispatch event=%s event_id=%s inviter_user_id=%d pixel_id=%s\n",
		event.EventName,
		event.EventID,
		inviterUserID,
		cfg.PixelID,
	)

	go func(target helpers.FacebookPixelTargetConfig) {
		if err := s.sendEvent(target, event); err != nil {
			fmt.Printf("[FacebookPixel] track %s failed pixel_id=%s inviter_user_id=%d: %v\n", event.EventName, target.PixelID, inviterUserID, err)
			return
		}
		fmt.Printf("[FacebookPixel] track %s ok pixel_id=%s inviter_user_id=%d event_id=%s\n", event.EventName, target.PixelID, inviterUserID, event.EventID)
	}(*cfg)
}

func buildUserData(userID uint64, ctx EventContext) fbUserData {
	ud := fbUserData{
		ClientIPAddress: ctx.ClientIP,
		ClientUserAgent: ctx.UserAgent,
		FBC:             ctx.FBC,
		FBP:             ctx.FBP,
	}
	if userID > 0 {
		ud.ExternalID = []string{hashID(fmt.Sprintf("%d", userID))}
	}
	return ud
}

func (s *FacebookPixelService) TrackCompleteRegistration(userID uint64, inviterUserID uint64, ctx EventContext) {
	event := fbEvent{
		EventName:      "CompleteRegistration",
		EventTime:      time.Now().Unix(),
		EventID:        fmt.Sprintf("reg-%d-%s", userID, uuid.NewString()),
		ActionSource:   "website",
		EventSourceURL: ctx.SourceURL,
		UserData:       buildUserData(userID, ctx),
	}
	s.trackAsync(inviterUserID, event)
}

// TrackPurchase uses orderID as event_id so browser pixel and CAPI can deduplicate.
func (s *FacebookPixelService) TrackPurchase(userID uint64, inviterUserID uint64, amount float64, currency, orderID string, ctx EventContext) {
	event := fbEvent{
		EventName:      "Purchase",
		EventTime:      time.Now().Unix(),
		EventID:        orderID,
		ActionSource:   "website",
		EventSourceURL: ctx.SourceURL,
		UserData:       buildUserData(userID, ctx),
		CustomData: &fbCustomData{
			Currency: currency,
			Value:    amount,
			OrderID:  orderID,
		},
	}
	s.trackAsync(inviterUserID, event)
}

var (
	facebookPixelService     *FacebookPixelService
	facebookPixelServiceOnce sync.Once
)

func GetFacebookPixelService() *FacebookPixelService {
	facebookPixelServiceOnce.Do(func() {
		facebookPixelService = &FacebookPixelService{
			httpClient: &http.Client{Timeout: 10 * time.Second},
		}
	})
	return facebookPixelService
}
