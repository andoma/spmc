package main

import "io"
import "os"
import "net/http"
import "log"
import "strings"
import "encoding/json"


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
				setApproved(u, c[2], c[3], true);
			case "unapprove":
				setApproved(u, c[2], c[3], false);
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
			log.Printf("%s\n", pv);
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

	http.HandleFunc("/spmc", roothandler);
	http.HandleFunc("/", roothandler);

	http.ListenAndServe("127.0.0.1:8080", httplog(http.DefaultServeMux));
}
