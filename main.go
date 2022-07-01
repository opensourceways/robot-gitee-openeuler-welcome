package main

import (
	"flag"
	"net/url"
	"os"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/opensourceways/community-robot-lib/logrusutil"
	liboptions "github.com/opensourceways/community-robot-lib/options"
	framework "github.com/opensourceways/community-robot-lib/robot-gitee-framework"
	"github.com/opensourceways/community-robot-lib/secret"
	cache "github.com/opensourceways/repo-file-cache/sdk"
	"github.com/sirupsen/logrus"
)

type options struct {
	service       liboptions.ServiceOptions
	gitee         liboptions.GiteeOptions
	cacheEndpoint string
	maxRetries    int
}

func (o *options) Validate() error {
	if _, err := url.ParseRequestURI(o.cacheEndpoint); err != nil {
		return err
	}

	if err := o.service.Validate(); err != nil {
		return err
	}

	return o.gitee.Validate()
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options

	o.gitee.AddFlags(fs)
	o.service.AddFlags(fs)
	retry := 3
	fs.StringVar(&o.cacheEndpoint, "cache-endpoint", "", "The endpoint of repo file cache")
	fs.IntVar(&o.maxRetries, "max-retries", retry, "The number of failed retry attempts to call the cache api")

	_ = fs.Parse(args)
	return o
}

func main() {
	logrusutil.ComponentInit(botName)

	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	secretAgent := new(secret.Agent)
	if err := secretAgent.Start([]string{o.gitee.TokenPath}); err != nil {
		logrus.WithError(err).Fatal("Error starting secret agent.")
	}

	defer secretAgent.Stop()

	c := giteeclient.NewClient(secretAgent.GetTokenGenerator(o.gitee.TokenPath))
	s := cache.NewSDK(o.cacheEndpoint, o.maxRetries)

	p := newRobot(c, s)

	framework.Run(p, o.service)
}
