package main

import "encoding/json"
import "strings"
import "strconv"
import "errors"
import "fmt"

var plugins map[string]*Plugin = make(map[string]*Plugin);




type Plugin struct {
	Id          string `json:"id"`;
	Owner       string `json:"owner"`;
	Upstream   *string `json:"upstream,omitempty"`;
	BetaSecret *string `json:"betasecret,omitempty"`;

	versions map[string]*PluginVersion;
}

type Plugins struct {
	Plugins []*Plugin `json:"plugins"`;
}

func (p Plugins) marshal() []byte {
	b, _ := json.Marshal(p);
	return b;
}

func getPlugins(u *User) *Plugins {
	var p Plugins;
	for _, v := range plugins {
		if u.Admin || u.Username == v.Owner {
			p.Plugins = append(p.Plugins, v);
		}
	}
	return &p;
}

type Version struct {
	v [4]int;
}


func parseVersionString(vstr string) (*Version, error) {
	ver := new(Version);
	vec := strings.Split(vstr, ".");
	var err error;

	if len(vec) < 1 || len(vec) > 4 {
		return nil, errors.New(fmt.Sprintf("Malformed version '%s'",
			vstr));
	}
	for i := 0; i < len(vec); i++ {
		ver.v[i], err = strconv.Atoi(vec[i]);
		if err != nil {
			return nil, errors.New(
				fmt.Sprintf("Malformed version '%s' -- %s",
				vstr, err));
		}
	}
	return ver, nil;
}


func (v *Version) isBiggerThan(l *Version) bool {
	for i := 0; i < 4; i++ {
		if v.v[i] > l.v[i] {
			return true;
		}
		if v.v[i] < l.v[i] {
			return false;
		}
	}
	return false;
}

func (v *Version) isBiggerOrEqThan(l *Version) bool {
	for i := 0; i < 4; i++ {
		if v.v[i] > l.v[i] {
			return true;
		}
		if v.v[i] < l.v[i] {
			return false;
		}
	}
	return true;
}


type PluginVersion struct {
	Approved        bool   `json:"approved"`;
	Published       bool   `json:"published"`;
	Comment         string `json:"comment"`;
	PluginId        string `json:"id"`;
	Version         string `json:"version"`;
	Type            string `json:"type"`;
	Author          string `json:"author"`;
	Downloads       int    `json:"downloads"`;
	ShowtimeVersion string `json:"showtimeVersion"`;
	Title           string `json:"title"`;
	Synopsis        string `json:"synopsis"`;
	Description     string `json:"description"`;
	Homepage        string `json:"homepage"`;
	PkgDigest       string `json:"pkgdigest"`;
	Category        string `json:"category"`;
	Icon            string `json:"icon"`;
	IconDigest      string `json:"icondigest"`;
}

type PluginVersions struct {
	PluginVersions []*PluginVersion `json:"versions"`;
}

func (p PluginVersions) marshal() []byte {
	b, _ := json.Marshal(p);
	return b;
}

func getVersions(id string) *PluginVersions {
	p := plugins[id];
	if p == nil {
		return nil;
	}
	
	var pv PluginVersions;

	for _, v := range p.versions {
		pv.PluginVersions = append(pv.PluginVersions, v);
	}
	return &pv;
}


type User struct {
	Username string;
	Email string;
	Admin bool;
	Autoapprove bool;
}
