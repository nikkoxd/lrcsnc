package lrclib

import (
	errs "lrcsnc/internal/lyrics/errors"
	"lrcsnc/internal/pkg/log"
	playerStructs "lrcsnc/internal/pkg/structs/player"
	"lrcsnc/internal/pkg/types"
	"strings"
)

func (l Provider) Get(song playerStructs.Song) (playerStructs.LyricsData, error) {
	var body []byte
	var err error
	var res playerStructs.LyricsData = playerStructs.LyricsData{LyricsState: types.LyricsStateNotFound}

	log.Debug("lyrics/providers/lrclib/get", "Trying to get lyrics directly...")
	body, err = getLyrics(song.Title, strings.Join(song.Artists, ", "), song.Album, song.Duration)
	if err == nil {
		outs, err := parseResps(body)
		if err == nil && outs[0].toLyricsData().LyricsState != types.LyricsStatePlain {
			return outs[0].toLyricsData(), nil
		}
	}
	log.Debug("lyrics/providers/lrclib/get", "Trying to search around for lyrics more...")
	body, err = searchLyrics(song.Title, strings.Join(song.AlbumArtists, ", "))
	if err == nil {
		res, err = responseListToLyricsData(&song, body)
	}
	if err != errs.NotFound {
		return res, err
	}

	log.Debug("lyrics/providers/lrclib/get", "Failed; the lyrics for this song don't exist")

	// If nothing is found, return a not found state
	return res, errs.NotFound
}
