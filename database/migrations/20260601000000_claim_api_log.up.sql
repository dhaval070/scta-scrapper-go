CREATE TABLE claim_api_log (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  event_id VARCHAR(128) NOT NULL,
  site VARCHAR(255) NOT NULL,
  datetime DATETIME NOT NULL,
  surface_id INT NOT NULL,
  status TINYINT NOT NULL DEFAULT 0,
  http_status_code INT DEFAULT NULL,
  response_body TEXT DEFAULT NULL,
  error_message TEXT DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uq_log_event_id (event_id),
  INDEX idx_log_status (status),
  INDEX idx_log_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
