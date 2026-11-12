# RelayX-Reality

Cloud Run WebSocket carrier for a byte-transparent Xray VLESS + XHTTP + REALITY connection.

## Architecture

Android Xray client -> local TCP-to-WebSocket adapter -> Cloud Run WSS -> server WebSocket-to-TCP bridge -> local Xray XHTTP + REALITY inbound -> freedom outbound.

The outer WebSocket is only the Cloud Run carrier. It forwards the inner TCP byte stream unchanged. Xray still performs VLESS, XHTTP, and REALITY end to end.

WebSocket + REALITY is not a native Xray transport combination. Xray documents REALITY with RAW, XHTTP, or gRPC. The project keeps WebSocket outside Xray as the Cloud Run carrier.

## Defaults

| Setting | Value |
| --- | --- |
| Xray | 26.11.9 |
| Docker image | DockerHub username / x-relay-real |
| XHTTP path | /reality |
| REALITY SNI | googleapis.com |
| REALITY target | googleapis.com:443 |
| XHTTP mode | auto |
| Xray server listener | 127.0.0.1:10000 |
| Cloud Run listener | PORT, default 8080 |
| Client adapter | 127.0.0.1:11000 |
| Local SOCKS5 | 127.0.0.1:10808 |

The local scripts/init.sh creates a .env file once. It does not overwrite existing credentials. Keep .env and generated client config private.

## Source layout

- cmd/relay: server, client adapter, config exporter, and Xray config validator.
- internal/bridge: WebSocket/TCP stream bridge, bearer authentication, session cap, and integration tests.
- internal/config: credential validation and Xray JSON generation.
- scripts: initialization, checks, client config export, client startup, and manual Cloud Run deployment.
- Dockerfile: Go bridge plus pinned Xray runtime.
- .github/workflows/ci.yml: source validation, Docker build, Xray config check, and Docker Hub publish on main only.

## CI validation

CI checks Go formatting, module integrity, go vet, race-enabled unit tests, WebSocket byte-stream integration, shell syntax, workflow syntax, Docker build, Xray version, and generated Xray server and client configs via xray run -test.

CI does not deploy to Cloud Run. A green run validates source/build/config checks, not end-to-end carrier or mobile-network connectivity.

Docker Hub publishing expects repository secrets named DOCKERHUB_USERNAME and DOCKERHUB_TOKEN. Images use the x-relay-real repository name.

## Client

1. Run scripts/init.sh once.
2. After Cloud Run exists, set REMOTE_WS_URL to wss://SERVICE_HOST/reality in .env.
3. Install Go and Xray in Termux.
4. Run scripts/start-client.sh.
5. Point the Android proxy app to SOCKS5 127.0.0.1:10808.

## Manual deployment

After the CI checks pass and the image has been published from main, run bash scripts/deploy.sh DOCKERHUB_USERNAME/x-relay-real:latest in Cloud Shell. The script reads the persistent credentials from .env. GitHub Actions never deploys to Cloud Run.

## Cloud Run behavior

Cloud Run supports WebSockets. WebSocket requests are subject to the service request timeout, up to 3600 seconds, so reconnect support is needed for long sessions. Do not enable end-to-end HTTP/2 for the outer WebSocket service.

Each outer WebSocket carries one TCP socket. Its binary frame boundaries are removed and the payload is copied to the inner TCP stream, preserving byte-stream order. XHTTP sessions and requests remain inside that stream.
