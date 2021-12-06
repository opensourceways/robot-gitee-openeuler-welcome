package main

import (
	"fmt"
	"strings"

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

	// CheckObsMetalOriginFile check whether obs meta related files are modified when PR is created
	CheckObsMetalOriginFile bool `json:"check_obs_meta_origin_file,omitempty"`
	//ObsMetaOriginFile file name related to obs meta
	ObsMetaOriginFile string `json:"obs_meta_origin_file,omitempty"`
	//GuideURL the guid url
	GuideURL string `json:"guide_url,omitempty"`
	//ObsMetaConfig the repository configuration for storing obs meta information
	ObsMetaConfig obsMetaConfig `json:"obs_meta_config,omitempty"`
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

	if c.CheckObsMetalOriginFile {
		if c.ObsMetaOriginFile == "" {
			return fmt.Errorf("the obs_meta_origin_file configuration can not be empty when check_obs_meta_file is true")
		}

		if c.GuideURL == "" {
			return fmt.Errorf("the guid_url configuration can not be empty when check_obs_meta_file is true")
		}

		if err := c.ObsMetaConfig.validate(); err != nil {
			return err
		}
	}

	if err := c.parseSigFilePath(); err != nil {
		return err
	}

	return c.PluginForRepo.Validate()
}

func (c *botConfig) parseSigFilePath() error {
	p := c.SigFilePath

	v := strings.Split(p, ":")
	if len(v) != 2 {
		return fmt.Errorf("invalid sig_file_path:%s", p)
	}

	v1 := strings.Split(v[0], "/")
	if len(v1) != 3 {
		return fmt.Errorf("invalid sig_file_path:%s", p)
	}

	c.sigFile = fileOfRepo{
		org:    v1[0],
		repo:   v1[1],
		branch: v1[2],
		path:   v[1],
	}

	return nil
}

type fileOfRepo struct {
	org    string
	repo   string
	branch string
	path   string
}

type obsMetaConfig struct {
	Owner  string `json:"owner" required:"true"`
	Repo   string `json:"repo" required:"true"`
	Branch string `json:"branch" required:"true"`
}

func (omi obsMetaConfig) validate() error {
	if omi.Owner == "" {
		return fmt.Errorf("the owner configuration can not be empty")
	}

	if omi.Repo == "" {
		return fmt.Errorf("the repo configuration can not be empty")
	}

	if omi.Branch == "" {
		return fmt.Errorf("the branch configuration can not be empty")
	}

	return nil
}
