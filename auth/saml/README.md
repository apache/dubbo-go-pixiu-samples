# SAML Authentication Sample

English | [中文](./README_CN.md)

## Overview

This sample shows how to protect a Pixiu HTTP route with the
`dgp.filter.http.auth.saml` filter.

The sample uses:

- Keycloak as the SAML Identity Provider (IdP)
- Pixiu as the SAML Service Provider (SP)
- a small backend server that echoes the SAML attributes forwarded by Pixiu

Once login succeeds, Pixiu forwards:

- `email -> X-User-Email`
- `displayName -> X-User-Name`

The backend returns those headers as JSON so you can confirm the SAML flow
works end to end.

## Files

```text
auth/
└── saml/
    ├── certs/
    │   ├── sp.crt
    │   └── sp.key
    ├── docker/
    │   ├── docker-compose.yml
    │   └── docker-health-check.sh
    ├── pixiu/
    │   └── conf.yaml
    ├── server/
    │   └── app/
    │       └── server.go
    └── test/
        └── pixiu_test.go
```

## Prerequisites

- Docker
- Go
- the `dubbo-go-pixiu` source tree on your machine so Pixiu can be started from source
- GNU Make and Bash if you want to render the sample config or run the full integration flow

## Step 1: Start Keycloak

```bash
cd auth/saml/docker
docker compose up -d
./docker-health-check.sh
```

Keycloak will be available at [http://localhost:18080](http://localhost:18080).

Default admin account:

- username: `admin`
- password: `admin`

## Step 2: Configure the Keycloak realm and SAML client

Open the Keycloak admin console and create the following resources.

### Create realm

- Realm name: `pixiu`

### Create test user

- Username: `alice`
- Email: `alice@example.com`
- First name: `Alice`
- Last name: `Pixiu`
- Password: `alice123`

### Create SAML client

- Client type / protocol: `SAML`
- Client ID: `pixiu-saml-sp`
- Name: `Pixiu SAML SP`
- Valid redirect URIs: `http://localhost:8888/*`
- Master SAML Processing URL: `http://localhost:8888/saml/acs`
- Home URL: `http://localhost:8888/app`

### Add protocol mappers

Add these attribute mappers so Pixiu can forward them to the backend:

1. Mapper for email
   - Mapper type: `User Property`
   - User property: `email`
   - SAML Attribute Name: `email`

2. Mapper for display name
   - Mapper type: `User Property`
   - User property: `firstName`
   - SAML Attribute Name: `displayName`

After saving the client, Keycloak publishes IdP metadata at:

```text
http://localhost:18080/realms/pixiu/protocol/saml/descriptor
```

This URL matches the `idp_metadata_url` used in `pixiu/conf.yaml`.

## Step 3: Start the backend server

```bash
cd auth/saml
go run server/app/*.go
```

The backend listens on `http://localhost:1314`.

## Step 4: Render the Pixiu config

The sample config uses `$PROJECT_DIR` for certificate paths, so render the
final config first:

```bash
cd /path/to/dubbo-go-pixiu-samples
make PROJECT_DIR=$(pwd)/auth/saml \
  PIXIU_DIR=/path/to/dubbo-go-pixiu \
  PROJECT_NAME=saml \
  BASE_DIR=$(pwd)/auth/saml/dist \
  -f igt/Makefile config
```

This generates a concrete config under
`auth/saml/dist/<os>_<arch>/pixiuconf/conf.yaml`.

## Step 5: Start Pixiu

Start Pixiu with the rendered config:

```bash
cd /path/to/dubbo-go-pixiu
go run cmd/pixiu/*.go gateway start -c /path/to/dubbo-go-pixiu-samples/auth/saml/dist/<os>_<arch>/pixiuconf/conf.yaml
```

## Step 6: Verify the sample

### Check the SP metadata endpoint

```bash
curl http://localhost:8888/saml/metadata
```

You should see XML containing:

- `pixiu-saml-sp`
- `AssertionConsumerService`

### Check the protected route

Open [http://localhost:8888/app](http://localhost:8888/app) in a browser.

Expected flow:

1. Pixiu redirects you to Keycloak
2. Sign in with `alice / alice123`
3. Keycloak posts the SAML response to Pixiu's ACS endpoint
4. Pixiu redirects back to `/app`
5. The backend returns JSON similar to:

```json
{
  "message": "saml login success",
  "email": "alice@example.com",
  "name": "Alice"
}
```

## Step 7: Run the smoke tests

```bash
go test -v ./auth/saml/test
```

The tests verify:

- the sample files are present
- the Pixiu config contains the SAML filter
- the metadata endpoint responds
- the protected route redirects unauthenticated users to Keycloak

## Step 8: Run the full integration flow

```bash
cd /path/to/dubbo-go-pixiu-samples
./integrate_test.sh auth/saml
```

## Notes

- This sample is designed for local HTTP development.
- `allow_idp_initiated: true` is enabled to make local HTTP testing easier.
- For production, prefer HTTPS and disable `allow_idp_initiated` unless you explicitly need that behavior.
