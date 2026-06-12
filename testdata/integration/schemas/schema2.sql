DROP DATABASE IF EXISTS svcomp_target;
CREATE DATABASE svcomp_target;
USE svcomp_target;

CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(100) NOT NULL,
  active TINYINT(1) NOT NULL DEFAULT 1,
  PRIMARY KEY (id)
);

CREATE VIEW v_active_users AS
SELECT id, name
FROM users
WHERE active = 1;
