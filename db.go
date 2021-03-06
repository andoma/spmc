package main
import "errors"
import "fmt"
import "io"
import "crypto/sha1"
import "crypto/rand"

import (
    "os"
    "log"
    "github.com/ziutek/mymysql/autorc"
    _ "github.com/ziutek/mymysql/thrsafe"
)

const (
    db_proto = "tcp"
    db_addr  = "127.0.0.1:3306"
    db_user  = "plugcentral"
    db_pass  = "plugcentral"
    db_name  = "showtime_plugins"
)

var db *autorc.Conn;

var (
	plugin_insert_stmt, version_insert_stmt, version_delete_stmt,
		status_set_stmt, published_set_stmt, user_query_stmt, admin_query_stmt,
		user_insert_stmt, version_dlinc_stmt, plugin_update_stmt,
		track_update_stmt,
		plugin_delete_stmt *autorc.Stmt;
)

func mysqlError(err error) (ret bool) {
	ret = (err != nil);
	if ret {
		log.Println("MySQL error:", err);
	}
	return;
}

func mysqlErrExit(err error) {
	if mysqlError(err) {
		os.Exit(1);
	}
}

func newPlugin(id, owner, betasecret string, downloadurl string) *Plugin {
	p := Plugin{id, owner, "", betasecret, downloadurl, 
		make(map[string]*PluginVersion)};
	return &p;
}


func init() {
	var err error;

	loadConf();

	if len(config.Db.Addr) < 1 ||
		len(config.Db.User) < 1 ||
		len(config.Db.Pass) < 1 ||
		len(config.Db.Name) < 1 {
		fmt.Printf("Missing db config\n");
		os.Exit(1);
	}

	db = autorc.New("tcp", "",
		config.Db.Addr,
		config.Db.User,
		config.Db.Pass,
		config.Db.Name);

	db.Raw.Register("SET NAMES utf8");

	plugin_insert_stmt, err = db.Prepare("INSERT INTO plugin (id, owner) VALUES(?, ?)");
	mysqlErrExit(err);

	version_insert_stmt, err = db.Prepare("INSERT INTO version (plugin_id, version, type, author, showtime_min_version, title, synopsis, description, homepage, pkg_digest, category, icon_digest, status) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)");
	mysqlErrExit(err);

	version_delete_stmt, err = db.Prepare("DELETE FROM version WHERE plugin_id=? AND version=?");
	mysqlErrExit(err);

	status_set_stmt, err = db.Prepare("UPDATE version SET status=? WHERE plugin_id=? AND version=?");
	mysqlErrExit(err);

	published_set_stmt, err = db.Prepare("UPDATE version SET published=? WHERE plugin_id=? AND version=?");
	mysqlErrExit(err);

	user_query_stmt, err = db.Prepare("SELECT salt, sha1, email, admin, autoapprove FROM users WHERE username = ?");
	mysqlErrExit(err);

	admin_query_stmt, err = db.Prepare("SELECT username, email, autoapprove FROM users WHERE admin = true");
	mysqlErrExit(err);

	user_insert_stmt, err = db.Prepare("INSERT INTO users (username, salt, sha1, email) VALUES(?, ?, ?, ?)");
	mysqlErrExit(err);

	version_dlinc_stmt, err = db.Prepare("UPDATE version SET downloads=downloads+1 WHERE plugin_id = ? AND version = ?");
	mysqlErrExit(err);

	plugin_update_stmt, err = db.Prepare("UPDATE plugin SET betasecret=?, downloadurl=? WHERE id=?");
	mysqlErrExit(err);

	track_update_stmt, err = db.Prepare("INSERT INTO tracking (id, ua, count, ipaddr, cc) VALUE (?, ?, 1, ?, ?) ON DUPLICATE KEY UPDATE count=count + 1, ua=?, updated=now(), ipaddr=?, cc=?");
	mysqlErrExit(err);

	plugin_delete_stmt, err = db.Prepare("DELETE FROM plugin WHERE id=?");
	mysqlErrExit(err);


	rows, _, err := db.Query("SELECT id, owner, betasecret,downloadurl FROM plugin");
	mysqlErrExit(err);

	for _, r := range rows {
		id    := r.Str(0);
		plugins[id] = newPlugin(id, r.Str(1), r.Str(2), r.Str(3));
	}


	rows, _, err = db.Query("SELECT plugin_id, version, type, author, downloads," + 
                                "published, showtime_min_version, title, synopsis, description," +
                                "homepage, pkg_digest, comment, category, status, icon_digest FROM version");
	mysqlErrExit(err);

	for _, r := range rows {
		plugin_id := r.Str(0);
		version   := r.Str(1);
		v := PluginVersion{
		PluginId: plugin_id, 
		Version: version,
		Type: r.Str(2),
		Author: r.Str(3),
		Downloads: r.Int(4),
		Published: r.Bool(5),
		ShowtimeVersion: r.Str(6),
		Title: r.Str(7),
		Synopsis: r.Str(8),
		Description: r.Str(9),
		Homepage: r.Str(10),
		PkgDigest: r.Str(11),
		Comment: r.Str(12),
		Category: r.Str(13),
		Status: r.Str(14),
		IconDigest: r.Str(15),
		};
		plugins[plugin_id].versions[version] = &v;
		pkgHashToVersion[v.PkgDigest] = &v;

		v.pkg_ver, err = parseVersionString(v.Version);
		if err != nil {
			log.Println(err);
			os.Exit(1);
		}

		v.showtime_ver, err = parseVersionString(v.ShowtimeVersion);
		if err != nil {
			log.Println(err);
			os.Exit(1);
		}


	}
}


