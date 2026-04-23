CREATE TABLE `tbl_account` (
  `account_id` bigint PRIMARY KEY AUTO_INCREMENT,
  `account_name` varchar(255),
  `password` varchar(255),
  `full_name` varchar(255),
  `email` varchar(255),
  `address` varchar(255),
  `phone_number` varchar(255),
  `avatar` varchar(255),
  `description` varchar(255),
  `account_type` int COMMENT '0-SuperAdmin; 1-Admin; 2-Normal (maps to permission: 0/1=admin, 2=user)',
  `status` boolean COMMENT '0: deleted, 1: live',
  `is_enable` boolean COMMENT '1: active, 0: inactive',
  `force_change_pass` boolean,
  `auto_password` boolean,
  `only_ad` boolean,
  `first_login` blob,
  `default_dashboard` int,
  `login_failure_count` int,
  `created_date` timestamp NOT NULL,
  `updated_date` timestamp,
  `last_login_time` timestamp,
  `last_change_pass` timestamp,
  `locked_time` timestamp,
  `created_by` varchar(255),
  `updated_by` varchar(255)
);

CREATE TABLE `cli_login_history` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `user_name` varchar(255) NOT NULL,
  `ip_address` varchar(255),
  `time_login` timestamp
);

CREATE TABLE `cli_ne` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `ne_name` varchar(255),
  `namespace` varchar(255),
  `site_name` varchar(255),
  `system_type` varchar(255),
  `description` varchar(255),
  `command_url` varchar(255),
  `conf_mode` varchar(255),
  `conf_master_ip` varchar(255),
  `conf_slave_ip` varchar(255),
  `conf_port_master_ssh` int,
  `conf_port_slave_ssh` int,
  `conf_port_master_tcp` int,
  `conf_port_slave_tcp` int,
  `conf_username` varchar(255),
  `conf_password` varchar(255),
  `ne_profile_id` bigint  -- classifies the NE by command set (SMF/AMF/UPF/...)
);

CREATE TABLE `cli_user_ne_mapping` (
  `user_id` bigint,
  `tbl_ne_id` bigint,
  PRIMARY KEY (`user_id`, `tbl_ne_id`)
);

CREATE TABLE `cli_group` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(255) NOT NULL UNIQUE,
  `description` varchar(255)
);

CREATE TABLE `cli_user_group_mapping` (
  `user_id` bigint,
  `group_id` bigint,
  PRIMARY KEY (`user_id`, `group_id`)
);

CREATE TABLE `cli_group_ne_mapping` (
  `group_id` bigint,
  `tbl_ne_id` bigint,
  PRIMARY KEY (`group_id`, `tbl_ne_id`)
);

CREATE TABLE `cli_operation_history` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `account` varchar(255) NOT NULL,
  `cmd_name` varchar(255) NOT NULL,
  `ne_name` varchar(255),
  `ne_ip` varchar(255),
  `ip_address` varchar(255),
  `scope` varchar(255) COMMENT 'necommand, neconfig, cliconfig',
  `result` varchar(255),
  `created_date` timestamp NOT NULL,
  `executed_time` timestamp
);

