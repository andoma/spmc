package main

import "errors"
import "io"
import "os"

var stashpath = "stash"


func stashSave(r io.Reader, digest string) error {
	if len(digest) != 40 {
		return errors.New("digest is not 40 chars");
	}

	basepath := stashpath + "/" + digest[0:2];

	err := os.MkdirAll(basepath, 0755);
	if err != nil {
		return err;
	}
	f, err := os.Create(basepath + "/" + digest[0:40]);
	if err != nil {
		return err;
	}
	_, err = io.Copy(f, r);
	if err != nil {
		return err;
	}

	return f.Close();
}



func stashLoad(w io.Writer, digest string) error {
	if len(digest) != 40 {
		return errors.New("digest is not 40 chars");
	}

	basepath := stashpath + "/" + digest[0:2];

	f, err := os.Open(basepath + "/" + digest[0:40]);
	if err != nil {
		return err;
	}
	_, err = io.Copy(w, f);
	if err != nil {
		return err;
	}

	return f.Close();
}
