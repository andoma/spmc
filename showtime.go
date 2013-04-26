package main

import "encoding/json"
import "sort"

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

type BlackListVersion struct {
	PluginId        string `json:"id"`;
	Version         string `json:"version"`;
}

type ShowtimeIndex struct {
	Version int `json:"version"`;
	Plugins []IndexedPlugin `json:"plugins"`;
	BlackList []BlackListVersion `json:"blacklist"`;
}


func buildShowtimeIndex(reqver *Version, betapasswords []string) ([]byte, error) {
	var si ShowtimeIndex;
	
	si.Version = 1;

	s_plugins := make([]string, len(plugins));
	i := 0;
	for k, _ := range plugins {
		s_plugins[i] = k;
		i++;
	}
	sort.Strings(s_plugins);

	for _, pid := range s_plugins {

		p := plugins[pid];

		var best *PluginVersion;
		var beta = false;
		var allAccess = false;
		for _, pw := range betapasswords {
			if len(config.AllAccess) > 1 && config.AllAccess == pw {
				allAccess = true;
			}
			if p.BetaSecret == pw {
				beta = true;
			}
		}

		s_versions := make([]string, len(p.versions));
		i := 0;
		for k, _ := range p.versions {
			s_versions[i] = k;
			i++;
		}
		sort.Strings(s_versions);

		for _, pvid := range s_versions {
			pv := p.versions[pvid];

			if pv.Status == "r" {
				// Rejected

				blv := BlackListVersion{
				PluginId: pv.PluginId,
				Version: pv.Version,
				};
				si.BlackList = append(si.BlackList, blv);
				continue;
			}

			if !allAccess {

				// If it's not approved, limit to 50 downloads for beta mode
				if pv.Status != "a" && (!beta || pv.Downloads >= 50) {
					continue;
				}

				if !pv.Published && !beta {
					continue;
				}
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
