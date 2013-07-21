package main

import "net/smtp"
import "log"
import "bytes"
import "fmt"


func notify(u *User, subject string, body string) {

	loadConf();

	if len(u.Email) == 0 {
		return;
	}

	if len(body) == 0 {
		body = "No further info";
	}

	go func() {
		c, err := smtp.Dial(config.Mail.Server);
		if err != nil {
			log.Print(err);
			return;
		}
		// Set the sender and recipient.
		c.Mail(config.Mail.Sender);
		c.Rcpt(u.Email);
		wc, err := c.Data();
		if err != nil {
			log.Print(err);
			return;
		}

		hdrs := fmt.Sprintf("Subject: [SPMC] %s\n" +
			"To: %s <%s>\n" +
			"From: SPMC <%s>\n\n",
			subject,
			u.Username, u.Email,
			config.Mail.Sender);

		defer wc.Close();
		buf := bytes.NewBufferString(hdrs + body);
		if _, err = buf.WriteTo(wc); err != nil {
			log.Print(err);
			return;
		}
		log.Printf("Mail sent to %s <%s> : %s\n%s",
			u.Username, u.Email, subject, body);
	}();
}

func notifyUser(username string, subject string, body string) {


	u, err := dbGetUser(username);
	if err != nil {
		log.Print(err);
		return;
	}
	notify(u, subject, body);
}


func notifyAdmin(subject string, body string) {

	admins, err := dbGetAdmins();
	if err != nil {
		log.Print(err);
		return;
	}

	for _, a := range admins {
		notify(a, subject, body);
	}
}
