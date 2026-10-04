// TODO: control MPD
package playback

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"go-home/config"

	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"

	"github.com/fhs/gompd/v2/mpd"
)

// needs work on init to be accessible at this point
// var (
// 	mpdUri   = config.ConfigSingleton.Playback.MpdUrl
// 	mpdProto = config.ConfigSingleton.Playback.MpdProto
// 	maxVol   = config.ConfigSingleton.Playback.MaxVolume
// )

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
		Args:  cobra.ExactArgs(1),
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
	mpdUri := config.ConfigSingleton.Playback.MpdUrl
	mpdProto := config.ConfigSingleton.Playback.MpdProto
	maxVol := config.ConfigSingleton.Playback.MaxVolume
	client, err := mpd.Dial(mpdProto, mpdUri)
	if err != nil {
		return errors.Join(fmt.Errorf(`cannot connect to mpd at %s %s`, mpdProto, mpdUri), err)
	}
	defer client.Close()
	err = client.Clear()
	if err != nil {
		return errors.Join(errors.New(`cannot clear`), err)
	}
	err = client.Add(playlistUrl.String())
	if err != nil {
		return errors.Join(errors.New(`cannot add to playlist`), err)
	}
	if fadeIn != time.Duration(0) {
		err = client.SetVolume(0)
		if err != nil {
			return err
		}
	}
	err = client.Play(-1)
	if err != nil {
		return errors.Join(errors.New(`cannot play`), err)
	}
	if fadeIn != time.Duration(0) {
		var i int
		delayDuration := time.Duration(fadeIn / time.Duration(maxVol))
		log.Tracef("delaySec: %v", delayDuration)
		for i = 0; i <= int(maxVol); i++ {
			err = client.SetVolume(i)
			if err != nil {
				return errors.Join(errors.New(`cannot set volume`), err)
			}
			log.Tracef("iterating play fade-in - iteration %v", i)
			time.Sleep(delayDuration)
		}
	}
	return nil
}

func Clear(fadeOut time.Duration) error {
	mpdUri := config.ConfigSingleton.Playback.MpdUrl
	mpdProto := config.ConfigSingleton.Playback.MpdProto
	maxVol := config.ConfigSingleton.Playback.MaxVolume
	client, err := mpd.Dial(mpdProto, mpdUri)
	if err != nil {
		return errors.Join(fmt.Errorf(`cannot connect to mpd at %s %s`, mpdProto, mpdUri), err)
	}
	defer client.Close()
	if fadeOut != time.Duration(0) {
		var i int
		delayDuration := time.Duration(fadeOut / time.Duration(maxVol))
		log.Tracef("delaySec: %v", delayDuration)
		for i = 0; i <= int(maxVol); i++ {
			err = client.SetVolume(int(maxVol) - i)
			if err != nil {
				return errors.Join(errors.New(`cannot set volume`), err)
			}
			log.Tracef("iterating clear fade-out - iteration %v", i)
			time.Sleep(delayDuration)
		}
	}
	err = client.Clear()
	if err != nil {
		return errors.Join(errors.New(`cannot clear`), err)
	}
	err = client.SetVolume(int(maxVol))
	if err != nil {
		return err
	}
	return nil
}
