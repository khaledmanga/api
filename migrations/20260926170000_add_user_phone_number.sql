-- Add nullable phone number to "users".
ALTER TABLE `users` ADD COLUMN `phone_number` varchar(255) NULL;
