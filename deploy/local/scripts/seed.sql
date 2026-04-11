-- MGT Service sample data
-- Run AFTER schema.sql (or after auto-migrate)
--
-- Usage: mysql -u <user> -p <db_name> < scripts/seed.sql
--
-- Default admin password: admin123
-- Default user password:  Pass@123

-- Users
-- password = bcrypt(username + plaintext_password, cost=14)
INSERT IGNORE INTO tbl_account (account_name, password, account_type, is_enable, status, created_date, updated_date, last_login_time, last_change_pass, locked_time, created_by) VALUES
('admin',     '$2a$14$gU4u.6TlEGFmAGgspn.RPeYXW2sRRXcuhJ9MymrrDf/bZX6q/IOtG', 0, 1, 1, NOW(), NOW(), NOW(), NOW(), NOW(), 'seed'),
('operator1', '$2a$14$V03k0T5YSR/xfPA/g5288unCaWudVaf0Brvlc3ZH2UJfAWbNurtP6', 2, 1, 1, NOW(), NOW(), NOW(), NOW(), NOW(), 'seed'),
('operator2', '$2a$14$QwyxiKe6Qa3wFJXe/y/y4eDlx.qccEwLDVqRTkbesjJKFZ/cX.3fm', 2, 1, 1, NOW(), NOW(), NOW(), NOW(), NOW(), 'seed'),
('viewer1',   '$2a$14$O.ExrOkvhYOfAPnNQTgdV.aKw0lTxfK5oO6bomWOCbJ9jfyPE1ZmS', 2, 1, 1, NOW(), NOW(), NOW(), NOW(), NOW(), 'seed');

-- Network Elements (5GC core nodes)
INSERT INTO cli_ne (name, site_name, ip_address, port, system_type, namespace, description) VALUES
('HTSMF01', 'HCM', '10.10.1.1', 22, '5GC', 'hcm-5gc', 'HCM SMF Node 01'),
('HTSMF02', 'HCM', '10.10.1.2', 22, '5GC', 'hcm-5gc', 'HCM SMF Node 02'),
('HTAMF01', 'HCM', '10.10.2.1', 22, '5GC', 'hcm-5gc', 'HCM AMF Node 01'),
('HTAMF02', 'HCM', '10.10.2.2', 22, '5GC', 'hcm-5gc', 'HCM AMF Node 02'),
('HNSMF01', 'HN',  '10.20.1.1', 22, '5GC', 'hn-5gc',  'Ha Noi SMF Node 01'),
('HNAMF01', 'HN',  '10.20.2.1', 22, '5GC', 'hn-5gc',  'Ha Noi AMF Node 01'),
('DNUPF01', 'DN',  '10.30.3.1', 22, '5GC', 'dn-5gc',  'Da Nang UPF Node 01');

-- Roles / Permissions
INSERT INTO cli_role (include_type, ne_type, scope, permission, path) VALUES
('include', '5GC', 'ext-config', 'admin',    '/'),
('include', '5GC', 'ext-config', 'operator', '/'),
('include', '5GC', 'ext-config', 'viewer',   '/');

-- User-Role mappings (account_id from INSERT order above: admin=1, operator1=2, operator2=3, viewer1=4)
INSERT INTO cli_role_user_mapping (user_id, permission) VALUES
(1, 'admin'),
(2, 'operator'),
(3, 'operator'),
(4, 'viewer');

-- User-NE mappings (ne id from INSERT order above: HTSMF01=1..DNUPF01=7)
INSERT INTO cli_user_ne_mapping (user_id, tbl_ne_id) VALUES
(1, 1), (1, 2), (1, 3), (1, 4),   -- admin -> all HCM
(2, 5), (2, 6),                     -- operator1 -> HN
(3, 7),                             -- operator2 -> DN
(4, 7), (4, 6);                     -- viewer1 -> DN + HN AMF
