# Prometheus and Grafana

This folder contains a docker-compose file that will start a Prometheus and Grafana instance.
It is useful for testing GO Feature Flag with Prometheus and Grafana.

To start the containers, run:

```bash
docker-compose up -d
```

It will launch Prometheus on port 9090 and Grafana on port 3000.
You can access Grafana at http://localhost:3000 with the default credentials `admin`/`grafana`.

To stop the containers, run:

```bash
docker-compose down
```
