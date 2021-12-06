package main

import (
	"fmt"
	"strings"

	"gitee.com/openeuler/go-gitee/gitee"
	"k8s.io/apimachinery/pkg/util/sets"
)

const autoAddPrjMsg = `Since you have added a item to the %s file, we will automatically generate a default package in project openEuler:Factory on OBS cluster for you.
If you need a more customized configuration, you can configure it according to the following [instructions](%s)`

func (bot *robot) handleCheckObsMetaFile(org, repo string, number int32, cfg *botConfig) error {
	if !cfg.CheckObsMetalOriginFile {
		return nil
	}

	file, err := bot.getObsMetaFileChanged(org, repo, number, cfg.ObsMetaOriginFile)
	if err != nil || file == nil {
		return err
	}

	proNames := getProjectNamesFromChangeFile(file)
	if len(proNames) == 0 {
		return nil
	}

	exist, err := bot.obsMetaRepoExistAllProjectNames(proNames, cfg)
	if err != nil || exist {
		return err
	}

	return bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(autoAddPrjMsg, cfg.ObsMetaOriginFile, cfg.GuideURL))
}

func (bot *robot) getObsMetaFileChanged(org, repo string, number int32, path string) (*gitee.PullRequestFiles, error) {
	changes, err := bot.cli.GetPullRequestChanges(org, repo, number)
	if err != nil {
		return nil, err
	}

	for _, file := range changes {
		if strings.Contains(file.Filename, path) {
			return &file, nil
		}
	}
	return nil, nil
}

func (bot *robot) obsMetaRepoExistAllProjectNames(pNames []string, cfg *botConfig) (bool, error) {
	if len(pNames) == 0 {
		return false, nil
	}

	org, repo, branch := cfg.ObsMetaConfig.Owner, cfg.ObsMetaConfig.Repo, cfg.ObsMetaConfig.Branch

	sha, err := bot.cli.GetRef(org, repo, branch)
	if err != nil {
		return false, err
	}

	tree, err := bot.cli.GetDirectoryTree(org, repo, sha, 1)
	if err != nil {
		return false, err
	}

	s := sets.NewString()
	for _, t := range tree.Tree {
		s.Insert(t.Path)
	}

	return s.HasAll(pNames...), nil
}

func getProjectNamesFromChangeFile(file *gitee.PullRequestFiles) (proNames []string) {
	str := strings.ReplaceAll(file.Patch.Diff," ","")
	diff := strings.Split(str, "\n")
	changePrefix := "+-name:"
	var validDiffs []string

	for _, v := range diff {
		if strings.HasPrefix(v, changePrefix) {
			validDiffs = append(validDiffs, v)
		}
	}

	for _, str := range validDiffs {
		proNames = append(proNames, strings.TrimPrefix(str, changePrefix))
	}

	return
}
