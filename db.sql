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
  `account_type` int COMMENT '0-SuperAdmin; 1-Admin; 2-Normal',
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
  `conf_password` varchar(255)
);

CREATE TABLE `cli_role_user_mapping` (
  `user_id` bigint,
  `permission` varchar(255),
  PRIMARY KEY (`user_id`, `permission`)
);

CREATE TABLE `cli_user_ne_mapping` (
  `user_id` bigint,
  `tbl_ne_id` bigint,
  PRIMARY KEY (`user_id`, `tbl_ne_id`)
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

ALTER TABLE `cli_user_ne_mapping` ADD CONSTRAINT `user_ne` FOREIGN KEY (`user_id`) REFERENCES `tbl_account` (`account_id`);

ALTER TABLE `cli_user_ne_mapping` ADD CONSTRAINT `mapping_ne` FOREIGN KEY (`tbl_ne_id`) REFERENCES `cli_ne` (`id`);

ALTER TABLE `cli_role_user_mapping` ADD CONSTRAINT `user_role` FOREIGN KEY (`user_id`) REFERENCES `tbl_account` (`account_id`);

ALTER TABLE `cli_login_history` ADD FOREIGN KEY (`time_login`) REFERENCES `cli_login_history` (`user_name`);
