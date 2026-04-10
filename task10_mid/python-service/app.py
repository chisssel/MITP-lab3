import os
from flask import Flask, jsonify

app = Flask(__name__)

HOST = os.environ.get("FLASK_HOST", "0.0.0.0")
PORT = int(os.environ.get("FLASK_PORT", 5000))
DEBUG = os.environ.get("FLASK_DEBUG", "false").lower() == "true"
SERVICE_NAME = os.environ.get("SERVICE_NAME", "python-users")
SERVICE_VERSION = os.environ.get("SERVICE_VERSION", "1.0")

print(f"Configuration:")
print(f"  Host: {HOST}")
print(f"  Port: {PORT}")
print(f"  Debug: {DEBUG}")
print(f"  Service Name: {SERVICE_NAME}")
print(f"  Service Version: {SERVICE_VERSION}")

users_db = [
    {"id": 1, "name": "Alice", "email": "alice@example.com"},
    {"id": 2, "name": "Bob", "email": "bob@example.com"},
    {"id": 3, "name": "Charlie", "email": "charlie@example.com"},
]

@app.route("/")
def index():
    return jsonify({"service": SERVICE_NAME, "version": SERVICE_VERSION})

@app.route("/users/")
def get_users():
    return jsonify(users_db)

@app.route("/users/<int:user_id>/")
def get_user(user_id):
    user = next((u for u in users_db if u["id"] == user_id), None)
    if user:
        return jsonify(user)
    return jsonify({"error": "User not found"}), 404

@app.route("/health")
def health():
    return jsonify({"status": "ok", "service": SERVICE_NAME})

@app.route("/health/")
def health_slash():
    return jsonify({"status": "ok", "service": SERVICE_NAME})

if __name__ == "__main__":
    app.run(host=HOST, port=PORT, debug=DEBUG)