/*
func getPlugins() *Plugins {
	rows, _, err := plugin_list_stmt.Exec();
	if mysqlError(err) {
		return nil;
	}
	
	ret := []Plugin{};
	for _, r := range rows {
		ret = append(ret, Plugin{Id: r.Str(0), Owner: r.Str(1)});
	}

	return &Plugins{ret};
}
*/


func getOrCreatePlugin(id string, u *User) (*Plugin, error) {
	p := plugins[id];
	if p == nil {
		_, _, err := plugin_insert_stmt.Exec(id, u.Username);
		if err != nil {
			return nil, err;
		}
		
		p = newPlugin(id, u.Username, "", "");
		plugins[id] = p;
	} else {
		if p.Owner != u.Username && u.Admin == false {
			return nil, errors.New(fmt.Sprintf("Not owner of plugin %s", id));
		}

	}
	return p, nil;
}

func getPluginVersion(id string, version string) (pv *PluginVersion) {
	p := plugins[id];
	if p != nil {
		pv = p.versions[version];
	}
	return;
}

func ingestVersion(pv *PluginVersion, u *User) (error) {
	p, err := getOrCreatePlugin(pv.PluginId, u);
	if err != nil {
		return err;
	}

	if p.versions[pv.Version] != nil {
		return errors.New(fmt.Sprintf("Version %s already exist for plugin %s",
			pv.Version, pv.PluginId));
	}

	_, _, err = version_insert_stmt.Exec(
		pv.PluginId,
		pv.Version,
		pv.Type,
		pv.Author,
		pv.ShowtimeVersion,
		pv.Title,
		pv.Synopsis,
		pv.Description,
		pv.Homepage,
		pv.PkgDigest,
		pv.Category,
		pv.IconDigest,
		pv.Status);

	if err != nil {
		log.Println("MySQL error:", err);
		return err;
	}

	p.versions[pv.Version] = pv;
	pkgHashToVersion[pv.PkgDigest] = pv;

	notifyUser(p.Owner,
		fmt.Sprintf("Ingested: %s (%s)", pv.PluginId, pv.Version),
		fmt.Sprintf("Version %s of %s was ingested by %s\n%s",
		pv.Version, pv.PluginId, u.Username, pv.liveStatus()));

	if pv.Status != "a" {
		notifyAdmin(
			fmt.Sprintf("Approval needed: %s (%s)", pv.PluginId, pv.Version),
			fmt.Sprintf("Version %s of %s was ingested by %s",
			pv.Version, pv.PluginId, u.Username));
	}

	return nil;
}

func deleteVersion(plugin, version string) {
	p := plugins[plugin];
	if p != nil {
		pv := p.versions[version];
		if pv.Status == "p" {
			version_delete_stmt.Exec(plugin, version);
			delete(pkgHashToVersion, pv.PkgDigest);
			delete(p.versions, version);
			log.Printf("Deleted version %s from plugin %s", version, plugin);
		}
	}
}



