package testflag

import (
	"encoding/json"
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

type Data struct {
	Rule        *string        `json:"rule,omitempty"`
	Percentage  *float64       `json:"percentage,omitempty"`
	True        *interface{}   `json:"true,omitempty"`
	False       *interface{}   `json:"false,omitempty"`
	Default     *interface{}   `json:"default,omitempty"`
	TrackEvents *bool          `json:"trackEvents,omitempty"`
	Disable     *bool          `json:"disable,omitempty"`
	Rollout     *model.Rollout `json:"rollout,omitempty"` // nolint: lll
}

func NewFlag(data Data) model.Flag {
	jsonFlag, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	}

	var flag model.Flag
	err = json.Unmarshal(jsonFlag, &flag)
	if err != nil {
		fmt.Println(err)
	}

	return flag
}
