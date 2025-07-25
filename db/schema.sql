CREATE DATABASE batch
USE batch

CREATE TABLE `reporting_status` (
  `id` int NOT NULL AUTO_INCREMENT,
  `application_no` varchar(20) NOT NULL,
  `candidate_name` varchar(100) NOT NULL,
  `branch` varchar(50) NOT NULL,
  `status` enum('Present','Reporting Slip Given') DEFAULT 'Present',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `application_no` (`application_no`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

CREATE TABLE `students` (
  `id` int NOT NULL AUTO_INCREMENT,
  `application_number` varchar(12) NOT NULL,
  `full_name` varchar(100) DEFAULT NULL,
  `father_name` varchar(100) DEFAULT NULL,
  `dob` date DEFAULT NULL,
  `gender` varchar(10) DEFAULT NULL,
  `contact_number` varchar(15) DEFAULT NULL,
  `email` varchar(100) DEFAULT NULL,
  `correspondence_address` text,
  `permanent_address` text,
  `branch` varchar(50) DEFAULT NULL,
  `lateral_entry` tinyint(1) DEFAULT NULL,
  `category` varchar(50) DEFAULT NULL,
  `sub_category` varchar(50) DEFAULT NULL,
  `exam_rank` int NOT NULL,
  `seat_quota` enum('Delhi','Outside Delhi') NOT NULL,
  `batch` varchar(10) DEFAULT NULL,
  `group_name` varchar(5) DEFAULT NULL,
  `status` enum('Reported','Withdrawn') DEFAULT 'Reported',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `last_upgrade_at` datetime DEFAULT NULL,
  `has_edited` tinyint(1) DEFAULT '0',
  `fee_mode` varchar(50) DEFAULT NULL,
  `fee_reference` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `application_number` (`application_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

CREATE TABLE `user_login` (
  `id` int NOT NULL AUTO_INCREMENT,
  `email` varchar(100) NOT NULL,
  `password_hash` text NOT NULL,
  `role` varchar(20) DEFAULT 'admin',
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

CREATE TABLE `uploads` (
  `id` int NOT NULL AUTO_INCREMENT,
  `application_number` varchar(20) DEFAULT NULL,
  `full_name` varchar(100) DEFAULT NULL,
  `photo_path` text,
  `jee_scorecard_path` text,
  `cet_admit_card_path` text,
  `candidate_profile_path` text,
  `fee_receipt_path` text,
  `reporting_slip_path` text,
  `uploaded_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

CREATE TABLE `notices` (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(255) NOT NULL,
  `description` text,
  `link` varchar(255) NOT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci

INSERT INTO user_login (email, password_hash, role)
VALUES (
  'studentcell@ipu.ac.in',
  SHA2('Student@123', 512),
  'admin'
);