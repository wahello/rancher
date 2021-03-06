package bootstrap

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rancher/rancher/pkg/settings"
	"github.com/rancher/rancher/pkg/systemtemplate"
)

var (
	defaultSystemAgentInstallScript = "https://raw.githubusercontent.com/rancher/system-agent/main/install.sh"
	localAgentInstallScripts        = []string{
		"/usr/share/rancher/ui/assets/system-agent-install.sh",
		"./system-agent-install.sh",
	}
)

func InstallScript() ([]byte, error) {
	data, err := installScript()
	if err != nil {
		return nil, err
	}
	if settings.SystemAgentVersion.Get() == "" || settings.ServerURL.Get() == "" {
		return data, nil
	}
	ca := systemtemplate.CAChecksum()
	return []byte(fmt.Sprintf(`#!/usr/bin/env sh
CATTLE_AGENT_BINARY_URL="%s"
CATTLE_CA_CHECKSUM="%s"

%s
`, settings.ServerURL.Get(), ca, data)), nil
}

func installScript() ([]byte, error) {
	url := settings.SystemAgentInstallScript.Get()
	if url == "" {
		for _, localAgentInstallScript := range localAgentInstallScripts {
			script, err := ioutil.ReadFile(localAgentInstallScript)
			if !os.IsNotExist(err) {
				return script, err
			}
		}
	}

	if url == "" {
		url = defaultSystemAgentInstallScript
	}

	resp, httpErr := http.Get(url)
	if httpErr != nil {
		return nil, httpErr
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func Bootstrap(token string) ([]byte, error) {
	script, err := InstallScript()
	if err != nil {
		return nil, err
	}

	url, ca := settings.ServerURL.Get(), systemtemplate.CAChecksum()
	return []byte(fmt.Sprintf(`#!/usr/bin/env sh
CATTLE_SERVER="%s"
CATTLE_CA_CHECKSUM="%s"
CATTLE_TOKEN="%s"

%s
`, url, ca, token, script)), nil
}
