package main

import "log"
import "os"
import "io"
import "encoding/json"
import "encoding/hex"
import "archive/zip"
import "errors"
import "crypto/sha1"


func findFile(r *zip.Reader, filename string) (f *zip.File) {
	for _, f := range r.File {
		if f.Name == filename {
			return f;
		}
	}
	return nil;
	
}

func pvFromZIPStream(r *zip.Reader) (pv *PluginVersion, err error) {
	f := findFile(r, "plugin.json");
	if f == nil {
		return nil, errors.New("No 'plugin.json' found in archive");
	}

	j, err := f.Open();
	if err != nil {
		log.Fatal(err);
		return nil, err;
	}			
	pv = new(PluginVersion);
	err = json.NewDecoder(j).Decode(pv);
	j.Close();
	if err != nil {
		return nil, err;
	}
	return pv, nil;
}


func pvFromZIPFile(name string) (pv *PluginVersion, err error) {
	r, err := zip.OpenReader(name);
	if err != nil {
		log.Fatal(err);
		return;
	}
	defer r.Close();
	return pvFromZIPStream(&r.Reader);
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


func ingestFile(f io.ReaderAt, size int64, u *User) (*PluginVersion, error) {

	h := sha1.New();
	io.Copy(h, io.NewSectionReader(f, 0, size));
	digest := hex.EncodeToString(h.Sum(nil));

	r, err := zip.NewReader(f, size);
	if err != nil {
		return nil, err;
	}

	pv, err := pvFromZIPStream(r);
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


	pv.PkgDigest = digest;

	if len(pv.Icon) > 0 {
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

	err = stashSave(io.NewSectionReader(f, 0, size), pv.PkgDigest);
	if err != nil {
		return nil, err;
	}

	err = ingestVersion(pv, u);
	if err != nil {
		return nil, err;
	}
	return pv, nil;
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
