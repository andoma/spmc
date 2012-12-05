package main

import "io"
import "os"
import "net/http"
import "log"
import "strings"
import "encoding/json"
import "regexp"
import "strconv"
import "fmt"
import "crypto/sha1"

var ua_re *regexp.Regexp;

func init() {
	ua_re = regexp.MustCompile("^Showtime [^ ]+ ([0-9]+)\\.([0-9]+)\\.([0-9]+)"); 
}

func httplog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL);
		handler.ServeHTTP(w, r);
	})
}


func getUser(r *http.Request) *User {
	cookie, err := r.Cookie("auth");
	if(err != nil) {
		return nil;
	}
	return validateCookie(cookie.Value);
}

func roothandler(w http.ResponseWriter, r *http.Request) {
	u := getUser(r);
	
	if r.Method == "GET" {
		var filename string;
		if u != nil {
			filename = "static/index.html";
		} else {
			filename = "static/login.html";
		}
		
		file, err := os.Open(filename);
		if err != nil {
			http.NotFound(w, r);
		} else {
			io.Copy(w, file);
			file.Close();
		}
		return;
	}

	if r.Method == "POST" {
		r.ParseForm();
		var u *User;
		var err error;
		if len(r.Form["login"]) > 0 {
			username := r.Form["username"][0];
			password := r.Form["password"][0];
			u, err = dbAuthUser(username, password);
		} else if len(r.Form["register"]) > 0 {
			username := r.Form["username"][0];
			password := r.Form["password"][0];
			email    := r.Form["email"][0];
			u, err = dbAddUser(username, password, email);
		} else {
			http.NotFound(w, r);
			return;
		}


		if err != nil {
			w.WriteHeader(401);
			w.Write([]byte(err.Error()));
			return;
		}


		value, err := authenticateUser(u);
		if err != nil {
			w.WriteHeader(401);
			w.Write([]byte(err.Error()));
			return;
		}

		c := new(http.Cookie);
		c.Name = "auth";
		c.Value = *value;
		http.SetCookie(w, c);
		http.Redirect(w, r, "spmc/", 301);
		return;
	}
}

func main() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))));
	http.HandleFunc("/plugins", func(w http.ResponseWriter, r *http.Request) {
		u := getUser(r);
		if u == nil {
			w.WriteHeader(401);
			return;
		}
		w.Write(getPlugins(u).marshal());
	});


	http.HandleFunc("/plugins/", func(w http.ResponseWriter, r *http.Request) {
		u := getUser(r);
		if u == nil {
			w.WriteHeader(401);
			return;
		}
		if r.Method == "POST" {
			c := strings.Split(r.URL.Path, "/");
			if len(c) != 3 {
				w.WriteHeader(400);
				return;
			}
			r.ParseForm();
			updatePlugin(u, c[2], r.Form["betasecret"][0]);
			w.WriteHeader(200);
			return;
		}
	});


	http.HandleFunc("/versions/", func(w http.ResponseWriter, r *http.Request) {
		u := getUser(r);
		if u == nil {
			w.WriteHeader(401);
			return;
		}

		c := strings.Split(r.URL.Path, "/");

		if r.Method == "GET" {
			p := getVersions(c[2]);
			if p == nil {
				w.WriteHeader(404);
				return;
			}

			w.Write(p.marshal());
			return;
		}
		if r.Method == "POST" {
			r.ParseForm();
			switch(r.Form["op"][0]) {
			case "delete":
				deleteVersion(c[2], c[3]);
			case "approve":
				setStatus(u, c[2], c[3], "a", "");
			case "unapprove":
				setStatus(u, c[2], c[3], "p", "");
			case "reject":
				setStatus(u, c[2], c[3], "r", r.Form["reason"][0]);
			case "publish":
				setPublished(u, c[2], c[3], true);
			case "revoke":
				setPublished(u, c[2], c[3], false);
			default:
				w.WriteHeader(400);
			}
			w.Write([]byte("ok"));
			return;
		}
	});

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		u := getUser(r);
		if u == nil {
			w.WriteHeader(401);
			return;
		}

		file, _, err := r.FormFile("plugin");
		if err != nil {
			w.WriteHeader(400);
			return;
		}
		flen, _ := file.Seek(0, 2);
		pv, err := ingestFile(file, flen, u);

		w.Header().Set("Content-Type", "text/html; charset=utf-8");

		if err != nil {
			out, _ := json.Marshal(struct {
				Success bool `json:"success"`;
				Error string `json:"error"`;
			}{false, err.Error()});
			w.Write(out);
		} else {
			out, _ := json.Marshal(struct {
				Success bool `json:"success"`;
				Version *PluginVersion `json:"result"`;
			}{true, pv});
			w.Write(out);
		}
	});

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		c := new(http.Cookie);
		c.Name = "auth";
		c.Value = "";
		c.MaxAge = -1;
		http.SetCookie(w, c);
		http.Redirect(w, r, "spmc/", 302);
	});

	http.HandleFunc("/register.html", func(w http.ResponseWriter, r *http.Request) {
		filename := "static/register.html";
		file, err := os.Open(filename);
		if err != nil {
			http.NotFound(w, r);
		} else {
			io.Copy(w, file);
			file.Close();
		}
		return;
	});

	http.HandleFunc("/plugins-v1.json", func(w http.ResponseWriter, r *http.Request) {

		var reqver *Version;

		if len(r.Header["User-Agent"]) > 0 {
			vers := ua_re.FindStringSubmatch(r.Header["User-Agent"][0]);
			if len(vers) > 2 {
				reqver = new(Version);
				reqver.v[0], _ = strconv.Atoi(vers[1]);
				reqver.v[1], _ = strconv.Atoi(vers[2]);
				if len(vers) > 3 {
					reqver.v[2], _ = strconv.Atoi(vers[3]);
				}
			}
		}

		msg, err := buildShowtimeIndex(reqver, r.URL.Query()["betapassword"]);
		if err != nil {
			w.WriteHeader(400);
			io.WriteString(w, err.Error());
			return;
		}

		h := sha1.New();
		h.Write(msg);
		digest := fmt.Sprintf("%x", h.Sum(nil));

		if len(r.Header["If-None-Match"]) > 0 {
			if digest == r.Header["If-None-Match"][0] {
				fmt.Printf("Not modified\n");
				w.WriteHeader(304);
				return;
				}
		}
		
		w.Header().Set("ETag", digest);

		w.Write(msg);
	});

	http.HandleFunc("/data/", func(w http.ResponseWriter, r *http.Request) {
		c := strings.Split(r.URL.Path, "/");
		if len(c) != 3 {
			http.NotFound(w, r);
		} else {
			// filename is a digest of the contents so we set a 1 year expiry
			w.Header().Set("Cache-Control", "max-age=31536000");
			err := stashLoad(w, c[2]);
			if err != nil {
				w.WriteHeader(404);
				io.WriteString(w, err.Error());
				return;
			}
			dbIncDownloads(c[2]);
		}
	});

	http.HandleFunc("/spmc", roothandler);
	http.HandleFunc("/", roothandler);

	http.ListenAndServe("127.0.0.1:8080", httplog(http.DefaultServeMux));
}
