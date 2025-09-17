CREATE DATABASE `ai_transform` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;

use ai_transform;

CREATE TABLE `transform_records` (
   `id` bigint NOT NULL AUTO_INCREMENT,
   `user_id` bigint NOT NULL,
   `project_name` varchar(255) NOT NULL DEFAULT '',
   `original_language` varchar(32) NOT NULL DEFAULT '',
   `translated_language` varchar(32) NOT NULL DEFAULT '',
   `original_video_url` varchar(255) DEFAULT NULL,
   `original_srt_url` varchar(255) DEFAULT NULL,
   `translated_srt_url` varchar(255) DEFAULT NULL,
   `translated_video_url` varchar(255) DEFAULT NULL,
   `expiration_at` bigint NOT NULL DEFAULT '0',
   `create_at` bigint NOT NULL DEFAULT '0',
   `update_at` bigint NOT NULL DEFAULT '0',
   PRIMARY KEY (`id`),
   KEY `index_user_id` (`user_id`),
   KEY `index_expiration_at` (`expiration_at`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;