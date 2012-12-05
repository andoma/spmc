package main

import "encoding/json"

type IndexedPlugin struct {
	PluginId        string `json:"id"`;
	Version         string `json:"version"`;
	Type            string `json:"type"`;
	Author          string `json:"author"`;
	ShowtimeVersion string `json:"showtimeVersion"`;
	Title           string `json:"title"`;
	Synopsis        string `json:"synopsis"`;
	Description     string `json:"description"`;
	Homepage        string `json:"homepage"`;
	Category        string `json:"category"`;
	DownloadURL     string `json:"downloadURL"`;
	Icon            string `json:"icon"`;
}

type ShowtimeIndex struct {
	Version int `json:"version"`;
	Plugins []IndexedPlugin `json:"plugins"`;
}


func buildShowtimeIndex(reqver *Version, betapasswords []string) ([]byte, error) {
	var si ShowtimeIndex;
	
	si.Version = 1;

	for _, p := range plugins {
		var best *PluginVersion;
		var beta = false;
		for _, pw := range betapasswords {
			if p.BetaSecret == pw {
				beta = true;
				break;
			}
		}

		for _, pv := range p.versions {
			if pv.Status != "a" {
				continue;
			}

			if !pv.Published && !beta {
				continue;
			}
			
			if reqver != nil && pv.showtime_ver.isBiggerThan(reqver) {
				// User's version of Showtime too low
				continue;
			}

			if best == nil || pv.pkg_ver.isBiggerThan(best.pkg_ver) {
				best = pv;
			}
		}
		if best == nil {
			continue;
		}

		ip := IndexedPlugin{
		PluginId: best.PluginId,
		Version: best.Version,
		Type: best.Version,
		Author: best.Author,
		ShowtimeVersion: best.ShowtimeVersion,
		Title: best.Title,
		Synopsis: best.Synopsis,
		Description: best.Description,
		Homepage: best.Homepage,
		Category: best.Category,
		DownloadURL: "data/" + best.PkgDigest,
		Icon: "data/" + best.IconDigest,
		};

		si.Plugins = append(si.Plugins, ip);
	}

	return json.Marshal(si);
}
