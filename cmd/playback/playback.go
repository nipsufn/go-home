// TODO: control MPD
package playback

import (
	"net/url"
	"time"

	"go-home/config"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"

	"github.com/fhs/gompd/v2/mpd"
)

var (
	mpdUri   = config.ConfigSingleton.Playback.MpdUrl.String()
	mpdProto = config.ConfigSingleton.Playback.MpdProto
	maxVol   = config.ConfigSingleton.Playback.MaxVolume
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
	client, err := mpd.Dial(mpdProto, mpdUri)
	if err != nil {
		return err
	}
	if client.Clear() != nil {
		return err
	}
	if client.Add(playlistUrl.String()) != nil {
		return err
	}
	if fadeIn != time.Duration(0) {
		if client.SetVolume(0) != nil {
			return err
		}
	}
	if client.Play(-1) != nil {
		return err
	}
	if fadeIn != time.Duration(0) {
		var i int
		delayDuration := time.Duration(fadeIn / time.Duration(maxVol))
		log.Tracef("delaySec: %v", delayDuration)
		for i = 0; i <= int(maxVol); i++ {
			if client.SetVolume(i) != nil {
				return err
			}
			log.Tracef("iterating play fade-in - iteration %v", i)
			time.Sleep(delayDuration)
		}
	}
	return nil
}

func Clear(fadeOut time.Duration) error {
	client, err := mpd.Dial(mpdProto, mpdUri)
	if err != nil {
		return err
	}
	if fadeOut != time.Duration(0) {
		var i int
		delayDuration := time.Duration(fadeOut / time.Duration(maxVol))
		log.Tracef("delaySec: %v", delayDuration)
		for i = 0; i <= int(maxVol); i++ {
			if client.SetVolume(int(maxVol)-i) != nil {
				return err
			}
			log.Tracef("iterating clear fade-out - iteration %v", i)
			time.Sleep(delayDuration)
		}
	}
	if client.Clear() != nil {
		return err
	}
	return nil
}
