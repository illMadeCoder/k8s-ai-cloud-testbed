# HTTP Baseline Load Test Scenario
# Simple GET/POST mix to establish baseline metrics

from locust import HttpUser, task, between


class BaselineUser(HttpUser):
    """Simulates typical HTTP traffic patterns for baseline measurement."""

    wait_time = between(0.5, 2)  # Wait between requests

    @task(7)
    def get_root(self):
        """GET request to root endpoint - most common operation."""
        with self.client.get("/", catch_response=True) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Got status {response.status_code}")

    @task(2)
    def get_health(self):
        """GET request to health endpoint."""
        self.client.get("/health")

    @task(1)
    def post_echo(self):
        """POST request with JSON payload."""
        self.client.post(
            "/echo",
            json={"message": "baseline test", "timestamp": "now"},
            headers={"Content-Type": "application/json"},
        )

    def on_start(self):
        """Called when a user starts."""
        pass

    def on_stop(self):
        """Called when a user stops."""
        pass
