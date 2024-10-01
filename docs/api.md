# SHIROXY DYNAMIC API DOCUMENTATION

   <a href="https://www.npmjs.com/package/@icons-pack/react-simple-icons" target="_blank">
    <img src="https://img.shields.io/badge/verison_-v1-blue
    " alt="version" />
  </a>
   <a href="https://www.npmjs.com/package/@icons-pack/react-simple-icons" target="_blank">
    <img src="https://img.shields.io/badge/latest-red
    " alt="version" />
  </a>

This documentation includes everything that you need to start using shiroxy's dynamic API endpoint to controll the behaviour of shiroxy.

<details>
  <summary>
    <strong>Analytics</strong>
  </summary>

<div style="padding-left: 20px; padding-top: 10px; padding-bottom: 10px">
  <details>
  <summary>
    <strong>1. Fetch System Analytics</strong>
    
  </summary>

<!-- ### 1. Fetch System Analytics -->

#### **Description**

Returns system/machine analytics on which the shiroxy is running.

#### **URL**

`[HTTP Method]` `/v1/analytics/system`

#### **Method**

`GET`

#### **Headers**

| Key           | Value            | Description                 |
| ------------- | ---------------- | --------------------------- |
| Authorization | Bearer <token>   | Required for authentication |
| Content-Type  | application/json | Type of content being sent  |

<!-- #### **Request Parameters** -->

<!-- - **Path Parameters:** -->
  <!-- - `[param_name]` - Short description of the path parameter -->
<!-- - **Query Parameters:** -->
  <!-- - `?key=value` - Short description of the query parameter -->
<!-- - **Body Parameters:** -->

<!-- ```json
{
  "key": "value",
  "key2": "value2"
}
``` -->

#### **Response**

```json

```

</details>
<details>
  <summary>
    <strong>2. Fetch Domain Analytics</strong>
  </summary>

<!-- ### 2. Fetch Domain Analytics -->

#### **Description**

Returns details analytics of a domain

#### **URL**

[HTTP Method] /v1/analytics/domains

#### **Method**

`GET`

#### **Headers**

| Key           | Value            | Description                 |
| ------------- | ---------------- | --------------------------- |
| Authorization | Bearer <token>   | Required for authentication |
| Content-Type  | application/json | Type of content being sent  |

<!-- #### **Request Parameters** -->

<!-- - **Path Parameters:** -->
  <!-- - `[param_name]` - Short description of the path parameter -->
<!-- - **Query Parameters:** -->
  <!-- - `?key=value` - Short description of the query parameter -->
<!-- - **Body Parameters:** -->

<!-- ```json
{
  "key": "value",
  "key2": "value2"
}
``` -->

</details>
</div>

</details>

<details>
  <summary>
    <strong>Domains</strong>
  </summary>

  <div style="padding-left: 20px; padding-top: 10px; padding-bottom: 10px">
    <details>
      <summary><strong>1. Register</strong></summary>
      <div>

#### **Description**

Returns system/machine analytics on which the shiroxy is running.

#### **URL**

`[HTTP Method]` `/v1/domain`

#### **Method**

`POST`

#### **Headers**

| Key          | Value            | Description                |
| ------------ | ---------------- | -------------------------- |
| Content-Type | application/json | Type of content being sent |

#### **Request Parameters**

- **Body Parameters:**

```json
{
  "domain": "<domain-name>",
  "email": "<email>",
  "metadata": {
    "<key1>": "<value1>"
  }
}
```

#### **Response**

- **Status Code:**

| Status Code | Means   | Description          |
| ----------- | ------- | -------------------- |
| 200         | Success | Domain Created       |
| 400         | Error   | Invalid Request Body |
| 500         | Error   | Server Side Error    |

- **Response Body:**

```json

```

</div>

</details>
<details>
  <summary><strong>2. Retry SSL</strong></summary>
        <div>

#### **Description**

Retry SSL for a domain.

#### **URL**

`[HTTP Method]` `/v1/domain/<domain-name>/retryssl`

#### **Method**

`PATCH`

#### **Headers**

| Key          | Value            | Description                |
| ------------ | ---------------- | -------------------------- |
| Content-Type | application/json | Type of content being sent |

#### **Request Parameters**

- **Body Parameters:**

```json

```

#### **Response**

- **Status Code:**

| Status Code | Means   | Description                      |
| ----------- | ------- | -------------------------------- |
| 200         | Success | Retry SSL Successfully Requested |
| 400         | Error   | Invalid Request Body             |
| 500         | Error   | Server Side Error                |

- **Response Body:**

```json

```

</div>
</details>
<details>
  <summary><strong>3. Update One Domain</strong></summary>
  <div>

#### **Description**

Update Basic details of a domain.

#### **URL**

`[HTTP Method]` `/v1/domain/<domain-name>`

#### **Method**

`PATCH`

#### **Headers**

| Key          | Value            | Description                |
| ------------ | ---------------- | -------------------------- |
| Content-Type | application/json | Type of content being sent |

#### **Request Parameters**

- **Body Parameters:**

```json
  "metadata": {}
```

#### **Response**

- **Status Code:**

| Status Code | Means   | Description          |
| ----------- | ------- | -------------------- |
| 200         | Success | Update Successfull   |
| 400         | Error   | Invalid Request Body |
| 500         | Error   | Server Side Error    |

- **Response Body:**

```json

```

</div>
</details>
<details>
  <summary><strong>4. Fetch One Domain</strong></summary>
  <div>

#### **Description**

Update Basic details of a domain.

#### **URL**

`[HTTP Method]` `/v1/domain/<domain-name>`

#### **Method**

`GET`

#### **Headers**

| Key          | Value            | Description                |
| ------------ | ---------------- | -------------------------- |
| Content-Type | application/json | Type of content being sent |

#### **Request Parameters**

#### **Response**

- **Status Code:**

| Status Code | Means   | Description          |
| ----------- | ------- | -------------------- |
| 200         | Success | Update Successfull   |
| 400         | Error   | Invalid Request Body |
| 500         | Error   | Server Side Error    |

- **Response Body:**

```json

```

</div>
</details>
<details>
  <summary><strong>5. Remove One Domain</strong></summary>
</details>
  </div
</details>
  
<!-- ### 1. Fetch All Backends

#### **Description**

Briefly describe what this endpoint does.

#### **URL**

[HTTP Method] /v1/backends

#### **Method**

`GET`

#### **Headers**

| Key           | Value            | Description                 |
| ------------- | ---------------- | --------------------------- |
| Authorization | Bearer <token>   | Required for authentication |
| Content-Type  | application/json | Type of content being sent  |

#### **Request Parameters**

- **Path Parameters:**
  - `[param_name]` - Short description of the path parameter
- **Query Parameters:**
  - `?key=value` - Short description of the query parameter
- **Body Parameters:**

```json
{
  "key": "value",
  "key2": "value2"
}
```

--- -->

<details>
  <summary>
    <strong>Backend</strong>
  </summary>
  
  Your hidden content goes here.
  
</details>
