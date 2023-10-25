# Auth Daemon Microservice

## Description
Auth Daemon is a microservice designed to provide API key authentication. 
Incoming calls to the `prem-gateway` are rerouted by the traefik forward-auth middleware to the Auth Daemon. 
Once here, the Auth Daemon verifies the API key's validity. 
If the key is found to be valid, the request is forwarded to the appropriate service.

## Exposed Paths

### 1. Login
- **Path**: `/auth/login`
- **Method**: `GET`
- **Description**: Endpoint for logging in.
- **Query Parameters**:
    - `user`: The username.
    - `pass`: The password.
- **Response**:
    - `200 OK`: Contains the root `api_key`.
    - `401 Unauthorized`: Contains error message.

### 2. Verify Request
- **Path**: `/auth/verify`
- **Method**: `GET`
- **Description**: Verifies if the request is allowed.
- **Headers**:
    - `Authorization`: The root API key.
- **Response**:
    - `200 OK`: If the request is authorized.
    - `401 Unauthorized`: Contains error message.

### 3. Create API Key
- **Path**: `/auth/api-key`
- **Method**: `POST`
- **Description**: Endpoint to create a new API key.
- **Headers**:
    - `Authorization`: The root API key.
- **Body**: JSON object.
- **Response**:
    - `201 Created`: Contains the `api_key`.
    - `400 Bad Request`: Contains error message.
    - `500 Internal Server Error`: Contains error message.

### 4. Get Service API Key
- **Path**: `/auth/api-key/service`
- **Method**: `GET`
- **Description**: Retrieves the API key for a given service.
- **Headers**:
    - `Authorization`: The root API key.
- **Query Parameters**:
    - `name`: The service name.
- **Response**:
    - `200 OK`: Contains the `api_key`.
    - `500 Internal Server Error`: Contains error message.