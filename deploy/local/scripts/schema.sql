-- MGT Service schema for MySQL
-- Tables are auto-created by GORM AutoMigrate on startup.
-- Use this file only if you need to create tables manually.
--
-- Usage: mysql -u <user> -p <db_name> < scripts/schema.sql

CREATE TABLE IF NOT EXISTS tbl_account (
  account_id      BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  account_name    VARCHAR(255),
  password        VARCHAR(255),
  full_name       VARCHAR(255) DEFAULT '',
  email           VARCHAR(255) DEFAULT '',
  address         VARCHAR(255) DEFAULT '',
  phone_number    VARCHAR(50)  DEFAULT '',
  login_failure_count INT      DEFAULT 0,
  force_change_pass   TINYINT(1) DEFAULT 0,
  created_date    DATETIME(3)  NOT NULL,
  updated_date    DATETIME(3),
  last_login_time DATETIME(3),
  last_change_pass DATETIME(3),
  avatar          VARCHAR(255) DEFAULT '',
  status          TINYINT(1)   DEFAULT 1 COMMENT 'delete: 0, live: 1',
  default_dashboard INT        DEFAULT 0,
  account_type    INT          DEFAULT 2 COMMENT '0-SuperAdmin, 1-Admin, 2-Normal',
  auto_password   TINYINT(1)   DEFAULT 0,
  description     VARCHAR(255) DEFAULT '',
  is_enable       TINYINT(1)   DEFAULT 1 COMMENT '1: active, 0: inactive',
  created_by      VARCHAR(255) DEFAULT '',
  updated_by      VARCHAR(255) DEFAULT '',
  locked_time     DATETIME(3),
  first_login     BLOB,
  onlyAD          TINYINT(1)   DEFAULT 0
);

CREATE TABLE IF NOT EXISTS cli_ne (
  id          BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  description VARCHAR(255) DEFAULT '',
  ip_address  VARCHAR(255) DEFAULT '',
  name        VARCHAR(255) DEFAULT '',
  namespace   VARCHAR(255) DEFAULT '',
  site_name   VARCHAR(255) DEFAULT '',
  system_type VARCHAR(255) DEFAULT '',
  port        INT          DEFAULT 0,
  meta_data   TEXT
);

CREATE TABLE IF NOT EXISTS cli_ne_monitor (
  id        BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ne_ip     VARCHAR(255) DEFAULT '',
  tbl_ne_id BIGINT       DEFAULT 0
);

CREATE TABLE IF NOT EXISTS cli_ne_slave (
  id         BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  ip_address VARCHAR(255) DEFAULT '',
  port       INT          DEFAULT 0,
  tbl_ne_id  BIGINT       DEFAULT 0
);

CREATE TABLE IF NOT EXISTS cli_role (
  role_id      BIGINT       NOT NULL AUTO_INCREMENT PRIMARY KEY,
  include_type VARCHAR(255) NOT NULL,
  ne_type      VARCHAR(255) NOT NULL,
  scope        VARCHAR(255) NOT NULL,
  permission   VARCHAR(255) NOT NULL,
  path         VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS cli_role_user_mapping (
  user_id    BIGINT       NOT NULL,
  permission VARCHAR(191) NOT NULL,
  PRIMARY KEY (user_id, permission)
);

CREATE TABLE IF NOT EXISTS cli_user_ne_mapping (
  user_id   BIGINT NOT NULL,
  tbl_ne_id BIGINT NOT NULL,
  PRIMARY KEY (user_id, tbl_ne_id)
);

CREATE TABLE IF NOT EXISTS cli_operation_history (
  id               INT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
  is_oss_type      INT          DEFAULT 0,
  cmd_name         VARCHAR(255) NOT NULL,
  function_name    VARCHAR(255) DEFAULT 'Configuration Management',
  created_date     DATETIME(3)  NOT NULL,
  executed_time    DATETIME(3),
  ne_ip            VARCHAR(255) DEFAULT '',
  ne_name          VARCHAR(255) DEFAULT '',
  scope            VARCHAR(255) DEFAULT '',
  result           VARCHAR(255) DEFAULT '',
  account          VARCHAR(255) NOT NULL,
  ip_address       VARCHAR(255) DEFAULT '',
  input_type       VARCHAR(255) DEFAULT '',
  time_to_complete BIGINT       DEFAULT 0,
  ne_id            INT          DEFAULT 0,
  session          VARCHAR(255) DEFAULT '',
  batch_id         VARCHAR(255) DEFAULT ''
);

CREATE TABLE IF NOT EXISTS cli_login_history (
  id         INT          NOT NULL AUTO_INCREMENT PRIMARY KEY,
  user_name  VARCHAR(255) NOT NULL,
  ip_address VARCHAR(255) DEFAULT '',
  time_login DATETIME(3)
);
