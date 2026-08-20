package homeassistant

import (
	"fmt"

	"goalfeed/config"
	"goalfeed/utils"
)

// validateOutboundHAURL is called immediately before any outbound request to
// Home Assistant is built, at every call site that would otherwise attach the
// configured long-lived access token as a Bearer header. It exists so that a
// poisoned or hand-edited home_assistant.url - set via the runtime config API,
// an environment variable, or a directly edited config.yaml - cannot cause
// Goalfeed to leak that token off the local network, even if it slipped past
// validation elsewhere (or was never validated at all).
//
// See utils.ValidateHomeAssistantURL for the exact rules. The
// home_assistant.allow_remote_url config key opts out of the private-address
// requirement for a genuinely remote HA instance; scheme and host validation
// always apply regardless.
func validateOutboundHAURL(rawURL string) error {
	allowRemote := config.GetBool("home_assistant.allow_remote_url")
	if err := utils.ValidateHomeAssistantURL(rawURL, allowRemote); err != nil {
		return fmt.Errorf("refusing to send Home Assistant request: %w", err)
	}
	return nil
}
