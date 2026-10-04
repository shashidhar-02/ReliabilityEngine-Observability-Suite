from fastapi.testclient import TestClient

from main import create_app


def test_http_payment_and_metrics():
    with TestClient(create_app(enable_tracing=False, simulate_failures=False)) as client:
        assert client.get("/health").json() == {"status": "healthy"}
        response = client.post("/process-payment")
        assert response.status_code == 200
        assert response.json()["transaction_id"].startswith("txn_")
        metrics = client.get("/metrics")
        assert metrics.status_code == 200
        assert 'path="/process-payment",status="200"} 1.0' in metrics.text
        assert "payments_http_request_duration_seconds_bucket" in metrics.text


def test_simulation_is_opt_in(monkeypatch):
    monkeypatch.setattr("main.random.random", lambda: 0)
    monkeypatch.setattr("main.random.uniform", lambda *_: 0)
    with TestClient(create_app(enable_tracing=False, simulate_failures=False)) as client:
        assert client.post("/process-payment").status_code == 200
    with TestClient(create_app(enable_tracing=False, simulate_failures=True)) as client:
        assert client.post("/process-payment").status_code == 500


def test_unknown_paths_have_bounded_labels():
    with TestClient(create_app(enable_tracing=False)) as client:
        for path in ("/unknown/one", "/unknown/two"):
            assert client.get(path).status_code == 404
        metrics = client.get("/metrics").text
        assert "unknown/one" not in metrics
        assert "unknown/two" not in metrics
        assert 'path="unmatched",status="404"} 2.0' in metrics
