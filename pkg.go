package main

import "log"
import "os"
import "io"
import "encoding/json"
import "encoding/hex"
import "archive/zip"
import "errors"
import "crypto/sha1"


func pvFromZIPStream(r zip.Reader) (pv *PluginVersion, err error) {
	for _, f := range r.File {
		if f.Name == "plugin.json" {
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
	}
	err = errors.New("No 'plugin.json' found in archive");
	return;
}


func pvFromZIPFile(name string) (pv *PluginVersion, err error) {
	r, err := zip.OpenReader(name);
	if err != nil {
		log.Fatal(err);
		return;
	}
	defer r.Close();
	return pvFromZIPStream(r.Reader);
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



func (pv *PluginVersion) updateSHA1(filename string) (err error) {
	r, err := os.Open(filename);
	if err != nil {
		log.Println(err);
		return;
	}
	h := sha1.New();
	io.Copy(h, r);
	r.Close();
	pv.SHA1 = hex.EncodeToString(h.Sum(nil));
	return;
}







func ingestFile(f io.ReaderAt, size int64, u *User) (*PluginVersion, error) {
	r, err := zip.NewReader(f, size);
	if err != nil {
		return nil, err;
	}

	pv, err := pvFromZIPStream(*r);
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
