package main

import "encoding/json"

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
	SHA1            string `json:"sha1"`;
	Category        string `json:"category"`;
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