func setStatus(u *User, plugin, version, newStatus, reason string) {
	if !u.Admin {
		return;
	}
	p := plugins[plugin];
	if p == nil {
		return;
	}
	status_set_stmt.Exec(newStatus, plugin, version);
	pv := p.versions[version];
	pv.Status = newStatus;

	switch newStatus {

	case "a":
		log.Printf("Plugin %s version %s approved by %s",
			plugin, version, u.Username);
		notifyUser(p.Owner,
			fmt.Sprintf("Approved: %s (%s)", plugin, version),
			fmt.Sprintf("Version %s of %s was approved by %s.\n%s",
			version, plugin, u.Username, pv.liveStatus()));
	case "r":
		log.Printf("Plugin %s version %s rejected by %s",
			plugin, version, u.Username);
		notifyUser(p.Owner,
			fmt.Sprintf("Rejected: %s (%s)", plugin, version),
			fmt.Sprintf("Version %s of %s was rejected by %s.\n%s",
			version, plugin, u.Username, reason));
	case "p":
		log.Printf("Plugin %s version %s back to pending by %s",
			plugin, version, u.Username);
			
	}
}


func setPublished(u *User, plugin, version string, set bool) {
	p := plugins[plugin];
	if p != nil && (u.Admin || p.Owner == u.Username) {
		published_set_stmt.Exec(set, plugin, version);
		pv := p.versions[version];
		pv.Published = set;
		log.Printf("Plugin %s version %s published set to %t by %s",
			plugin, version, set, u.Username);
		notifyUser(p.Owner,
			fmt.Sprintf("Published: %s (%s)", plugin, version),
			fmt.Sprintf("Version %s of %s was published by %s.\n%s",
			version, plugin, u.Username, pv.liveStatus()));

	}
}


func dbIncDownloads(digest string) {
	pv := pkgHashToVersion[digest];
	if pv != nil {
		pv.Downloads++;
		version_dlinc_stmt.Exec(pv.PluginId, pv.Version);
	}
}

func dbAuthUser(username, password string) (*User, error) {
	rows, _, err := user_query_stmt.Exec(username);
	if err != nil {
		return nil, err;
	}
	
	if len(rows) == 0 {
		return nil, errors.New("Invalid username or password");
	}

	row := rows[0];
	h := sha1.New();

	io.WriteString(h, row.Str(0));
	io.WriteString(h, password);
	hexdigest := fmt.Sprintf("%x", h.Sum(nil));
	if hexdigest != row.Str(1) {
		return nil, errors.New("Invalid username or password");
	}

	u := User{username, row.Str(2), row.Bool(3), row.Bool(4)};

	return &u, nil;
}



func dbAddUser(username, password, email string) (*User, error) {

	b := make([]byte, 16);
	n, err := io.ReadFull(rand.Reader, b);
	if err != nil {
		return nil, err;
	}

	if n != 16 {
		return nil, errors.New("Not enough salt in store");
	}
	salt := fmt.Sprintf("%x", b);
	h := sha1.New();
	io.WriteString(h, salt);
	io.WriteString(h, password);
	digest := fmt.Sprintf("%x", h.Sum(nil));

	_, _, err = user_insert_stmt.Exec(username, salt, digest, email);
	if err != nil {
		return nil, err;
	}

	u := User{username, email, false, false};
	return &u, nil;
}


func dbGetUser(username string) (*User, error) {
	rows, _, err := user_query_stmt.Exec(username);
	if err != nil {
		return nil, err;
	}
	
	if len(rows) == 0 {
		return nil, errors.New("Invalid username or password");
	}

	row := rows[0];
	u := User{username, row.Str(2), row.Bool(3), row.Bool(4)};
	return &u, nil;
}


func dbGetAdmins() ([]*User, error) {
	rows, _, err := admin_query_stmt.Exec();
	if err != nil {
		return nil, err;
	}

	users := make([]*User,0,0);
	
	for _, row := range rows {
		u := User{row.Str(0), row.Str(1), true, row.Bool(2)};
		users = append(users, &u);
	}
	return users, nil;
}


func updatePlugin(u *User, id, betasecret string, downloadurl string) {
	p := plugins[id];
	if p == nil {
		return;
	}
	if p.Owner != u.Username && u.Admin == false {
		return;
	}

	p.BetaSecret = betasecret;
	p.DownloadURL = downloadurl;
	plugin_update_stmt.Exec(betasecret, downloadurl, id);
}


func updateTracking(id string, ua string, ipaddr string, cc string) {
	track_update_stmt.Exec(id, ua, ipaddr, cc, ua, ipaddr, cc);
}


func erasePlugin(id string) (error) {
	fmt.Printf("id = %s\n", id);
	p := plugins[id];
	if p == nil {
		return errors.New(fmt.Sprintf("Plugin '%s' not found", id));
	}

	for _, v := range p.versions {
		version_delete_stmt.Exec(id, v);
		delete(pkgHashToVersion, v.PkgDigest);
	}
	plugin_delete_stmt.Exec(id);
	delete(plugins, id);
	return nil;
}