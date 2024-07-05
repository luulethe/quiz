CREATE TABLE `quiz_tab` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `status` tinyint(1) UNSIGNED NOT NULL,
  `name` varchar(1024) NOT NULL,
  `created_time` bigint(20) UNSIGNED NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE `quiz_participant_tab` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `quiz_id` bigint(20),
    `user_id` bigint(20),
    `score` bigint(20),
    `created_time` bigint(20) UNSIGNED NOT NULL,
    `updated_time` bigint(20) UNSIGNED NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE UNIQUE INDEX quiz_participant_tab_quiz_user_index ON quiz_participant_tab (quiz_id, user_id);