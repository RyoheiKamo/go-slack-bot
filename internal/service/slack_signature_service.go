package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

type SlackSignatureService struct {
	signingSecret string
}

func NewSlackSignatureService(signingSecret string) *SlackSignatureService {
	return &SlackSignatureService{
		signingSecret: signingSecret,
	}
}

func (s *SlackSignatureService) Verify(
	timestamp string,
	signature string,
	body []byte,
) bool {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}

	// リプレイ攻撃対策
	if time.Now().Unix()-ts > 60*5 {
		return false
	}

	baseString := fmt.Sprintf(
		"v0:%s:%s",
		timestamp,
		string(body),
	)

	mac := hmac.New(sha256.New, []byte(s.signingSecret))
	mac.Write([]byte(baseString))

	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal(
		[]byte(expectedSignature),
		[]byte(signature),
	)
}
