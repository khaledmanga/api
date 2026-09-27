CREATE TABLE `posts` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NOT NULL DEFAULT '',
  `content` longtext NOT NULL,
  `author_id` bigint NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `post_author_id` (`author_id`),
  KEY `post_created_at` (`created_at`)
) CHARSET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE `comments` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL,
  `author_id` bigint NOT NULL,
  `parent_id` bigint NULL,
  `content` longtext NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `comment_post_id_parent_id` (`post_id`, `parent_id`),
  KEY `comment_author_id` (`author_id`),
  CONSTRAINT `comments_comments_replies` FOREIGN KEY (`parent_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE,
  CONSTRAINT `comments_posts_comments` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_bin;

CREATE TABLE `post_likes` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `post_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `postlike_post_id_user_id` (`post_id`, `user_id`),
  KEY `postlike_user_id` (`user_id`),
  CONSTRAINT `post_likes_posts_likes` FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE
) CHARSET utf8mb4 COLLATE utf8mb4_bin;
