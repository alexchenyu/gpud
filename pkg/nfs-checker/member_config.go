package nfschecker

import (
	"errors"
)

// MemberConfig configures a "single" NFS checker.
type MemberConfig struct {
	// Config is the common configuration for all the NFS checker group
	// members, which then translates into a single NFS checker.
	Config

	// ID is a unique ID for the writer, which is used as a file name
	// in the directory. This helps avoiding race conditions when writing
	// to the same file.
	//
	// This can be set to the machine ID from each host.
	// This ID just needs to be different from other writers
	// that mounts the same NFS mount point
	ID string `json:"id"`
}

// MemberConfigs is a list of MemberConfig.
type MemberConfigs []MemberConfig

// Validate validates the member configurations.
func (cfgs MemberConfigs) Validate() error {
	for _, cfg := range cfgs {
		if err := cfg.Validate(); err != nil {
			return err
		}
	}
	return nil
}

var ErrIDEmpty = errors.New("ID is empty")

// Validate validates the configuration.
func (c *MemberConfig) Validate() error {
	if err := c.Config.ValidateAndMkdir(); err != nil {
		return err
	}

	if c.ID == "" {
		return ErrIDEmpty
	}

	return nil
}
