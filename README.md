<div align="center">
  <img src="https://raw.githubusercontent.com/goipay/goipay.github.io/refs/heads/master/static/img/goipay-logo-with-name.svg" alt="Logo" width="400" height="100">

  ![Build Status](https://img.shields.io/github/actions/workflow/status/goipay/goipay/cd.yml)
  ![Version](https://img.shields.io/github/v/release/goipay/goipay)
  ![Docker Pulls](https://img.shields.io/docker/pulls/chekist32/goipay)
  ![License](https://img.shields.io/github/license/goipay/goipay)
</div>

## Description
> **Note:**
> The project is in development. This is not a release version.  
> As for now, only XMR invoices are implemented.

A lightweight crypto payment processor microservice, written in Golang, designed for creating and processing cryptocurrency invoices via gRPC.

## Getting Started
### Prerequisites
- Go ≥ 1.22
- PostgreSQL ≥ 12

### Installation
#### Docker
- Clone the repo
  ```sh
  git clone https://github.com/goipay/goipay.git
  ```
- Inside the root dir create and populate ```.env``` file on the base of ```.env.example``` file
  ```ini
  # Can be either 'prod' or 'dev'.
  # In 'dev' mode, a reflection server is established.
  MODE=dev

  SERVER_HOST=0.0.0.0
  SERVER_PORT=3000

  # As for now, only PostgreSQL is supported
  DATABASE_HOST=db
  DATABASE_PORT=5432
  DATABASE_USER=postgres
  DATABASE_PASS=postgres
  DATABASE_NAME=goipay_db

  XMR_DAEMON_URL=http://node.monerodevs.org:38089
  XMR_DAEMON_USER=
  XMR_DAEMON_PASS=
  ```
- Inside the root dir you can find an example ```docker-compose.yml``` file. For testing purposes can be run without editing.
  ```sh
  docker compose up
  ```

## Usage

- Get a quick overview of how GoiPay works by watching this [simple showcase video](https://youtu.be/b6TJBiHKJXE?feature=shared).
- Check out an [example project](https://github.com/goipay/example) to see GoiPay in action.
- For detailed information on using GoiPay's API, refer to the [API Reference](https://goipay.github.io/docs/api/grpc).

## Use cases

GoiPay is designed as a microservice that can be integrated into larger projects. If you need a simple, lightweight solution for just generating and processing crypto invoices, GoiPay is the perfect choice.