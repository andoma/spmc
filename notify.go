package main

import "net/smtp"
import "log"
import "bytes"
import "fmt"

func notifyUser(username string, subject string, body string) {

	loadConf();

	u, err := dbGetUser(username);
	if err != nil {
		log.Fatal(err);
		return;
	}

	if len(u.Email) == 0 {
		return;
	}

	if len(body) == 0 {
		body = "No further info";
	}

	go func() {
		c, err := smtp.Dial(config.Mail.Server);
		if err != nil {
			log.Fatal(err);
			return;
		}
		// Set the sender and recipient.
		c.Mail(config.Mail.Sender);
		c.Rcpt(u.Email);
		wc, err := c.Data();
		if err != nil {
			log.Fatal(err);
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
			log.Fatal(err);
			return;
		}
		log.Printf("Mail sent to %s <%s> : %s\n%s",
			username, u.Email, subject, body);
	}();
}