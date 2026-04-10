use actix_web::{web, App, HttpResponse, HttpServer};
use serde::Serialize;
use std::sync::Mutex;

#[derive(Serialize)]
struct Stats {
    service: String,
    version: String,
    requests: u64,
    uptime_seconds: u64,
}

struct AppState {
    requests: Mutex<u64>,
    start_time: std::time::Instant,
}

async fn index(data: web::Data<AppState>) -> HttpResponse {
    let mut count = data.requests.lock().unwrap();
    *count += 1;
    
    let uptime = data.start_time.elapsed().as_secs();
    
    HttpResponse::Ok().json(Stats {
        service: "rust-stats".to_string(),
        version: "1.0".to_string(),
        requests: *count,
        uptime_seconds: uptime,
    })
}

async fn stats(data: web::Data<AppState>) -> HttpResponse {
    let mut count = data.requests.lock().unwrap();
    *count += 1;
    
    let uptime = data.start_time.elapsed().as_secs();
    
    HttpResponse::Ok().json(serde_json::json!({
        "total_requests": *count,
        "uptime_seconds": uptime,
        "average_per_second": if uptime > 0 { *count as f64 / uptime as f64 } else { 0.0 }
    }))
}

async fn health() -> HttpResponse {
    HttpResponse::Ok().json(serde_json::json!({"status": "ok"}))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let app_state = web::Data::new(AppState {
        requests: Mutex::new(0),
        start_time: std::time::Instant::now(),
    });

    println!("Rust Stats service starting on :4000");

    HttpServer::new(move || {
        App::new()
            .app_data(app_state.clone())
            .route("/", web::get().to(index))
            .route("/stats", web::get().to(stats))
            .route("/health", web::get().to(health))
    })
    .bind("0.0.0.0:4000")?
    .run()
    .await
}
