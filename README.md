# 🦦 OtterMQ

[![Go](https://github.com/andrelcunha/OtterMq/actions/workflows/go.yml/badge.svg)](https://github.com/andrelcunha/OtterMq/actions/workflows/go.yml)
[![Docker Image CI](https://github.com/andrelcunha/ottermq/actions/workflows/docker-image.yml/badge.svg)](https://github.com/andrelcunha/ottermq/actions/workflows/docker-image.yml)
[![GitHub issues](https://img.shields.io/github/issues/andrelcunha/ottermq.svg)](https://github.com/andrelcunha/ottermq/issues)


**OtterMQ** is a high-performance message broker written in Go, inspired by RabbitMQ. It aims to provide a reliable, scalable, and easy-to-use messaging solution for distributed systems. OtterMQ is being developed with the goal of full compliance with the **AMQP 0.9.1 protocol**, ensuring compatibility with existing tools and workflows. It also features a modern management UI built with **Vue + Quasar**.

OtterMq already supports basic interoperability with RabbitMQ clients, including:
- [.NET RabbitMQ.Client](https://github.com/rabbitmq/rabbitmq-dotnet-client)
- [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)



## 🐾 About the Name
The name "OtterMQ" comes from my son's nickname and is a way to honor him. He brings joy and inspiration to my life, and this project is a reflection of that. And, of course, it is also a pun on **RabbitMQ**.

## ✨ Features
- AMQP-style Message Queuing
- Exchanges and Bindings
- Persistent Storage (in progress)
- High Availability (planned)
- Management Interface (Vue + Quasar)
- Docker Support via `docker-compose`
- RabbitMQ Client Compatibility (basic)

## ⚙️ Installation
### Broker Setup
```sh
git clone https://github.com/andrelcunha/ottermq.git
cd ottermq
make build && make install
```

### UI Setup (Vue + Quasar)
```sh
cd ottermq_ui
npm install
```

## 🚀 Usage
### Development Mode (UI runs separately)
```sh
# Run the broker:
ottermq

# Run the UI
cd ottermq_ui
quasar dev
```
### Integrated Mode (UI served by OtterMq)
1. Build the UI as a SPA:
```sh
quasar build
```
2. Link the built UI to the broker:
```sh
cd ..
ln -s ./ottermq_ui/dist/spa ./ui
```
Alternatively, copy the contents of dist/spa into a folder named ui at the project root (not recommended due to duplication).

3. Run the broker:
```sh
ottermq
```
OtterMq uses:

- Port **5672** for the AMQP broker

- Port **3000** for the management UI

## ⚙️ Configuration
OtterMQ can be configured using environment variables or a `.env` file. Environment variables take precedence over `.env` file settings, which in turn take precedence over default values.

### Configuration Options
Copy `.env.example` to `.env` and customize as needed:

```sh
cp .env.example .env
```

Available configuration options:

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `OTTERMQ_BROKER_PORT` | `5672` | AMQP broker port |
| `OTTERMQ_BROKER_HOST` | `` | Broker bind address (empty = all interfaces) |
| `OTTERMQ_USERNAME` | `guest` | Default username for authentication |
| `OTTERMQ_PASSWORD` | `guest` | Default password for authentication |
| `OTTERMQ_HEARTBEAT_INTERVAL` | `60` | Heartbeat interval in seconds |
| `OTTERMQ_CHANNEL_MAX` | `2048` | Maximum number of channels |
| `OTTERMQ_FRAME_MAX` | `131072` | Maximum frame size in bytes |
| `OTTERMQ_SSL` | `false` | Enable SSL/TLS |
| `OTTERMQ_QUEUE_BUFFER_SIZE` | `100000` | Queue message buffer size |
| `OTTERMQ_WEB_PORT` | `3000` | Web management UI port |
| `OTTERMQ_JWT_SECRET` | `secret` | JWT secret key for authentication |

### Example Configuration
Create a `.env` file in the project root:

```env
OTTERMQ_BROKER_PORT=5672
OTTERMQ_WEB_PORT=8080
OTTERMQ_USERNAME=admin
OTTERMQ_PASSWORD=secure_password
OTTERMQ_QUEUE_BUFFER_SIZE=200000
OTTERMQ_JWT_SECRET=my-secret-key
```

Or use environment variables directly:
```sh
export OTTERMQ_BROKER_PORT=15672
export OTTERMQ_WEB_PORT=8080
ottermq
```

## 🐳 Docker
You can run OtterMq using Docker:
```sh
docker compose up --build
```
This uses the provided `Dockerfile` and `docker-compose.yml` for convenience.

## 🚧 Development Status
OtterMq is under active development. While it follows the AMQP 0.9.1 protocol, several features are still in progress or not yet implemented, including:

- `basic.consume` command

- Message persistence flags

- Advanced routing and delivery guarantees

Basic compatibility with RabbitMQ clients is already functional, and more protocol features are being added incrementally.

## 📚 API Documentation
OtterMq provides a built-in Swagger UI for exploring and testing the API.

Access it at: `http://<server-address>/docs`

If you make changes to the API and need to regenerate the documentation, run:
```sh
make docs
```
This will update the Swagger spec and refresh the documentation served at `/docs`.

## ⚖️ License
OtterMQ is released under the MIT License. See [License](https://github.com/andrelcunha/ottermq/blob/master/LICENSE) for more information.

## 💬 Contact
For questions, suggestions, or issues, please open an issue in the [GitHub repository](https://github.com/andrelcunha/ottermq.git).
