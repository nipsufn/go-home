// TODO: control MPD
package playback

import (
	"net/url"
	"time"

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
	var fadeIn int
	playUrlCmd = &cobra.Command{
		Use:   "start",
		Short: "Start playing URI",
		RunE: func(cmd *cobra.Command, args []string) error {
			firstArg, _ := url.Parse(args[0])
			return PlayURL(url.URL(*firstArg), time.Duration(fadeIn*int(time.Second)))
		},
	}
	playUrlCmd.Flags().IntVarP(&fadeIn, "fadeInSec", "f", 0, "Fade-in duration")
	return playUrlCmd
}

func newClearCommand() (clearCmd *cobra.Command) {
	var fadeOut int
	clearCmd = &cobra.Command{
		Use:   "clear",
		Short: "Clear playlist",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Clear(time.Duration(fadeOut * int(time.Second)))
		},
	}
	clearCmd.Flags().IntVarP(&fadeOut, "fadeOutSec", "f", 0, "Fade-out duration")
	return clearCmd
}

func PlayURL(playlistUrl url.URL, fadeIn time.Duration) error {
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
	if fadeIn != time.Duration(0) {
		if client.SetVolume(0) != nil {
			log.Warnf("playback.go: func PlayURL: unable to set MPD volume")
			return nil
		}
	}
	if client.Play(-1) != nil {
		log.Warnf("playback.go: func PlayURL: unable to play MPD playlist")
		return nil
	}
	if fadeIn != time.Duration(0) {
		var i int
		delayDuration := time.Duration(fadeIn / 100.0)
		log.Tracef("delaySec: %v", delayDuration)
		for i = 0; i <= 100; i++ {
			if client.SetVolume(i) != nil {
				log.Warnf("playback.go: func PlayURL: unable to set MPD volume")
				return nil
			}
			log.Tracef("iterating play fade-in - iteration %v", i)
			time.Sleep(delayDuration)
		}
	}
	return nil
}

func Clear(fadeOut time.Duration) error {
	client, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		log.Warnf("playback.go: func PlayURL: unable to dial MPD daemon")
		return nil
	}
	if fadeOut != time.Duration(0) {
		var i int
		delayDuration := time.Duration(fadeOut / 100.0)
		log.Tracef("delaySec: %v", delayDuration)
		for i = 0; i <= 100; i++ {
			if client.SetVolume(100-i) != nil {
				log.Warnf("playback.go: func PlayURL: unable to set MPD volume")
				return nil
			}
			log.Tracef("iterating clear fade-out - iteration %v", i)
			time.Sleep(delayDuration)
		}
	}
	if client.Clear() != nil {
		log.Warnf("playback.go: func PlayURL: unable to clear MPD playlist")
		return nil
	}
	return nil
}
