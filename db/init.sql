-- 智能售后服务系统 - 数据库初始化脚本

-- 设置字符集
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- 创建数据库（如不存在）
CREATE DATABASE IF NOT EXISTS `smart_aftercare`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `smart_aftercare`;

-- 文档表（GORM AutoMigrate 会自动创建，此处为参考）
-- CREATE TABLE IF NOT EXISTS `documents` (...);

-- 预置常见家电故障代码（示例数据）
CREATE TABLE IF NOT EXISTS `error_codes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `code` varchar(50) NOT NULL COMMENT '故障代码',
  `brand` varchar(100) NOT NULL COMMENT '品牌',
  `model` varchar(100) NOT NULL COMMENT '型号',
  `reason` text NOT NULL COMMENT '故障原因',
  `solution` text NOT NULL COMMENT '解决方案',
  `category` varchar(50) DEFAULT NULL COMMENT '故障分类',
  `severity` varchar(20) DEFAULT 'medium' COMMENT '严重程度',
  `source` varchar(255) DEFAULT NULL COMMENT '数据来源',
  PRIMARY KEY (`id`),
  KEY `idx_error_codes_code` (`code`),
  KEY `idx_error_codes_brand` (`brand`),
  KEY `idx_error_codes_model` (`model`),
  KEY `idx_error_codes_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入示例故障代码数据
INSERT INTO `error_codes` (`created_at`, `updated_at`, `code`, `brand`, `model`, `reason`, `solution`, `category`, `severity`, `source`) VALUES
(NOW(), NOW(), 'E1', '美的', 'KFR-35GW', '室内温度传感器故障', '1. 检查室内温度传感器连接是否松脱\n2. 用万用表测量传感器阻值（25°C时约10kΩ）\n3. 如阻值异常，更换温度传感器\n4. 如问题仍在，检查主板传感器接口', '传感器', 'medium', '美的空调维修手册'),
(NOW(), NOW(), 'E2', '美的', 'KFR-35GW', '室外温度传感器故障', '1. 检查室外机温度传感器接线\n2. 测量传感器阻值\n3. 更换传感器或检查主板', '传感器', 'medium', '美的空调维修手册'),
(NOW(), NOW(), 'E3', '美的', 'KFR-35GW', '压缩机排气温度过高保护', '1. 检查制冷剂是否不足（需专业加氟）\n2. 清洗冷凝器（室外机散热片）\n3. 检查室外风扇电机是否正常运转\n4. 检查压缩机是否老化', '压缩机', 'high', '美的空调维修手册'),
(NOW(), NOW(), 'F1', '格力', 'KFR-26GW', '室内环境温度传感器开路或短路', '1. 检查传感器插头连接\n2. 更换室内环境温度传感器\n3. 检查控制板', '传感器', 'medium', '格力空调维修手册'),
(NOW(), NOW(), 'F2', '格力', 'KFR-26GW', '室内蒸发器温度传感器故障', '1. 检查蒸发器温度传感器接线\n2. 测量传感器阻值\n3. 更换传感器', '传感器', 'medium', '格力空调维修手册'),
(NOW(), NOW(), 'H6', '格力', 'KFR-26GW', '无室内电机反馈', '1. 检查室内风机电机接线\n2. 检查电机运行电容\n3. 更换室内风机电机\n4. 检查主板电机驱动电路', '电机', 'high', '格力空调维修手册');
