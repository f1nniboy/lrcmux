package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/f1nniboy/lrcmux/internal/cache"
	"github.com/f1nniboy/lrcmux/internal/config"
	"github.com/f1nniboy/lrcmux/internal/format"
	"github.com/f1nniboy/lrcmux/internal/isrc"
	"github.com/f1nniboy/lrcmux/internal/logging"
	"github.com/f1nniboy/lrcmux/internal/lyrics"
	"github.com/f1nniboy/lrcmux/internal/orchestrator"
	"github.com/f1nniboy/lrcmux/internal/providers/registry"
	"github.com/f1nniboy/lrcmux/internal/proxy"
)

// one /get against a fresh memory cache, so it shows the fanout but not what the server has cached
func main() {
	cfgPath := flag.String("config", "config.toml", "path to config file")
	sources := flag.String("sources", "", "comma-separated providers to restrict the fanout to, prefix with '!' to exclude")
	level := flag.String("level", "word", "highest sync level to accept, or the exact one with -strict")
	formatName := flag.String("format", "lrc", "output format")
	var req orchestrator.Request
	flag.StringVar(&req.Artist, "artist", "", "artist")
	flag.StringVar(&req.Title, "title", "", "title")
	flag.StringVar(&req.Album, "album", "", "album")
	flag.StringVar(&req.ISRC, "isrc", "", "isrc, instead of artist and title")
	flag.Int64Var(&req.Duration, "duration", 0, "duration in seconds")
	flag.BoolVar(&req.Strict, "strict", false, "fail instead of falling back to a lower sync level")
	flag.Parse()
	if *sources != "" {
		req.Sources = strings.Split(*sources, ",")
	}

	if err := run(*cfgPath, *level, *formatName, req); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(cfgPath, levelName, formatName string, req orchestrator.Request) error {
	if req.ISRC == "" && (req.Artist == "" || req.Title == "") {
		return errors.New("usage: probe [-sources a,b] (-isrc isrc | -artist artist -title title)")
	}
	enc, err := format.Get(formatName)
	if err != nil {
		return err
	}
	if req.Level, err = lyrics.ParseLevel(levelName); err != nil {
		return err
	}
	lo, hi := enc.Levels()
	req.Level = min(max(req.Level, lo), hi)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	logging.Init(cfg.Log)

	pools, err := proxy.LoadAll(cfg.Proxies)
	if err != nil {
		return err
	}
	c := cache.NewMemory()
	provs, err := registry.Build(cfg, c, pools, logging.New("probe"))
	if err != nil {
		return err
	}
	resolver := isrc.New(&http.Client{Timeout: isrc.Timeout}, c, cfg.Cache.TTL.Miss.Duration, logging.New("isrc"))
	orch := orchestrator.New(provs, c, orchestrator.NewBreaker(c, logging.New("breaker")), resolver, nil, orchestrator.Options{
		Timeout: cfg.Provider.Timeout.Duration,
		TTL:     cfg.Cache.TTL,
	})

	start := time.Now()
	resp, err := orch.Get(context.Background(), req)
	took := time.Since(start).Round(time.Millisecond)
	if err != nil {
		return fmt.Errorf("after %s: %w", took, err)
	}
	r := resp.Result
	fmt.Fprintf(os.Stderr, "track: %s - %s (%s, deezer %d, %d s)\n", r.Track.Artist, r.Track.Title, r.Track.ISRC, r.Track.DeezerID, r.Track.Duration)
	if r.Instrumental {
		fmt.Fprintf(os.Stderr, "%s: instrumental, %s\n", r.Source.ID, took)
		return nil
	}
	fmt.Fprintf(os.Stderr, "%s: %s, %d lines, %s\n", r.Source.ID, r.SyncLevel, len(r.Lines), took)
	if r.SyncLevel < lo {
		return fmt.Errorf("format %q requires %s-synced lyrics, try -format txt", formatName, lo)
	}
	return enc.Encode(os.Stdout, r)
}
