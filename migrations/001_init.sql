CREATE TABLE IF NOT EXISTS `t_user` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `sn` VARCHAR(32) NOT NULL DEFAULT '',
    `name` VARCHAR(64) NOT NULL DEFAULT '',
    `password` VARCHAR(128) NOT NULL DEFAULT '',
    `nick_name` VARCHAR(64) NOT NULL DEFAULT '',
    `icon` VARCHAR(256) NOT NULL DEFAULT '',
    `gender` VARCHAR(4) NOT NULL DEFAULT 'N',
    `birth` VARCHAR(16) NOT NULL DEFAULT '',
    `create_from` VARCHAR(32) NOT NULL DEFAULT '',
    `location` VARCHAR(128) NOT NULL DEFAULT '',
    `city` VARCHAR(64) NOT NULL DEFAULT '',
    `wx_union_id` VARCHAR(64) NOT NULL DEFAULT '',
    `email` VARCHAR(64) NOT NULL DEFAULT '',
    `tel` VARCHAR(32) NOT NULL DEFAULT '',
    `create_at` BIGINT NOT NULL DEFAULT 0,
    `account_non_expired` TINYINT NOT NULL DEFAULT 1,
    `account_non_locked` TINYINT NOT NULL DEFAULT 1,
    `credentials_non_expired` TINYINT NOT NULL DEFAULT 1,
    `enabled` TINYINT NOT NULL DEFAULT 1,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_sn` (`sn`),
    KEY `idx_name` (`name`),
    KEY `idx_wx_union_id` (`wx_union_id`),
    KEY `idx_email` (`email`),
    KEY `idx_tel` (`tel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS `t_names` (
    `login_name` VARCHAR(128) NOT NULL,
    `user_id` BIGINT NOT NULL,
    `app_id` VARCHAR(32) NOT NULL DEFAULT '',
    `create_at` BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (`login_name`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
