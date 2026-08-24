/*
 Navicat Premium Dump SQL

 Source Server         : local
 Source Server Type    : MySQL
 Source Server Version : 80043 (8.0.43)
 Source Host           : localhost:3306
 Source Schema         : tols

 Target Server Type    : MySQL
 Target Server Version : 80043 (8.0.43)
 File Encoding         : 65001

 Date: 19/12/2025 17:39:23
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for game_types
-- ----------------------------
DROP TABLE IF EXISTS `game_types`;
CREATE TABLE `game_types`  (
  `id` int UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '分类名称 (电子, 真人, 体育)',
  `sort` int NULL DEFAULT 0 COMMENT '排序权重',
  `code` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '代码字符串',
  `status` tinyint NULL DEFAULT 1 COMMENT '1:启用 0:禁用',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 12 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '游戏分类' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of game_types
-- ----------------------------
INSERT INTO `game_types` VALUES (1, 'Other(其他)', 1, 'OTHER', 1);
INSERT INTO `game_types` VALUES (2, 'Slots (老虎机)', 2, 'SLOT', 1);
INSERT INTO `game_types` VALUES (3, 'Live Casino (真人)', 3, 'CASINO', 1);
INSERT INTO `game_types` VALUES (4, 'Sports (体育)', 4, 'SPORTBOOK', 1);
INSERT INTO `game_types` VALUES (5, 'Keno', 5, 'KENO', 1);
INSERT INTO `game_types` VALUES (6, 'Lottery', 6, 'LOTTERY', 1);
INSERT INTO `game_types` VALUES (7, 'Fishing(捕鱼)', 7, 'FISHING', 1);
INSERT INTO `game_types` VALUES (8, 'Esport(电竞)', 8, 'ESPORT', 1);

SET FOREIGN_KEY_CHECKS = 1;
