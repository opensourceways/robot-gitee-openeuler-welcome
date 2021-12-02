package main

import (
	"fmt"

	libconfig "github.com/opensourceways/community-robot-lib/config"
)

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
}

func (c *configuration) configFor(org, repo string) *botConfig {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	v := make([]libconfig.IPluginForRepo, len(items))
	for i := range items {
		v[i] = &items[i]
	}

	if i := libconfig.FindConfig(org, repo, v); i >= 0 {
		return &items[i]
	}
	return nil
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) SetDefault() {
	if c == nil {
		return
	}

	Items := c.ConfigItems
	for i := range Items {
		Items[i].setDefault()
	}
}

type botConfig struct {
	libconfig.PluginForRepo

	// CommunityName is the name of community
	CommunityName string `json:"community_name" required:"true"`

	// CommandLink is the link to command help document page.
	CommandLink string `json:"command_link" required:"true"`

	// SigFilePath is file path and the file includes information about
	// Special Interest Groups (SIG) in the community.
	// The format is org/repo/branch:path
	SigFilePath string     `json:"sig_file_path" required:"true"`
	sigFile     fileOfRepo `json:"-"`
}

func (c *botConfig) setDefault() {
}

func (c *botConfig) validate() error {
	if c.CommunityName == "" {
		return fmt.Errorf("the community_name configuration can not be empty")
	}
	if c.CommandLink == "" {
		return fmt.Errorf("the command_link configuration can not be empty")
	}
	return c.PluginForRepo.Validate()
}

func parseSigFilePath(p string) (fileOfRepo, error) {
	return fileOfRepo{}, nil
}

type fileOfRepo struct {
	org    string
	repo   string
	branch string
	path   string
}
