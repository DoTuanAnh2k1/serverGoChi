-- v2 schema — flat, no role hierarchy. See models/db_models for canonical
-- field types. v1 tables (tbl_account, cli_*) are NOT migrated; drop them
-- explicitly in dev/staging or run with a fresh DB. GORM auto-migrate
-- handles the v2 tables on startup; this file documents the intended shape
-- for code review and the Mongo index plan in pkg/repository/mongodb.
--
-- Authorization model in one sentence: "user X may execute command Y on NE
-- Z" iff (X enabled, not locked, not blacklisted) AND (X ∈ ne_access_group
-- containing Z) AND (X ∈ cmd_exec_group containing Y) AND (Y is registered
-- against Z's id).

DROP TABLE IF EXISTS `tbl_account`;
DROP TABLE IF EXISTS `cli_user_ne_mapping`;
DROP TABLE IF EXISTS `cli_user_group_mapping`;
DROP TABLE IF EXISTS `cli_group_ne_mapping`;
DROP TABLE IF EXISTS `cli_group_cmd_permission`;
DROP TABLE IF EXISTS `cli_group_mgt_permission`;
DROP TABLE IF EXISTS `cli_command_group_mapping`;
DROP TABLE IF EXISTS `cli_command_def`;
DROP TABLE IF EXISTS `cli_command_group`;
DROP TABLE IF EXISTS `cli_ne_profile`;
DROP TABLE IF EXISTS `cli_password_history`;
DROP TABLE IF EXISTS `cli_password_policy`;
DROP TABLE IF EXISTS `cli_group`;
DROP TABLE IF EXISTS `cli_ne`;
DROP TABLE IF EXISTS `cli_login_history`;
DROP TABLE IF EXISTS `cli_operation_history`;
DROP TABLE IF EXISTS `cli_config_backup`;
DROP TABLE IF EXISTS `cli_ne_monitor`;
DROP TABLE IF EXISTS `cli_ne_slave`;
DROP TABLE IF EXISTS `cli_ne_config`;

CREATE TABLE `user` (
  `id`                    bigint PRIMARY KEY AUTO_INCREMENT,
  `username`              varchar(64)  NOT NULL UNIQUE,
  `password_hash`         varchar(256) NOT NULL,
  `email`                 varchar(128),
  `full_name`             varchar(128),
  `phone`                 varchar(32),
  `is_enabled`            boolean      NOT NULL DEFAULT TRUE,
  `password_expires_at`   timestamp    NULL,
  `login_failure_count`   int          NOT NULL DEFAULT 0,
  `locked_at`             timestamp    NULL,
  `last_login_at`         timestamp    NULL,
  `created_at`            timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`            timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE `ne` (
  `id`            bigint PRIMARY KEY AUTO_INCREMENT,
  `namespace`     varchar(64)  NOT NULL UNIQUE,
  `ne_type`       varchar(32)  NOT NULL,
  `site_name`     varchar(64),
  `description`   varchar(255),
  `master_ip`     varchar(64),
  `master_port`   int,
  `ssh_username`  varchar(64),
  `ssh_password`  varchar(255),
  `command_url`   varchar(255),
  `conf_mode`     varchar(32)
);
CREATE INDEX `idx_ne_type` ON `ne` (`ne_type`);

CREATE TABLE `command` (
  `id`            bigint PRIMARY KEY AUTO_INCREMENT,
  `ne_id`         bigint       NOT NULL,
  `service`       varchar(16)  NOT NULL,   -- 'ne-config' | 'ne-command'
  `cmd_text`      varchar(512) NOT NULL,
  `description`   varchar(512),
  UNIQUE KEY `uq_command` (`ne_id`, `service`, `cmd_text`)
);
ALTER TABLE `command` ADD CONSTRAINT `fk_command_ne` FOREIGN KEY (`ne_id`) REFERENCES `ne` (`id`) ON DELETE CASCADE;

CREATE TABLE `ne_access_group` (
  `id`          bigint PRIMARY KEY AUTO_INCREMENT,
  `name`        varchar(64) NOT NULL UNIQUE,
  `description` varchar(255)
);
CREATE TABLE `ne_access_group_user` (
  `group_id` bigint NOT NULL,
  `user_id`  bigint NOT NULL,
  PRIMARY KEY (`group_id`, `user_id`)
);
CREATE TABLE `ne_access_group_ne` (
  `group_id` bigint NOT NULL,
  `ne_id`    bigint NOT NULL,
  PRIMARY KEY (`group_id`, `ne_id`)
);
ALTER TABLE `ne_access_group_user` ADD CONSTRAINT `fk_nag_user_group` FOREIGN KEY (`group_id`) REFERENCES `ne_access_group` (`id`) ON DELETE CASCADE;
ALTER TABLE `ne_access_group_user` ADD CONSTRAINT `fk_nag_user_user`  FOREIGN KEY (`user_id`)  REFERENCES `user` (`id`)            ON DELETE CASCADE;
ALTER TABLE `ne_access_group_ne`   ADD CONSTRAINT `fk_nag_ne_group`   FOREIGN KEY (`group_id`) REFERENCES `ne_access_group` (`id`) ON DELETE CASCADE;
ALTER TABLE `ne_access_group_ne`   ADD CONSTRAINT `fk_nag_ne_ne`      FOREIGN KEY (`ne_id`)    REFERENCES `ne` (`id`)              ON DELETE CASCADE;

CREATE TABLE `cmd_exec_group` (
  `id`          bigint PRIMARY KEY AUTO_INCREMENT,
  `name`        varchar(64) NOT NULL UNIQUE,
  `description` varchar(255)
);
CREATE TABLE `cmd_exec_group_user` (
  `group_id` bigint NOT NULL,
  `user_id`  bigint NOT NULL,
  PRIMARY KEY (`group_id`, `user_id`)
);
CREATE TABLE `cmd_exec_group_command` (
  `group_id`   bigint NOT NULL,
  `command_id` bigint NOT NULL,
  PRIMARY KEY (`group_id`, `command_id`)
);
ALTER TABLE `cmd_exec_group_user`    ADD CONSTRAINT `fk_ceg_user_group` FOREIGN KEY (`group_id`)   REFERENCES `cmd_exec_group` (`id`) ON DELETE CASCADE;
ALTER TABLE `cmd_exec_group_user`    ADD CONSTRAINT `fk_ceg_user_user`  FOREIGN KEY (`user_id`)    REFERENCES `user` (`id`)           ON DELETE CASCADE;
ALTER TABLE `cmd_exec_group_command` ADD CONSTRAINT `fk_ceg_cmd_group`  FOREIGN KEY (`group_id`)   REFERENCES `cmd_exec_group` (`id`) ON DELETE CASCADE;
ALTER TABLE `cmd_exec_group_command` ADD CONSTRAINT `fk_ceg_cmd_cmd`    FOREIGN KEY (`command_id`) REFERENCES `command` (`id`)        ON DELETE CASCADE;

-- Singleton (enforced by service.GetPasswordPolicy seeding id=1).
CREATE TABLE `password_policy` (
  `id`                bigint PRIMARY KEY AUTO_INCREMENT,
  `min_length`        int     NOT NULL DEFAULT 8,
  `max_age_days`      int     NOT NULL DEFAULT 0,
  `require_uppercase` boolean NOT NULL DEFAULT FALSE,
  `require_lowercase` boolean NOT NULL DEFAULT FALSE,
  `require_digit`     boolean NOT NULL DEFAULT FALSE,
  `require_special`   boolean NOT NULL DEFAULT FALSE,
  `history_count`     int     NOT NULL DEFAULT 0,
  `max_login_failure` int     NOT NULL DEFAULT 0,
  `lockout_minutes`   int     NOT NULL DEFAULT 0
);

CREATE TABLE `password_history` (
  `id`            bigint PRIMARY KEY AUTO_INCREMENT,
  `user_id`       bigint       NOT NULL,
  `password_hash` varchar(256) NOT NULL,
  `changed_at`    timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE `password_history` ADD CONSTRAINT `fk_pwh_user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`) ON DELETE CASCADE;
CREATE INDEX `idx_pwh_user_changed` ON `password_history` (`user_id`, `changed_at`);

CREATE TABLE `user_access_list` (
  `id`         bigint PRIMARY KEY AUTO_INCREMENT,
  `list_type`  varchar(16)  NOT NULL,    -- 'blacklist' | 'whitelist'
  `match_type` varchar(16)  NOT NULL,    -- 'username' | 'ip_cidr' | 'email_domain'
  `pattern`    varchar(255) NOT NULL,
  `reason`     varchar(255),
  `created_at` timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY `uq_user_acl` (`list_type`, `match_type`, `pattern`)
);

CREATE TABLE `operation_history` (
  `id`            int PRIMARY KEY AUTO_INCREMENT,
  `account`       varchar(64)  NOT NULL,
  `cmd_text`      varchar(512) NOT NULL,
  `ne_namespace`  varchar(64),
  `ne_ip`         varchar(64),
  `ip_address`    varchar(64),
  `scope`         varchar(16)  COMMENT 'ne-command, ne-config, mgt',
  `result`        varchar(255),
  `created_date`  timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `executed_time` timestamp    NULL
);
CREATE INDEX `idx_history_account`     ON `operation_history` (`account`);
CREATE INDEX `idx_history_created`     ON `operation_history` (`created_date`);
CREATE INDEX `idx_history_scope`       ON `operation_history` (`scope`);

CREATE TABLE `login_history` (
  `id`         int PRIMARY KEY AUTO_INCREMENT,
  `username`   varchar(64) NOT NULL,
  `ip_address` varchar(64),
  `time_login` timestamp   NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX `idx_login_username` ON `login_history` (`username`);

CREATE TABLE `config_backup` (
  `id`         bigint PRIMARY KEY AUTO_INCREMENT,
  `ne_name`    varchar(64) NOT NULL,
  `ne_ip`      varchar(64),
  `file_path`  varchar(255) NOT NULL,
  `size`       bigint,
  `created_at` timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX `idx_cfgbk_ne_name` ON `config_backup` (`ne_name`);
CREATE INDEX `idx_cfgbk_created` ON `config_backup` (`created_at`);
