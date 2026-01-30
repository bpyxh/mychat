
CREATE DATABASE IF NOT EXISTS `mychat` CHARACTER SET utf8mb4;

CREATE TABLE IF NOT EXISTS `user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT '姓名',
  `email` varchar(128) NOT NULL DEFAULT '',
  `password` varchar(128) NOT NULL,
  `salt` varchar(30) NOT NULL,
  `phone` varchar(30) NOT NULL DEFAULT '',
  `login_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `logout_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` timestamp,
  PRIMARY KEY (`id`),
  UNIQUE KEY (`username`),
  UNIQUE KEY (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '用户表';

CREATE TABLE `message` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `msg_id` BIGINT UNSIGNED NOT NULL COMMENT '全局唯一消息ID（分布式雪花算法生成）',
    `conv_id` VARCHAR(64) NOT NULL COMMENT '会话ID，单聊为 uid1:uid2，群聊为 group_id',
    `from_id` BIGINT UNSIGNED NOT NULL COMMENT '发送者UID',
    `to_id` BIGINT UNSIGNED NOT NULL COMMENT '接收者ID（用户UID或群组ID）',
    `chat_type` TINYINT NOT NULL DEFAULT 1 COMMENT '聊天类型：1-单聊，2-群聊',
    `msg_type` TINYINT NOT NULL DEFAULT 1 COMMENT '消息类型：1-文本，2-图片，3-语音，4-视频，5-文件，6-撤回通知',
    `content` TEXT COMMENT '消息内容，如果是媒体文件，通常存储 URL 或 JSON 格式的元数据',
    `extra` JSON COMMENT '扩展字段',
    `seq_id` BIGINT UNSIGNED NOT NULL COMMENT '会话内连续递增序列号，用于多端同步和空洞检测',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp,

    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_msg_id` (`msg_id`),
    INDEX `idx_conv_seq` (`conv_id`, `seq_id`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';

CREATE TABLE `conv_member` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `uid` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    `conv_id` VARCHAR(64) NOT NULL COMMENT '会话ID',
    `chat_type` TINYINT NOT NULL COMMENT '1-单聊, 2-群聊',
    `last_read_seq` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '该用户在该会话已读到的最大seq_id',
    `custom_name` VARCHAR(64) COMMENT '好友备注或群昵称',
    `is_pinned` TINYINT DEFAULT 0 COMMENT '是否置顶',
    `is_muted` TINYINT DEFAULT 0 COMMENT '是否免打扰（消息不提醒）',
    `is_hidden` TINYINT DEFAULT 0 COMMENT '是否隐藏会话（不显示在列表）',
    -- `join_time` TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '加入会话时间',
    `status` TINYINT DEFAULT 1 COMMENT '成员状态：1-正常, 2-已退出, 3-被禁言',
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` timestamp,

    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_uid_conv` (`uid`, `conv_id`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户会话关系表';