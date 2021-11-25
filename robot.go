package main

import (
	"encoding/base64"
	"fmt"
	"strings"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	libconfig "github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	libplugin "github.com/opensourceways/community-robot-lib/giteeplugin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	botName        = "welcome"
	welcomeMessage = `Hi ***%s***, welcome to the %s Community.
I'm the Bot here serving you. You can find the instructions on how to interact with me at
<%s>.
If you have any questions, please contact the SIG: [%s](https://gitee.com/openeuler/community/tree/master/sig/%s),and any of the maintainers: %s`
)

type iClient interface {
	CreatePRComment(owner, repo string, number int32, comment string) error
	CreateIssueComment(owner, repo string, number string, comment string) error
	GetPathContent(org, repo, path, ref string) (sdk.Content, error)
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewPluginConfig() libconfig.PluginConfig {
	return &configuration{}
}

func (bot *robot) getConfig(cfg libconfig.PluginConfig, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}
	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}
	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(p libplugin.HandlerRegitster) {
	p.RegisterIssueHandler(bot.handleIssueEvent)
	p.RegisterPullRequestHandler(bot.handlePREvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, pc libconfig.PluginConfig, log *logrus.Entry) error {
	action := giteeclient.GetPullRequestAction(e)
	if action != giteeclient.PRActionOpened {
		return nil
	}

	prInfo := giteeclient.GetPRInfoByPREvent(e)
	cfg, err := bot.getConfig(pc, prInfo.Org, prInfo.Repo)
	if err != nil {
		return err
	}

	comment, err := bot.genWelcomeMessage(prInfo.Org, prInfo.Repo, prInfo.Author, cfg, log)
	if err != nil {
		return err
	}

	return bot.cli.CreatePRComment(prInfo.Org, prInfo.Repo, prInfo.Number, comment)
}

func (bot *robot) handleIssueEvent(e *sdk.IssueEvent, pc libconfig.PluginConfig, log *logrus.Entry) error {
	ew := giteeclient.NewIssueEventWrapper(e)
	if giteeclient.StatusOpen != ew.GetAction() {
		return nil
	}

	org, repo := ew.GetOrgRep()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	author := ew.GetIssueAuthor()
	number := ew.GetIssueNumber()
	comment, err := bot.genWelcomeMessage(org, repo, author, cfg, log)
	if err != nil {
		return err
	}

	return bot.cli.CreateIssueComment(org, repo, number, comment)
}

func (bot robot) genWelcomeMessage(org, repo, author string, cfg *botConfig, log *logrus.Entry) (string, error) {
	path := fmt.Sprintf("%s/%s", org, repo)
	sigName := bot.repoSigName(path, cfg, log)
	if sigName == "" {
		return "", fmt.Errorf("cant get sig name by %s", path)
	}

	v, err := bot.getMaintainers(sigName, cfg)
	if err != nil {
		return "", err
	}

	maintainerMsg := ""
	if len(v) > 0 {
		maintainerMsg = fmt.Sprintf("**@%s**", strings.Join(v, "** ,**@"))
	}

	return fmt.Sprintf(welcomeMessage, author, cfg.CommunityName, cfg.CommandLink, sigName, sigName, maintainerMsg), nil
}

func (bot robot) repoSigName(repoPath string, cfg *botConfig, log *logrus.Entry) string {
	c, err := bot.getPathContent(cfg.CommunityName, cfg.CommunityRepository, cfg.SigFilePath)
	if err != nil {
		log.Error(err)
		return ""
	}

	// intercept the sig configuration item string containing a repository from the yaml file
	s := string(c)
	keyOfName := "- name: "
	repoSigConfig := interceptString(s, keyOfName, repoPath)
	if repoSigConfig == "" {
		return ""
	}

	// intercept the sig name
	keyOfRepos := "repositories:"
	s = interceptString(repoSigConfig, keyOfName, keyOfRepos)
	if s == "" {
		return ""
	}

	s = strings.TrimPrefix(s, keyOfName)
	s = strings.TrimSuffix(s, keyOfRepos)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", "")

	return s
}

func (bot *robot) getMaintainers(sig string, cfg *botConfig) ([]string, error) {
	path := fmt.Sprintf("sig/%s/OWNERS", sig)
	c, err := bot.getPathContent(cfg.CommunityName, cfg.CommunityRepository, path)
	if err != nil {
		return nil, err
	}

	var m struct {
		Maintainers []string `yaml:"maintainers"`
	}
	if err = yaml.Unmarshal(c, &m); err != nil {
		return nil, err
	}

	return m.Maintainers, nil
}

func (bot *robot) getPathContent(owner, repo, path string) ([]byte, error) {
	content, err := bot.cli.GetPathContent(owner, repo, path, "master")
	if err != nil {
		return nil, err
	}

	c, err := base64.StdEncoding.DecodeString(content.Content)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// interceptString intercept the substring between the last matching `start` and the first matching `end` in a string.
// for example: enter abab12389898 ab 98 will return ab123898".
func interceptString(s, start, end string) string {
	if s == "" || start == "" || end == "" {
		return s
	}

	eIdx := strings.Index(s, end)
	if eIdx == -1 {
		return ""
	}

	eIdx += len(end)
	sIdx := strings.LastIndex(s[:eIdx], start)
	if eIdx == -1 {
		return ""
	}

	return s[sIdx:eIdx]
}
