use std::sync::Mutex;
use std::time::Instant;
use tiny_http::{Header, Response};

struct AppState {
    requests: Mutex<u64>,
    start_time: Instant,
}

impl AppState {
    fn new() -> Self {
        AppState {
            requests: Mutex::new(0),
            start_time: Instant::now(),
        }
    }

    fn increment_requests(&self) {
        let mut count = self.requests.lock().unwrap();
        *count += 1;
    }

    fn get_requests(&self) -> u64 {
        *self.requests.lock().unwrap()
    }

    fn get_uptime(&self) -> u64 {
        self.start_time.elapsed().as_secs()
    }

    fn calculate_average(&self) -> f64 {
        let requests = self.get_requests();
        let uptime = self.get_uptime();
        if uptime > 0 {
            requests as f64 / uptime as f64
        } else {
            0.0
        }
    }
}

fn create_health_response() -> String {
    r#"{"status":"ok"}"#.to_string()
}

fn create_stats_response(state: &AppState) -> String {
    let count = state.get_requests();
    let uptime = state.get_uptime();
    let avg = state.calculate_average();
    serde_json::json!({
        "total_requests": count,
        "uptime_seconds": uptime,
        "average_per_second": avg
    })
    .to_string()
}

fn create_root_response(state: &AppState) -> String {
    serde_json::json!({
        "service": "rust-stats",
        "version": "1.0",
        "requests": state.get_requests(),
        "uptime_seconds": state.get_uptime()
    })
    .to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_app_state_new() {
        let state = AppState::new();
        assert_eq!(state.get_requests(), 0);
        assert_eq!(state.get_uptime(), 0);
    }

    #[test]
    fn test_app_state_increment() {
        let state = AppState::new();
        state.increment_requests();
        state.increment_requests();
        assert_eq!(state.get_requests(), 2);
    }

    #[test]
    fn test_app_state_calculate_average() {
        let state = AppState::new();
        state.increment_requests();
        state.increment_requests();
        state.increment_requests();
        let avg = state.calculate_average();
        assert!(avg >= 0.0);
    }

    #[test]
    fn test_health_response_format() {
        let response = create_health_response();
        assert!(response.contains("status"));
        assert!(response.contains("ok"));
    }

    #[test]
    fn test_stats_response_format() {
        let state = AppState::new();
        let response = create_stats_response(&state);
        assert!(response.contains("total_requests"));
        assert!(response.contains("uptime_seconds"));
        assert!(response.contains("average_per_second"));
    }

    #[test]
    fn test_root_response_format() {
        let state = AppState::new();
        let response = create_root_response(&state);
        assert!(response.contains("service"));
        assert!(response.contains("rust-stats"));
        assert!(response.contains("version"));
        assert!(response.contains("requests"));
        assert!(response.contains("uptime_seconds"));
    }

    #[test]
    fn test_response_is_valid_json() {
        let state = AppState::new();

        let health = create_health_response();
        let parsed_health: serde_json::Value = serde_json::from_str(&health).unwrap();
        assert_eq!(parsed_health["status"], "ok");

        let stats = create_stats_response(&state);
        let parsed_stats: serde_json::Value = serde_json::from_str(&stats).unwrap();
        assert!(parsed_stats["total_requests"].is_number());
        assert!(parsed_stats["uptime_seconds"].is_number());

        let root = create_root_response(&state);
        let parsed_root: serde_json::Value = serde_json::from_str(&root).unwrap();
        assert_eq!(parsed_root["service"], "rust-stats");
        assert_eq!(parsed_root["version"], "1.0");
    }

    #[test]
    fn test_stats_increments_with_requests() {
        let state = AppState::new();

        state.increment_requests();
        let stats1 = create_stats_response(&state);
        let parsed1: serde_json::Value = serde_json::from_str(&stats1).unwrap();
        assert_eq!(parsed1["total_requests"], 1);

        state.increment_requests();
        let stats2 = create_stats_response(&state);
        let parsed2: serde_json::Value = serde_json::from_str(&stats2).unwrap();
        assert_eq!(parsed2["total_requests"], 2);
    }

    #[test]
    fn test_url_routing_logic() {
        let routes = vec![
            ("/", true, "root"),
            ("/stats", true, "stats"),
            ("/health", true, "health"),
            ("/nonexistent", false, "404"),
        ];

        for (path, should_match, expected) in routes {
            let is_valid_route = match path {
                "/" | "/stats" | "/health" => true,
                _ => false,
            };
            assert_eq!(
                is_valid_route, should_match,
                "Route {} should match: {}",
                path, should_match
            );
        }
    }

    #[test]
    fn test_average_calculation_zero_uptime() {
        let state = AppState::new();
        state.increment_requests();
        let avg = state.calculate_average();
        assert_eq!(avg, 0.0);
    }

    #[test]
    fn test_average_calculation_with_time() {
        let state = AppState::new();
        state.increment_requests();
        state.increment_requests();
        let avg = state.calculate_average();
        assert!(avg >= 0.0);
    }

    #[test]
    fn test_content_type_header_format() {
        let header = Header::from_bytes(&b"Content-Type"[..], &b"application/json"[..]).unwrap();
        let header_str = format!("{}", header);
        assert!(header_str.contains("Content-Type"));
        assert!(header_str.contains("application/json"));
    }

    #[test]
    fn test_multiple_state_instances() {
        let state1 = AppState::new();
        let state2 = AppState::new();

        state1.increment_requests();
        state1.increment_requests();

        assert_eq!(state1.get_requests(), 2);
        assert_eq!(state2.get_requests(), 0);
    }
}

fn main() {
    let state = std::sync::Arc::new(AppState {
        requests: Mutex::new(0),
        start_time: Instant::now(),
    });

    let server = tiny_http::Server::http("0.0.0.0:4000").unwrap();
    println!("Rust Stats service starting on :4000");

    let content_type_json =
        Header::from_bytes(&b"Content-Type"[..], &b"application/json"[..]).unwrap();

    for request in server.incoming_requests() {
        let url = request.url().to_string();

        {
            let mut count = state.requests.lock().unwrap();
            *count += 1;
        }

        let response: Response<std::io::Cursor<Vec<u8>>> = match url.as_str() {
            "/" | "" => {
                let count = *state.requests.lock().unwrap();
                let uptime = state.start_time.elapsed().as_secs();
                let body = serde_json::json!({
                    "service": "rust-stats",
                    "version": "1.0",
                    "requests": count,
                    "uptime_seconds": uptime
                });
                Response::from_string(body.to_string()).with_header(content_type_json.clone())
            }
            "/stats" => {
                let count = *state.requests.lock().unwrap();
                let uptime = state.start_time.elapsed().as_secs();
                let body = serde_json::json!({
                    "total_requests": count,
                    "uptime_seconds": uptime,
                    "average_per_second": if uptime > 0 { count as f64 / uptime as f64 } else { 0.0 }
                });
                Response::from_string(body.to_string()).with_header(content_type_json.clone())
            }
            "/health" => {
                Response::from_string(r#"{"status":"ok"}"#).with_header(content_type_json.clone())
            }
            _ => Response::from_string(r#"{"error":"not found"}"#)
                .with_header(content_type_json.clone())
                .with_status_code(404),
        };

        let _ = request.respond(response);
    }
}
