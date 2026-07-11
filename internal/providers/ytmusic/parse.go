package ytmusic

import (
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"

	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/normalize"
)

var noteChars = "♩♪♫♬𝄞𝄟𝄠"

func isOnlyNotes(s string) bool {
	return strings.TrimFunc(s, func(r rune) bool {
		return strings.ContainsRune(noteChars, r)
	}) == ""
}

// unmarshals a JSON number or JSON string into int64, as YouTube sometimes
// returns millisecond timestamps as quoted strings (?)
type int64S int64

func (v *int64S) UnmarshalJSON(b []byte) error {
	var n int64
	if json.Unmarshal(b, &n) == nil {
		*v = int64S(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*v = int64S(n)
	return nil
}

type searchResp struct {
	Contents struct {
		TabbedSearchResultsRenderer struct {
			Tabs []struct {
				TabRenderer struct {
					Content struct {
						SectionListRenderer struct {
							Contents []struct {
								MusicShelfRenderer *struct {
									Contents []struct {
										MusicResponsiveListItemRenderer struct {
											Overlay struct {
												MusicItemThumbnailOverlayRenderer struct {
													Content struct {
														MusicPlayButtonRenderer struct {
															PlayNavigationEndpoint struct {
																WatchEndpoint struct {
																	VideoID                            string `json:"videoId"`
																	WatchEndpointMusicSupportedConfigs struct {
																		WatchEndpointMusicConfig struct {
																			MusicVideoType string `json:"musicVideoType"`
																		} `json:"watchEndpointMusicConfig"`
																	} `json:"watchEndpointMusicSupportedConfigs"`
																} `json:"watchEndpoint"`
															} `json:"playNavigationEndpoint"`
														} `json:"musicPlayButtonRenderer"`
													} `json:"content"`
												} `json:"musicItemThumbnailOverlayRenderer"`
											} `json:"overlay"`
											FlexColumns []struct {
												MusicResponsiveListItemFlexColumnRenderer struct {
													Text struct {
														Runs []struct {
															Text string `json:"text"`
														} `json:"runs"`
													} `json:"text"`
												} `json:"musicResponsiveListItemFlexColumnRenderer"`
											} `json:"flexColumns"`
										} `json:"musicResponsiveListItemRenderer"`
									} `json:"contents"`
								} `json:"musicShelfRenderer"`
							} `json:"contents"`
						} `json:"sectionListRenderer"`
					} `json:"content"`
				} `json:"tabRenderer"`
			} `json:"tabs"`
		} `json:"tabbedSearchResultsRenderer"`
	} `json:"contents"`
}

func (r *searchResp) videoID(inputArtist, wantTitle string, log *slog.Logger) string {
	for _, tab := range r.Contents.TabbedSearchResultsRenderer.Tabs {
		for _, section := range tab.TabRenderer.Content.SectionListRenderer.Contents {
			if section.MusicShelfRenderer == nil {
				continue
			}
			for _, item := range section.MusicShelfRenderer.Contents {
				renderer := item.MusicResponsiveListItemRenderer
				we := renderer.Overlay.MusicItemThumbnailOverlayRenderer.Content.MusicPlayButtonRenderer.PlayNavigationEndpoint.WatchEndpoint
				if we.WatchEndpointMusicSupportedConfigs.WatchEndpointMusicConfig.MusicVideoType != "MUSIC_VIDEO_TYPE_ATV" || we.VideoID == "" {
					continue
				}
				var ytTitle, ytArtist string
				if len(renderer.FlexColumns) > 0 {
					runs := renderer.FlexColumns[0].MusicResponsiveListItemFlexColumnRenderer.Text.Runs
					if len(runs) > 0 {
						ytTitle = runs[0].Text
					}
				}
				if len(renderer.FlexColumns) > 1 {
					runs := renderer.FlexColumns[1].MusicResponsiveListItemFlexColumnRenderer.Text.Runs
					if len(runs) > 0 {
						ytArtist = runs[0].Text
					}
				}
				titleOK := normalize.Title(ytTitle) == wantTitle
				artistOK := normalize.ArtistMatch(ytArtist, inputArtist)
				log.Debug("candidate",
					"yt_title", ytTitle, "yt_artist", ytArtist,
					"want_title", wantTitle, "want_artist", inputArtist,
					"title_ok", titleOK, "artist_ok", artistOK,
				)
				if titleOK && artistOK {
					return we.VideoID
				}
			}
		}
	}
	return ""
}

type nextResp struct {
	Contents struct {
		SingleColumnMusicWatchNextResultsRenderer struct {
			TabbedRenderer struct {
				WatchNextTabbedResultsRenderer struct {
					Tabs []struct {
						TabRenderer struct {
							Endpoint struct {
								BrowseEndpoint struct {
									BrowseID                              string `json:"browseId"`
									BrowseEndpointContextSupportedConfigs struct {
										BrowseEndpointContextMusicConfig struct {
											PageType string `json:"pageType"`
										} `json:"browseEndpointContextMusicConfig"`
									} `json:"browseEndpointContextSupportedConfigs"`
								} `json:"browseEndpoint"`
							} `json:"endpoint"`
							Unselectable bool `json:"unselectable"`
						} `json:"tabRenderer"`
					} `json:"tabs"`
				} `json:"watchNextTabbedResultsRenderer"`
			} `json:"tabbedRenderer"`
		} `json:"singleColumnMusicWatchNextResultsRenderer"`
	} `json:"contents"`
}

func (r *nextResp) browseID() string {
	tabs := r.Contents.SingleColumnMusicWatchNextResultsRenderer.TabbedRenderer.WatchNextTabbedResultsRenderer.Tabs
	for _, tab := range tabs {
		tr := tab.TabRenderer
		if tr.Unselectable {
			continue
		}
		be := tr.Endpoint.BrowseEndpoint
		if be.BrowseEndpointContextSupportedConfigs.BrowseEndpointContextMusicConfig.PageType == "MUSIC_PAGE_TYPE_TRACK_LYRICS" {
			return be.BrowseID
		}
	}
	return ""
}

type timedBrowseResp struct {
	Contents struct {
		ElementRenderer struct {
			NewElement struct {
				Type struct {
					ComponentType struct {
						Model struct {
							TimedLyricsModel struct {
								LyricsData struct {
									SourceMessage   string `json:"sourceMessage"`
									TimedLyricsData []struct {
										LyricLine string `json:"lyricLine"`
										CueRange  struct {
											StartTimeMilliseconds int64S `json:"startTimeMilliseconds"`
											EndTimeMilliseconds   int64S `json:"endTimeMilliseconds"`
										} `json:"cueRange"`
									} `json:"timedLyricsData"`
								} `json:"lyricsData"`
							} `json:"timedLyricsModel"`
						} `json:"model"`
					} `json:"componentType"`
				} `json:"type"`
			} `json:"newElement"`
		} `json:"elementRenderer"`
	} `json:"contents"`
}

func (r *timedBrowseResp) parse() ([]lyrics.Line, string) {
	data := r.Contents.ElementRenderer.NewElement.Type.ComponentType.Model.TimedLyricsModel.LyricsData
	if len(data.TimedLyricsData) == 0 {
		return nil, ""
	}
	out := make([]lyrics.Line, 0, len(data.TimedLyricsData))
	for _, tl := range data.TimedLyricsData {
		text := strings.TrimSpace(tl.LyricLine)
		if text == "" || isOnlyNotes(text) {
			continue
		}
		out = append(out, lyrics.Line{
			StartMs: int64(tl.CueRange.StartTimeMilliseconds),
			EndMs:   int64(tl.CueRange.EndTimeMilliseconds),
			Text:    tl.LyricLine,
		})
	}
	return out, data.SourceMessage
}
