package app

import (
	"fmt"
	"io"
	"time"

	"github.com/moncho/dry/appui"
	"github.com/moncho/dry/ui"
)

type viewMode uint16

//known view modes
const (
	Main viewMode = iota
	Images
	Networks
	HelpMode
	StatsMode
	ImageHistoryMode
	InfoMode
	InspectImageMode
	InspectNetworkMode
	InspectMode
)

const (
	//The position from the top (0) where a line describing what is
	//being shown is placed. Kind of a magic number.
	screenDescriptionIndex = 5
)

//Render renders dry in the given screen
func Render(d *Dry, screen *ui.Screen, status *ui.StatusBar) {
	switch d.state.viewMode {
	case Main:
		{
			//after a refresh, sorting is needed
			d.dockerDaemon.Sort(d.state.SortMode)
			d.renderer.SortMode(d.state.SortMode)
			status.Render()
			screen.RenderLine(0, 0, `<right><white>`+time.Now().Format(`15:04:05`)+`</></right>`)
			screen.Render(1, d.renderer.Render())
			screen.RenderLine(0, screenDescriptionIndex,
				fmt.Sprintf(
					"<b><blue>Containers: </><yellow>%d</></>", d.dockerDaemon.ContainersCount()))

			screen.RenderLineWithBackGround(0, screen.Height-1, keyMappings, ui.MenuBarBackgroundColor)
			d.state.changed = false
		}
	case Images:
		{
			status.Render()
			screen.RenderLine(0, 0, `<right><white>`+time.Now().Format(`15:04:05`)+`</></right>`)
			d.dockerDaemon.SortImages(d.state.SortImagesMode)

			screen.Render(1,
				appui.NewDockerImagesRenderer(d.dockerDaemon, screen.Height, screen.Cursor, d.state.SortImagesMode).Render())
			screen.RenderLine(0, screenDescriptionIndex,
				fmt.Sprintf(
					"<b><blue>Images: </><yellow>%d</></>", d.dockerDaemon.ImagesCount()))
			screen.RenderLineWithBackGround(0, screen.Height-1, imagesKeyMappings, ui.MenuBarBackgroundColor)
			d.state.changed = false
		}
	case Networks:
		{
			status.Render()
			screen.RenderLine(0, 0, `<right><white>`+time.Now().Format(`15:04:05`)+`</></right>`)

			screen.Render(1,
				appui.NewDockerNetworksRenderer(d.dockerDaemon, screen.Height, screen.Cursor, d.state.SortNetworksMode).Render())
			screen.RenderLine(0, screenDescriptionIndex,
				fmt.Sprintf(
					"<b><blue>Networks: </><yellow>%d</></>", d.dockerDaemon.NetworksCount()))
			screen.RenderLineWithBackGround(0, screen.Height-1, networkKeyMappings, ui.MenuBarBackgroundColor)
			d.state.changed = false
		}

	}

	screen.Flush()
}

//Write sends dry output to the given writer
func Write(d *Dry, w io.Writer) {
	switch d.viewMode() {
	case StatsMode:
		{
			if d.stats != nil {
				io.WriteString(w, appui.NewDockerStatsRenderer(d.stats).Render())
			} else {
				io.WriteString(w, "Could not read stats")
			}
		}
	case ImageHistoryMode:
		io.WriteString(w, appui.NewDockerImageHistoryRenderer(d.imageHistory).Render())
	case InspectMode:
		io.WriteString(w, appui.NewDockerInspectRenderer(d.inspectedContainer).Render())
	case InspectImageMode:
		io.WriteString(w, appui.NewDockerInspectImageRenderer(d.inspectedImage).Render())
	case InspectNetworkMode:
		io.WriteString(w, appui.NewDockerInspectNetworkRenderer(d.inspectedNetwork).Render())
	case HelpMode:
		io.WriteString(w, help)
	case InfoMode:
		io.WriteString(w, appui.NewDockerInfoRenderer(d.info).Render())
	default:
		{
			io.WriteString(w, "Dry is not ready yet for rendering, be patient...")
		}
	}
}
