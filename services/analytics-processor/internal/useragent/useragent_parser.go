package useragent

import (
	"github.com/medama-io/go-useragent"
	"github.com/medama-io/go-useragent/agents"
)

type UserAgent struct {
	Browser      agents.Browser
	OS           agents.OS
	Device       agents.Device
	Version      []byte
	VersionIndex int
}

type UserAgentParser struct {
	ua *useragent.Parser
}

func NewUserAgentParser() *UserAgentParser {
	return &UserAgentParser{
		ua: useragent.NewParser(),
	}
}

func (uap *UserAgentParser) Parse(userAgent string) *UserAgent {
	if userAgent == "" {
		return &UserAgent{
			Device: "Unknown",
		}
	}

	ua := uap.ua.Parse(userAgent)

	return &UserAgent{
		Browser:      ua.Browser(),
		OS:           ua.OS(),
		Device:       ua.Device(),
		Version:      []byte(ua.BrowserVersion()),
		VersionIndex: len(ua.BrowserVersion()),
	}
}
