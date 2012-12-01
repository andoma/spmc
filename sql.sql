CREATE DATABASE showtime_plugins DEFAULT CHARACTER SET = 'utf8';
CREATE USER 'plugcentral' IDENTIFIED BY 'plugcentral';
GRANT ALL PRIVILEGES ON showtime_plugins.* TO 'plugcentral';


DROP TABLE version;
DROP TABLE plugin;
DROP TABLE users;

CREATE TABLE users (
       username VARCHAR(64) NOT NULL PRIMARY KEY,
       salt VARCHAR(32) NOT NULL,
       sha1 VARCHAR(64) NOT NULL,
       email TEXT NOT NULL,
       created TIMESTAMP DEFAULT NOW(),
       admin BOOL NOT NULL DEFAULT false,
       autoapprove BOOL NOT NULL DEFAULT false
) ENGINE InnoDB CHARACTER SET utf8 COLLATE utf8_general_ci;

CREATE TABLE plugin (
       id VARCHAR(128) NOT NULL PRIMARY KEY,
       created TIMESTAMP DEFAULT NOW(),
       owner VARCHAR(64) NOT NULL,
       pingsecret TEXT NOT NULL,
       pingstatus TEXT NOT NULL,
       FOREIGN KEY (owner) REFERENCES users(username) ON DELETE RESTRICT
) ENGINE InnoDB CHARACTER SET utf8 COLLATE utf8_general_ci;

CREATE TABLE version (
       plugin_id VARCHAR(128) NOT NULL,
       created TIMESTAMP DEFAULT NOW(),
       version VARCHAR(32) NOT NULL,
       type TEXT NOT NULL,
       author TEXT NOT NULL,
       downloads INT DEFAULT 0,
       showtime_min_version TEXT NOT NULL,
       title TEXT NOT NULL,
       category TEXT NOT NULL,
       synopsis TEXT NOT NULL,
       description TEXT NOT NULL,
       homepage TEXT NOT NULL,
       sha1 TEXT NOT NULL,
       approved BOOL NOT NULL DEFAULT false,
       published BOOL NOT NULL DEFAULT false,
       comment TEXT,
       UNIQUE(plugin_id, version),
       FOREIGN KEY (plugin_id) REFERENCES plugin(id) ON DELETE CASCADE

) ENGINE InnoDB CHARACTER SET utf8 COLLATE utf8_general_ci;

