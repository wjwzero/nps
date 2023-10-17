DROP TABLE IF EXISTS `nps_client_info`;
CREATE TABLE `nps_client_info` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `verify_key` varchar(64) NOT NULL COMMENT '唯一验证key',
  `addr` varchar(64) DEFAULT NULL COMMENT 'ip地址',
  `basic_auth_user` varchar(64) DEFAULT NULL COMMENT 'Basic 认证用户名',
  `basic_auth_pass` varchar(64) DEFAULT NULL COMMENT 'Basic 认证密码',
  `product_key` varchar(64) DEFAULT NULL COMMENT '产品Key',
  `device_key` varchar(64) DEFAULT NULL COMMENT '设备Key',
  `version` varchar(64) DEFAULT NULL COMMENT '版本',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态 1:开放 0:关闭',
  `remark` varchar(256) DEFAULT NULL COMMENT '备注',
  `is_connect` tinyint NOT NULL DEFAULT '1' COMMENT '是否连接 1:在线 0:离线',
  `is_config_conn_allow` tinyint NOT NULL DEFAULT '1' COMMENT '是否允许客户端通过配置文件连接 1:是 0:否',
  `is_compress` tinyint NOT NULL DEFAULT '0' COMMENT '是否压缩 1:是 0:否',
  `is_crypt` tinyint NOT NULL DEFAULT '0' COMMENT '是否加密 1:是 0:否',
  `no_display` tinyint NOT NULL DEFAULT '0' COMMENT '在web页面是否不显示 1:是 0:否',
  `no_store` tinyint NOT NULL DEFAULT '0' COMMENT '是否不存储 1:是 0:否',
  `max_channel_num` int unsigned NOT NULL DEFAULT '0' COMMENT '最大隧道数',
  `max_connect_num` int unsigned NOT NULL DEFAULT '0' COMMENT '最大连接数',
  `rate_limit` bigint unsigned NOT NULL DEFAULT '0' COMMENT '带宽限制 kb/s',
  `flow_limit` bigint unsigned NOT NULL DEFAULT '0' COMMENT '流量限制 B',
  `web_user` varchar(64) DEFAULT NULL COMMENT 'web 登陆用户名',
  `web_pass` varchar(64) DEFAULT NULL COMMENT 'web 登陆密码',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_verify_key` (`verify_key`) USING BTREE
) ENGINE=InnoDB  DEFAULT CHARSET=utf8 COMMENT='客户端信息表';


DROP TABLE IF EXISTS `nps_client_statistic_connect`;
CREATE TABLE `nps_client_statistic_connect` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `client_id` int unsigned NOT NULL COMMENT '客户端主键',
  `now_connect_num` int unsigned NOT NULL DEFAULT '0' COMMENT '当前连接数',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_connect_client_id` (`client_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='客户端连接统计表';

DROP TABLE IF EXISTS `nps_client_statistic_flow`;
CREATE TABLE `nps_client_statistic_flow` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `client_id` int unsigned NOT NULL COMMENT '客户端主键',
  `flow_inlet` bigint unsigned NOT NULL DEFAULT '0' COMMENT '入口流量 B',
  `flow_export` bigint unsigned NOT NULL DEFAULT '0' COMMENT '出口流量 B',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_flow_client_id` (`client_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='客户端流量统计表';

DROP TABLE IF EXISTS `nps_client_statistic_rate`;
CREATE TABLE `nps_client_statistic_rate` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `client_id` int unsigned NOT NULL COMMENT '客户端主键',
  `rate_now` bigint unsigned NOT NULL DEFAULT '0' COMMENT '网速 B/S',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `idx_rate_client_id` (`client_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='客户端网速统计表';

DROP TABLE IF EXISTS `nps_client_host_info`;
CREATE TABLE `nps_client_host_info` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `client_id` int unsigned NOT NULL COMMENT '客户端主键',
  `remark` varchar(256) DEFAULT NULL COMMENT '备注',
  `host` varchar(64) DEFAULT NULL COMMENT '主机',
  `host_change` varchar(64) DEFAULT NULL COMMENT '请求主机信息修改',
  `cert_file_path` varchar(64) DEFAULT NULL COMMENT '证书文件路径',
  `key_file_path` varchar(64) DEFAULT NULL COMMENT '密钥文件路径',
  `header_change` varchar(256) DEFAULT NULL COMMENT '请求头部信息修改;多个冒号分割',
  `location` varchar(64) DEFAULT NULL COMMENT 'URL 路由',
  `no_store` tinyint NOT NULL DEFAULT '0' COMMENT '是否不存储 1:是 0:否',
  `is_close` tinyint NOT NULL DEFAULT '0' COMMENT '是否关闭 1:是 0:否',
  `scheme` varchar(32) DEFAULT NULL COMMENT '模式',
  `target_str` varchar(256) DEFAULT NULL COMMENT '目标 (IP:端口)',
  `is_local_proxy` tinyint NOT NULL DEFAULT '0' COMMENT '是否为本地代理 1:是 0:否',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_host_client_id` (`client_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='客户端主机信息表';

DROP TABLE IF EXISTS `nps_client_task_info`;
CREATE TABLE `nps_client_task_info` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `client_id` int unsigned NOT NULL COMMENT '客户端主键',
  `mode` varchar(16) DEFAULT NULL COMMENT 'p2p/tcp/udp/httpProxy/socks5/secret/file',
  `remark` varchar(256) DEFAULT NULL COMMENT '备注',
  `server_ip` varchar(64) DEFAULT NULL COMMENT '服务器ip',
  `port` int unsigned NOT NULL DEFAULT '0' COMMENT '端口',
  `password` varchar(64) DEFAULT NULL COMMENT '唯一标识密钥',
  `ports` varchar(64) DEFAULT NULL COMMENT '端口集合',
  `account` varchar(256) DEFAULT NULL COMMENT 'socks5账号',
  `target_addr` varchar(128) DEFAULT NULL COMMENT '内网目标',
  `local_path` varchar(64) DEFAULT NULL COMMENT '本地文件目录',
  `strip_pre` varchar(64) DEFAULT NULL COMMENT '前缀',
  `no_store` tinyint NOT NULL DEFAULT '0' COMMENT '是否不存储 1:是 0:否',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态 1:开放 0:关闭',
  `run_status` tinyint NOT NULL DEFAULT '0' COMMENT '状态 1:开放 0:关闭',
  `target_str` varchar(256) DEFAULT NULL COMMENT '目标 (IP:端口)',
  `is_local_proxy` tinyint NOT NULL DEFAULT '0' COMMENT '是否为本地代理 1:是 0:否',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
  `update_time` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
  PRIMARY KEY (`id`) USING BTREE,
  KEY `idx_task_client_id` (`client_id`) USING BTREE,
  KEY `idx_task_mode` (`mode`) USING BTREE,
  KEY `idx_password` (`password`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='客户端隧道信息表';
