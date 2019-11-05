package main

import (
    "errors"
    "fmt"
    "os"
    "encoding/json"
    "strings"
    "sort"

    "code.cloudfoundry.org/cli/plugin"
)

// BToMb is used to convert bytes to mb
var BToMb int64 = 1048576

// Instance struct used to hold the stats returned from cf curl
type Instance struct {
    State string `json:"state"`
    Stats struct {
	Name        string `json:"name"`
        Host        string `json:"host"`
        Port        int    `json:"port"`
        MemoryQuota int64  `json:"mem_quota"`
        Usage       struct {
            CPU    float64 `json:"cpu"`
            Memory int64   `json:"mem"`
            } `json:"usage"`
        } `json:"stats"`
}

// AppStats struct used to handle multiple instances in cf curl response
type AppStats struct {
    AppInstances []string
    AppData map[string]Instance
}

// HandleError will print the error and exit when err is not nil
func HandleError(err error) {
    if err != nil {
	fmt.Fprintln(os.Stderr, "error: ", err)
	os.Exit(1)
    }
}

// CfInstances plugin
type CfInstances struct{}

// getAppName returns the application name from positional args
func (c *CfInstances) getAppName(args []string) (string, error) {
    if len(args) < 2 {
	return "", errors.New("missing application name")
    }
    return args[1], nil
}

// getAppStats returns the app's instances and instance info from stats API endpoint
func (c *CfInstances) getAppStats(cliConnection plugin.CliConnection, appName string) (AppStats, error) {
    app, err := cliConnection.GetApp(appName)
    HandleError(err)

    url := fmt.Sprintf("/v2/apps/%s/stats", app.Guid)
    output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", url)
    HandleError(err)

    var stats AppStats
    err = json.Unmarshal([]byte(strings.Join(output, "")), &stats.AppData)
    HandleError(err)

    var instances []string
    for key, _ := range stats.AppData {
        instances = append(instances, key)
    }
    sort.StringSlice(instances).Sort()
    stats.AppInstances = instances

    return stats, err
}

func (c *CfInstances) Run(cliConnection plugin.CliConnection, args []string) {
    if len(args) < 2 {
        fmt.Println("ERROR: App name required to get stats")
    }

    appName, err := c.getAppName(args)
    HandleError(err)

    appStats, err := c.getAppStats(cliConnection, appName)
    HandleError(err)

    for ai, index := range appStats.AppInstances {
    fmt.Printf("\nInstance: %v\nName: %v\nHost: %v\nPort: %v\nMemory: %vMB / %vMB\n",
        ai,
        appStats.AppData[index].Stats.Name,
        appStats.AppData[index].Stats.Host,
        appStats.AppData[index].Stats.Port,
        appStats.AppData[index].Stats.Usage.Memory/BToMb,
        appStats.AppData[index].Stats.MemoryQuota/BToMb)
    }
    fmt.Println("")
}

// GetMetadata provides plugin information
func (c *CfInstances) GetMetadata() plugin.PluginMetadata {
    return plugin.PluginMetadata{
        Name: "instances",
        Version: plugin.VersionType{
            Major: 0,
            Minor: 4,
            Build: 0,
        },
        Commands: []plugin.Command{
            {
                Name:     "instances",
                HelpText: "Grabs instance informaiton for the provided app like IP/Port of AI",
                UsageDetails: plugin.Usage{
                Usage: "instances\n   cf instances APP-NAME",
                },
            },
        },
    }
}


func main() {
	plugin.Start(new(CfInstances))
}
