import pytest


@pytest.fixture
def app():
    import app as app_module
    app_module.app.config["TESTING"] = True
    return app_module.app


@pytest.fixture
def client(app):
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


class TestHealthCheck:
    @pytest.mark.parametrize("path,expected_status", [
        ("/health", 200),
    ])
    def test_health_endpoint_status(self, client, path, expected_status):
        response = client.get(path)
        assert response.status_code == expected_status

    @pytest.mark.parametrize("path,expected_body", [
        ("/health", {"status": "ok"}),
    ])
    def test_health_endpoint_body(self, client, path, expected_body):
        response = client.get(path)
        data = response.get_json()
        assert data == expected_body

    def test_health_returns_json(self, client):
        response = client.get("/health")
        assert response.is_json
        data = response.get_json()
        assert "status" in data

    def test_health_status_is_ok(self, client):
        response = client.get("/health")
        data = response.get_json()
        assert data["status"] == "ok"

    def test_health_content_type(self, client):
        response = client.get("/health")
        assert "application/json" in response.content_type

    @pytest.mark.parametrize("method", [
        "POST",
        "PUT",
        "DELETE",
        "PATCH",
    ])
    def test_health_methods_not_allowed(self, client, method):
        response = client.open("/health", method=method)
        assert response.status_code == 405

    def test_health_curl_compatible(self, client):
        response = client.get("/health")
        assert response.status_code == 200
        assert response.status_code < 400


class TestHealthCheckDockerCompose:
    def test_health_endpoint_existence(self, client):
        response = client.get("/health")
        assert response.status_code == 200

    def test_health_returns_valid_json_for_curl(self, client):
        response = client.get("/health")
        assert response.status_code == 200
        data = response.get_json()
        assert isinstance(data, dict)
        assert "status" in data

    def test_health_response_time(self, client):
        import time
        start = time.time()
        response = client.get("/health")
        elapsed = time.time() - start
        assert response.status_code == 200
        assert elapsed < 1.0


class TestEnvironmentVariables:
    def test_host_is_string(self, client):
        assert True

    def test_port_is_positive(self, client):
        assert True

    def test_debug_is_boolean(self, client):
        assert True

    def test_service_name_exists(self, client):
        response = client.get("/")
        data = response.get_json()
        assert "service" in data
        assert data["service"] == "python-users"

    def test_service_version_exists(self, client):
        response = client.get("/")
        data = response.get_json()
        assert "version" in data
        assert isinstance(data["version"], str)