CREATE TABLE `cli_config_backup` (
  `id`         bigint PRIMARY KEY AUTO_INCREMENT,
  `ne_name`    varchar(255) NOT NULL,
  `ne_ip`      varchar(255),
  `file_path`  varchar(255) NOT NULL COMMENT 'đường dẫn file XML trên disk',
  `size`       bigint       COMMENT 'kích thước file tính bằng byte',
  `created_at` timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE `cli_user_ne_mapping` ADD CONSTRAINT `user_ne` FOREIGN KEY (`user_id`) REFERENCES `tbl_account` (`account_id`);
ALTER TABLE `cli_user_ne_mapping` ADD CONSTRAINT `mapping_ne` FOREIGN KEY (`tbl_ne_id`) REFERENCES `cli_ne` (`id`);
ALTER TABLE `cli_user_group_mapping` ADD CONSTRAINT `user_group_user` FOREIGN KEY (`user_id`) REFERENCES `tbl_account` (`account_id`);
ALTER TABLE `cli_user_group_mapping` ADD CONSTRAINT `user_group_group` FOREIGN KEY (`group_id`) REFERENCES `cli_group` (`id`);
ALTER TABLE `cli_group_ne_mapping` ADD CONSTRAINT `group_ne_group` FOREIGN KEY (`group_id`) REFERENCES `cli_group` (`id`);
ALTER TABLE `cli_group_ne_mapping` ADD CONSTRAINT `group_ne_ne` FOREIGN KEY (`tbl_ne_id`) REFERENCES `cli_ne` (`id`);
-- Email must be unique when present; empty string must be stored as NULL (MySQL allows multiple NULLs per UNIQUE).
ALTER TABLE `tbl_account` ADD CONSTRAINT `uq_tbl_account_email` UNIQUE (`email`);

-- ─────────────────────────────────────────────────────────────────────────
-- RBAC (docs/rbac-design.md): NE Profile + Command Registry + Group→Cmd
-- perm rules. Decoupled from tbl_account.account_type for backward compat;
-- the legacy account_type is still used by CheckRole middleware while the
-- fine-grained permissions are evaluated by service/rbac.
-- ─────────────────────────────────────────────────────────────────────────

CREATE TABLE `cli_ne_profile` (
  `id`          bigint PRIMARY KEY AUTO_INCREMENT,
  `name`        varchar(64)  NOT NULL UNIQUE,
  `description` varchar(512)
);

CREATE TABLE `cli_command_def` (
  `id`          bigint PRIMARY KEY AUTO_INCREMENT,
  `service`     varchar(32)  NOT NULL,     -- 'ne-command' | 'ne-config' | '*'
  `ne_profile`  varchar(64)  NOT NULL DEFAULT '*',
  `pattern`     varchar(256) NOT NULL,
  `category`    varchar(32)  NOT NULL,     -- 'monitoring' | 'configuration' | 'admin' | 'debug'
  `risk_level`  int          NOT NULL DEFAULT 0,
  `description` varchar(512),
  `created_by`  varchar(64),
  UNIQUE KEY `uq_cmd_def_service_profile_pattern` (`service`, `ne_profile`, `pattern`)
);

CREATE TABLE `cli_command_group` (
  `id`          bigint PRIMARY KEY AUTO_INCREMENT,
  `name`        varchar(64)  NOT NULL UNIQUE,
  `ne_profile`  varchar(64)  NOT NULL DEFAULT '*',
  `service`     varchar(32)  NOT NULL DEFAULT '*',
  `description` varchar(512),
  `created_by`  varchar(64)
);

CREATE TABLE `cli_command_group_mapping` (
  `command_group_id` bigint NOT NULL,
  `command_def_id`   bigint NOT NULL,
  PRIMARY KEY (`command_group_id`, `command_def_id`)
);

CREATE TABLE `cli_group_cmd_permission` (
  `id`          bigint PRIMARY KEY AUTO_INCREMENT,
  `group_id`    bigint       NOT NULL,
  `service`     varchar(32)  NOT NULL,     -- 'ne-command' | 'ne-config' | '*'
  `ne_scope`    varchar(128) NOT NULL DEFAULT '*',
  `grant_type`  varchar(16)  NOT NULL,     -- 'command_group' | 'category' | 'pattern'
  `grant_value` varchar(256) NOT NULL,
  `effect`      varchar(8)   NOT NULL,     -- 'allow' | 'deny'
  UNIQUE KEY `uq_group_cmd_perm` (`group_id`, `service`, `ne_scope`, `grant_type`, `grant_value`)
);

ALTER TABLE `cli_ne`                 ADD CONSTRAINT `ne_profile_fk`    FOREIGN KEY (`ne_profile_id`)    REFERENCES `cli_ne_profile` (`id`);
ALTER TABLE `cli_command_group_mapping` ADD CONSTRAINT `cgm_group_fk`  FOREIGN KEY (`command_group_id`) REFERENCES `cli_command_group` (`id`) ON DELETE CASCADE;
ALTER TABLE `cli_command_group_mapping` ADD CONSTRAINT `cgm_def_fk`    FOREIGN KEY (`command_def_id`)   REFERENCES `cli_command_def`   (`id`) ON DELETE CASCADE;
ALTER TABLE `cli_group_cmd_permission`  ADD CONSTRAINT `gcp_group_fk`  FOREIGN KEY (`group_id`)         REFERENCES `cli_group` (`id`)        ON DELETE CASCADE;
