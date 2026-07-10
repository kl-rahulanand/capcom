package gantry

import (
	"encoding/json"

	"capcom/internal/domain"
)

type gantryAgent struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Display     string         `json:"displayName"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Raw         map[string]any `json:"-"`
}

func (g *gantryAgent) UnmarshalJSON(data []byte) error {
	type alias gantryAgent
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*g = gantryAgent(decoded)
	g.Raw = raw
	return nil
}

func (g gantryAgent) DisplayName() string {
	if g.Display != "" {
		return g.Display
	}
	if g.Name != "" {
		return g.Name
	}
	return g.ID
}

func (g gantryAgent) StatusDomain() domain.AgentStatus {
	switch g.Status {
	case "active", "enabled", "running":
		return domain.AgentStatusEnabled
	case "disabled":
		return domain.AgentStatusDisabled
	default:
		return domain.AgentStatusUnknown
	}
}

type gantryAccess struct {
	AgentID    string            `json:"agentId"`
	Selections []gantrySelection `json:"selections"`
	Raw        map[string]any    `json:"-"`
}

func (g *gantryAccess) UnmarshalJSON(data []byte) error {
	type alias gantryAccess
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*g = gantryAccess(decoded)
	g.Raw = raw
	return nil
}

type gantrySelection struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}
