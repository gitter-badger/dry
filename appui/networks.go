package appui

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
	"text/template"

	godocker "github.com/fsouza/go-dockerclient"
	"github.com/moncho/dry/docker"
	"github.com/moncho/dry/ui"
)

type networksColumn struct {
	name  string // The name of the field in the struct.
	title string // Title to display in the tableHeader.
	mode  docker.SortNetworksMode
}

//DockerNetworksRenderer knows how render a container list
type DockerNetworksRenderer struct {
	columns               []networksColumn // List of columns.
	networksTableTemplate *template.Template
	networksTemplate      *template.Template
	cursor                *ui.Cursor
	daemon                docker.ContainerDaemon
	dockerInfo            string // Docker environment information
	sortMode              docker.SortNetworksMode
	networkTableStart     int
	height                int
}

//NewDockerNetworksRenderer creates a renderer for a container list
func NewDockerNetworksRenderer(daemon docker.ContainerDaemon, screenHeight int, cursor *ui.Cursor, sortMode docker.SortNetworksMode) *DockerNetworksRenderer {
	r := &DockerNetworksRenderer{}

	r.columns = []networksColumn{
		{`NetworkID`, `NETWORK ID`, docker.SortNetworksByID},
		{`Name`, `NAME`, docker.SortNetworksByName},
		{`Driver`, `DRIVER`, docker.SortNetworksByDriver},
		{`Driver`, `CONTAINERS`, docker.NoSortNetworks},
	}

	di := dockerInfo(daemon)

	r.networksTableTemplate = buildNetworkTableTemplate(di)
	r.networksTemplate = buildNetworksTemplate()
	r.cursor = cursor
	r.daemon = daemon
	r.sortMode = sortMode
	//Safe guess about how many lines from the start of screen (including network table header) before
	//networks are actually written to screen
	r.networkTableStart = 10
	r.height = screenHeight
	return r
}

//SortMode sets the sort mode to use when rendering the container list
func (r *DockerNetworksRenderer) SortMode(sortMode docker.SortNetworksMode) {
	r.sortMode = sortMode
}

//Render docker ps
func (r *DockerNetworksRenderer) Render() string {
	if ok, err := r.daemon.Ok(); !ok { // If there was an error connecting to the Docker host...
		return err.Error() // then simply return the error string.
	}
	updateCursorPosition(r.cursor, r.daemon.NetworksCount())

	vars := struct {
		NetworkTable string
	}{
		r.networksTable(),
	}

	buffer := new(bytes.Buffer)
	r.networksTableTemplate.Execute(buffer, vars)

	return buffer.String()
}
func (r *DockerNetworksRenderer) networksTable() string {
	buffer := new(bytes.Buffer)
	t := tabwriter.NewWriter(buffer, 22, 0, 1, ' ', 0)
	replacer := strings.NewReplacer(`\t`, "\t", `\n`, "\n")
	fmt.Fprintln(t, replacer.Replace(r.tableHeader()))
	fmt.Fprint(t, replacer.Replace(r.networkInformation()))
	t.Flush()
	return buffer.String()
}
func (r *DockerNetworksRenderer) tableHeader() string {
	columns := make([]string, len(r.columns))
	for i, col := range r.columns {
		if r.sortMode != col.mode {
			columns[i] = col.title
		} else {
			columns[i] = arrow() + col.title
		}
	}
	return "<green>" + strings.Join(columns, "\t") + "</>"
}

func (r *DockerNetworksRenderer) networkInformation() string {
	buf := bytes.NewBufferString("")
	networks := r.networksToShow()
	selected := len(networks) - 1
	if r.cursor.Line < selected {
		selected = r.cursor.Line
	}
	context := docker.FormattingContext{
		Output:   buf,
		Template: r.networksTemplate,
		Trunc:    true,
		Selected: selected,
	}
	docker.FormatNetworks(
		context,
		networks)

	return buf.String()
}

func (r *DockerNetworksRenderer) networksToShow() []godocker.Network {
	networks, _ := r.daemon.Networks()
	cursorPos := r.cursor.Line
	linesForNetworks := r.height - r.networkTableStart - 1

	if len(networks) < linesForNetworks {
		return networks
	}

	start := 0
	end := len(networks)

	if cursorPos > linesForNetworks {
		start = cursorPos + 1 - linesForNetworks
		end = cursorPos + 1
	} else if cursorPos == linesForNetworks {
		start = 1
		end = linesForNetworks + 1
	} else {
		start = 0
		end = linesForNetworks
	}

	return networks[start:end]
}

func buildNetworkTableTemplate(dockerInfo string) *template.Template {
	markup := dockerInfo +
		`


{{.NetworkTable}}
`
	return template.Must(template.New(`networks`).Parse(markup))
}

func buildNetworksTemplate() *template.Template {

	return template.Must(template.New(`network`).Parse(docker.DefaultNetworkTableFormat))
}
