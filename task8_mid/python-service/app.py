from flask import Flask, jsonify

app = Flask(__name__)

users_db = [
    {"id": 1, "name": "Alice", "email": "alice@example.com"},
    {"id": 2, "name": "Bob", "email": "bob@example.com"},
    {"id": 3, "name": "Charlie", "email": "charlie@example.com"},
]

@app.route("/")
def index():
    return jsonify({"service": "python-users", "version": "1.0"})

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
    return jsonify({"status": "ok"})

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
