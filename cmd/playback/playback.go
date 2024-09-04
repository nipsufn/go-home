// TODO: control MPD
package playback

import (
	"net/url"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"

	"github.com/fhs/gompd/v2/mpd"
)

func NewPlaybackCmd() (serveCmd *cobra.Command) {
	playbackCmd := &cobra.Command{
		Use:   "playback",
		Short: "Control MPD",
	}
	playbackCmd.AddCommand(newPlayUrlCommand())
	playbackCmd.AddCommand(newClearCommand())
	return playbackCmd
}

func newPlayUrlCommand() (playUrlCmd *cobra.Command) {
	playUrlCmd = &cobra.Command{
		Use:   "start",
		Short: "Start playing URI",
		RunE: func(cmd *cobra.Command, args []string) error {
			firstArg, _ := url.Parse(args[0])
			return PlayURL(url.URL(*firstArg))
		},
	}
	return playUrlCmd
}

func newClearCommand() (clearCmd *cobra.Command) {
	clearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear playlist",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Clear()
		},
	}
	return clearCmd
}

func PlayURL(playlistUrl url.URL) error {
	client, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		log.Warnf("playback.go: func PlayURL: unable to dial MPD daemon")
		return nil
	}
	if client.Clear() != nil {
		log.Warnf("playback.go: func PlayURL: unable to clear MPD playlist")
		return nil
	}
	if client.Add(playlistUrl.String()) != nil {
		log.Warnf("playback.go: func PlayURL: unable to add %v to playlist", playlistUrl)
		return nil
	}
	if client.Play(-1) != nil {
		log.Warnf("playback.go: func PlayURL: unable to play MPD playlist")
		return nil
	}
	return nil
}

func Clear() error {
	client, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		log.Warnf("playback.go: func PlayURL: unable to dial MPD daemon")
		return nil
	}
	if client.Clear() != nil {
		log.Warnf("playback.go: func PlayURL: unable to clear MPD playlist")
		return nil
	}
	return nil
}
