import pytest
from app import app


@pytest.fixture
def client():
    app.config["TESTING"] = True
    with app.test_client() as client:
        yield client


@pytest.mark.parametrize("path,expected_status", [
    ("/health", 200),
    ("/", 200),
    ("/users/", 200),
])
def test_get_endpoints(client, path, expected_status):
    response = client.get(path)
    assert response.status_code == expected_status


@pytest.mark.parametrize("path,expected_content_type", [
    ("/health", "application/json"),
    ("/users/", "application/json"),
    ("/", "application/json"),
])
def test_content_type(client, path, expected_content_type):
    response = client.get(path)
    assert expected_content_type in response.content_type


@pytest.mark.parametrize("path,expected_keys", [
    ("/health", ["status"]),
    ("/", ["service", "version"]),
    ("/users/", ["id", "name", "email"]),
])
def test_response_json_keys(client, path, expected_keys):
    response = client.get(path)
    data = response.get_json()
    assert data is not None
    for key in expected_keys:
        assert key in data or all(key in item for item in data)


@pytest.mark.parametrize("user_id,expected_status", [
    (1, 200),
    (2, 200),
    (3, 200),
    (999, 404),
    (0, 404),
])
def test_get_user_by_id(client, user_id, expected_status):
    response = client.get(f"/users/{user_id}/")
    assert response.status_code == expected_status


@pytest.mark.parametrize("user_id,expected_fields", [
    (1, {"id": 1, "name": "Alice", "email": "alice@example.com"}),
    (2, {"id": 2, "name": "Bob", "email": "bob@example.com"}),
    (3, {"id": 3, "name": "Charlie", "email": "charlie@example.com"}),
])
def test_user_data(client, user_id, expected_fields):
    response = client.get(f"/users/{user_id}/")
    if response.status_code == 200:
        data = response.get_json()
        for key, value in expected_fields.items():
            assert data[key] == value


@pytest.mark.parametrize("path,method", [
    ("/health", "GET"),
    ("/users/", "GET"),
    ("/", "GET"),
])
def test_methods(client, path, method):
    response = client.open(path, method=method)
    assert response.status_code == 200


def test_health_returns_ok(client):
    response = client.get("/health")
    data = response.get_json()
    assert data["status"] == "ok"


def test_index_returns_service_info(client):
    response = client.get("/")
    data = response.get_json()
    assert "service" in data
    assert "version" in data
    assert data["service"] == "python-users"


def test_users_returns_list(client):
    response = client.get("/users/")
    data = response.get_json()
    assert isinstance(data, list)
    assert len(data) == 3


def test_nonexistent_user_returns_404(client):
    response = client.get("/users/999/")
    assert response.status_code == 404


def test_user_not_found_error_message(client):
    response = client.get("/users/999/")
    data = response.get_json()
    assert "error" in data
    assert data["error"] == "User not found"
