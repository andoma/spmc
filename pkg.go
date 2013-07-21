package main

import "log"
import "fmt"
import "os"
import "io"
import "bytes"
import "io/ioutil"
import "encoding/json"
import "encoding/hex"
import "archive/zip"
import "errors"
import "crypto/sha1"
import "net/http"
import "strings"

type zipblob struct {
	reader io.ReaderAt;
	size int64;
	


}


func findFile(r *zip.Reader, filename string) (f *zip.File) {
	for _, f := range r.File {
		if f.Name == filename {
			return f;
		}
	}
	return nil;
}

func removeLeadingDirInZipFile(r *zip.Reader) (zb *zipblob) {

	var pfx *string;

	for _, f := range r.File {
		if pfx == nil {
			if ! strings.HasSuffix(f.Name, "/") {
				return nil;
			}
			pfx = &f.Name;
		} else if ! strings.HasPrefix(f.Name, *pfx) {
			return nil;
		}
	}

	if pfx == nil {
		return nil;
	}
	
	pfxlen := len(*pfx);
	var b bytes.Buffer;
	out := zip.NewWriter(&b);
	for _, f := range r.File {
		newfilename := f.Name[pfxlen:];
		if len(newfilename) == 0 {
			continue;
		}
		w, _ := out.Create(newfilename);
		r, _ := f.Open();
		io.Copy(w, r);
		r.Close();
	}
	out.Close();

	return &zipblob{bytes.NewReader(b.Bytes()), int64(b.Len())};
}


func pvFromZIPblob(zb zipblob) (pv *PluginVersion, rzb zipblob, err error) {

	r, err := zip.NewReader(zb.reader, zb.size);
	if err != nil {
		return nil, zb, err;
	}
	
	f := findFile(r, "plugin.json");
	if f == nil {
		fz := removeLeadingDirInZipFile(r);
		if fz == nil {
			return nil, zb, errors.New("No 'plugin.json' found in archive");
		}
		r, err = zip.NewReader(fz.reader, fz.size);
		if err != nil {
			return nil, zb, err;
		}
		f = findFile(r, "plugin.json");
		if f == nil {
			return nil, zb, errors.New("No 'plugin.json' found in flatterned archive");
		}
		zb = *fz;
	}

	j, err := f.Open();
	if err != nil {
		log.Print(err);
		return nil, zb, err;
	}			
	pv = new(PluginVersion);
	err = json.NewDecoder(j).Decode(pv);
	j.Close();
	if err != nil {
		return nil, zb, err;
	}
	return pv, zb, nil;
}


func pvFromJSON(name string) (pv *PluginVersion, err error) {
	r, err := os.Open(name);
	if err == nil {
		pv = new(PluginVersion);
		err = json.NewDecoder(r).Decode(pv);
		if err != nil {
			return nil, err;
		}
		r.Close();
	}
	return;
}


func ingestFile(zb zipblob, u *User, chkp *string) (*PluginVersion, error) {

	pv, zb, err := pvFromZIPblob(zb);
	if err != nil {
		return nil, err;
	}

	pv.pkg_ver, err = parseVersionString(pv.Version);
	if err != nil {
		return nil, errors.New("'version' field in JSON -- " + err.Error());
	}

	pv.showtime_ver, err = parseVersionString(pv.ShowtimeVersion);
	if err != nil {
		return nil, errors.New("'showtimeVersion' field in JSON -- " + err.Error());
	}


	if chkp != nil && pv.PluginId != *chkp {
		return nil, errors.New(fmt.Sprintf("Plugin ID '%s' != '%s'",
			*chkp, pv.PluginId));
	}

	if u.Autoapprove {
		pv.Status = "a";
	} else {
		pv.Status = "p";
	}

	pv.Published = false;

	h := sha1.New();
	io.Copy(h, io.NewSectionReader(zb.reader, 0, zb.size));
	pv.PkgDigest = hex.EncodeToString(h.Sum(nil));

	if len(pv.Icon) > 0 {
		r, err := zip.NewReader(zb.reader, zb.size);
		if err != nil {
			return nil, err;
		}
		iconfile := findFile(r, pv.Icon);
		if iconfile != nil {
			f, _ := iconfile.Open();
			h := sha1.New();
			io.Copy(h, f);
			f.Close();
			pv.IconDigest = hex.EncodeToString(h.Sum(nil));

			f, _ = iconfile.Open();
			defer f.Close();
			err = stashSave(f, pv.IconDigest);
			if err != nil {
				return nil, err;
			}
		}
	}

	err = stashSave(io.NewSectionReader(zb.reader, 0, zb.size), pv.PkgDigest);
	if err != nil {
		return nil, err;
	}

	err = ingestVersion(pv, u);
	if err != nil {
		return nil, err;
	}
	return pv, nil;
}



func downloadFile(url string, u *User, pluginid *string) (*PluginVersion, error) {

	log.Printf("Downloading %s", url);
	resp, err := http.Get(url);
	if err != nil {
		return nil, err;
	}
	defer resp.Body.Close();

	body, err := ioutil.ReadAll(resp.Body);
	if err != nil {
		return nil, err;
	}

	zb := zipblob{bytes.NewReader(body), int64(len(body))};

	return ingestFile(zb, u, pluginid);
}

/*func ingestTest() {

	owner := "andoma";

	pv, _ := pvFromZIP(filename);
	pv.updateSHA1(filename);
	fmt.Printf("%s\n", pv);

	err := ingestVersion(pv, owner);
	fmt.Printf("%s\n", err);

	pv, _ := pvFromJSON("/home/andoma/showtime-plugin-svtplay/plugin.json");
	err := ingestVersion(pv, owner);
	fmt.Printf("%s\n", err);


	pv, _ := pvFromJSON("/home/andoma/showtime-plugin-headweb/plugin.json");

	
	owner := "andoma";

	cur := getPluginVersion(pv.PluginId, pv.Version);
	if cur != nil {
		fmt.Printf("Already got version %s for plugin %s", pv.Version, pv.PluginId);
	} else {
		ingestVersion(pv, owner);
	}
}
*/
