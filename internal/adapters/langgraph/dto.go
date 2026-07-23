package langgraph

import "encoding/json"

type serverInfo struct {
	Version            string         `json:"version"`
	LangGraphPyVersion string         `json:"langgraph_py_version"`
	Flags              map[string]any `json:"flags"`
	Metadata           map[string]any `json:"metadata"`
}

type assistant struct {
	AssistantID string         `json:"assistant_id"`
	GraphID     string         `json:"graph_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Version     int            `json:"version"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
	Config      map[string]any `json:"config"`
	Context     map[string]any `json:"context"`
	Metadata    map[string]any `json:"metadata"`
	Raw         map[string]any `json:"-"`
}

func (a *assistant) UnmarshalJSON(data []byte) error {
	type alias assistant
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*a = assistant(decoded)
	a.Raw = raw
	return nil
}

type thread struct {
	ThreadID     string         `json:"thread_id"`
	Status       string         `json:"status"`
	CreatedAt    string         `json:"created_at"`
	UpdatedAt    string         `json:"updated_at"`
	StateUpdated string         `json:"state_updated_at"`
	Metadata     map[string]any `json:"metadata"`
	Config       map[string]any `json:"config"`
	Values       map[string]any `json:"values"`
	Interrupts   map[string]any `json:"interrupts"`
	Raw          map[string]any `json:"-"`
}

func (t *thread) UnmarshalJSON(data []byte) error {
	type alias thread
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*t = thread(decoded)
	t.Raw = raw
	return nil
}

type run struct {
	RunID                string         `json:"run_id"`
	ThreadID             string         `json:"thread_id"`
	AssistantID          string         `json:"assistant_id"`
	Status               string         `json:"status"`
	CreatedAt            string         `json:"created_at"`
	UpdatedAt            string         `json:"updated_at"`
	Metadata             map[string]any `json:"metadata"`
	Kwargs               map[string]any `json:"kwargs"`
	MultitaskStrategy    string         `json:"multitask_strategy"`
	LangSmithSessionName string         `json:"langsmith_session_name"`
	Raw                  map[string]any `json:"-"`
}

func (r *run) UnmarshalJSON(data []byte) error {
	type alias run
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*r = run(decoded)
	r.Raw = raw
	return nil
}
