package main

import (
    "fmt"
    "bytes"
    "strings"
    "webhook"
    "encoding/json"
    "io/ioutil"
    "code.cloudfoundry.org/cli/plugin"
    "code.cloudfoundry.org/cli/plugin/models"
)

type ICDPlugin struct{}

type GitValues struct {
    GitURL string
    GitBranch string
    GitCommitID string
}

func GitInfo () (GitValues, error) {
    files, err := ioutil.ReadDir(".")
    if err != nil {
        fmt.Println(err)
    }

    for _, file := range files {
        fmt.Println(file.Name())
        if file.Name() == ".git" {
           head, err := ioutil.ReadFile(".git/HEAD")
           check(err)
           fmt.Println(string(head))
           parts := strings.Split(string(head), "/")
           branch_name := parts[len(parts) - 1]
           fmt.Println(branch_name)
           id, err := ioutil.ReadFile(".git/refs/heads/" + strings.Trim(branch_name, "\n\r \b"))
           check(err)
           fmt.Println(string(id))
           break
        }
    }

   result := GitValues {
      GitURL: "https://github.com/",
      GitBranch: "master",
      GitCommitID: "123",
   }
   return result, nil
}

func (c *ICDPlugin) Run(cliConnection plugin.CliConnection, args []string) {
    var shouldRequest bool = false
    var method string = "POST"
    if args[0] == "icd" && len(args) > 2 && args[1] == "--create-connection" {
       shouldRequest = true
       method = "POST"
    } else if args[0] == "icd" && len(args) > 1 && args[1] == "--git-info" {
       GitInfo()
    } else if args[0] == "icd" && len(args) > 2 && args[1] == "--delete-connection" {
       shouldRequest = true
       method = "DELETE"
    }

    if (shouldRequest) {
        webhook_url := args[2]
        appName := args[3]
        fmt.Println("w: %s, a: %s", webhook_url, appName)
        current_org, err := cliConnection.GetCurrentOrg()
        check(err)
        current_space, err := cliConnection.GetCurrentSpace()
        check(err)
        apiEndpoint, err := cliConnection.ApiEndpoint()
        check(err)
        current_app, err := cliConnection.GetApp(appName)
        check(err)
        git_info, err := GitInfo()
        check(err)
        type Message struct {
            Org plugin_models.Organization
            Space plugin_models.Space
            App plugin_models.GetAppModel
            ApiEndpoint string
            Method string
            GitData GitValues
        }
        amp := Message {
            Org: current_org,
            Space: current_space,
            App: current_app,
            ApiEndpoint: apiEndpoint,
            Method: method,
            GitData: git_info,
        }
        fmt.Println(amp)
        js, err := json.Marshal(amp)
        check(err)
        var buf = bytes.NewBufferString(string(js))

        body := webhook.Request(webhook_url, method, buf)
        fmt.Println(body)
    }
}

func (c *ICDPlugin) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name: "IBM Continuous Delivery",
        Version: plugin.VersionType{
            Major: 0,
            Minor: 0,
            Build: 1,
        },
        MinCliVersion: plugin.VersionType{
            Major: 6,
            Minor: 7,
            Build: 0,
        },
        Commands: []plugin.Command{
            {
                Name:     "icd",
                HelpText: "IBM Continous Delivery plugin command's help text",

                // UsageDetails is optional
                // It is used to show help of usage of each command
                UsageDetails: plugin.Usage{
                    Usage: "IBM Continous Delivery:\n   cf icd",
                },
            },
        },
    }
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func main() {
    plugin.Start(new(ICDPlugin))
}
