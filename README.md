<div align="center">

<!-- <h1>IN DEVELOPMENT</h1> -->
<img src="https://raw.githubusercontent.com/ShikharY10/shiroxy/main/media/shiroxy_logo_.png" alt="shiroxy Logo">

  <h1>A Reverse Proxy with Multiple Domains, Automatic SSL and Dynamic Routing</h1>
</div>
<!-- <hr> -->

Welcome to the Shirxoy! This Go-based reverse proxy is designed to provide seamless and secure web traffic management with the following key features:

## Key Features

- **Automatic SSL Certificates**: Effortlessly handle SSL certificates for a finite number of domain names, ensuring secure connections.
- **Custom Traffic Routing**: Implement custom logic to route traffic for specific domains to designated locations.
- **Dynamic Domain Management**: Utilize REST API endpoints to dynamically add and remove domains, enabling flexible and scalable domain management.
- **System and Process Analytics**: Access detailed system and process analytics through REST APIs, providing valuable insights into the performance and health of the system running the proxy.

## Table of Contents

- [Getting Started](#getting-started)
- [Running shiroxy](#running-shiroxy)
- [Configuration File](#configuration-file)
  - [Default Section](#default-section)
  - [Frontend Section](#frontend-section)
  - [Backend Section](#backend-section)
  - [Logging Section](#logging-section)
  - [Webhook Section](#webhook-section)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [License](#license)

## Getting Started

To get started with shiroxy, clone the repository and follow the installation instructions.

```bash
git clone https://github.com/yourusername/shiroxy.git
cd shiroxy
go build -o build/shiroxy

./shiroxy -config /path/to/your/config.yml
```

## Configuration File

The configuration file for shiroxy is written in YAML. It is recommended to make a copy of the default configuration file and update the settings according to your needs.

### Default Configuration File

```yaml
#        ##### #   # # #####  ####  #   #  #   #
#       #     #   # # #   # #    #  # #    # #
#      ##### ##### # ##### #    #   #      #
#         # #   # # # #   #    #  # #    #
#    ##### #   # # #  #   ####  #   #  #

###### Default Configuration File ######

# We recommend making a copy of this file and then updating any
# settings in this file according to your needs.

# Version of this configuration file
version: "1.0.0"

# Default section of the configuration. It contains settings
# that are global to shiroxy
default:
  # The mode in which shiroxy will be running.
  mode: "http"

  # Path where the logs will be stored. It is recommended not to
  # change the default path unless you know what you are doing.
  logpath: "/home/shikharcode/Main/opensource/shiroxy"

  # Path where the data will be persisted in case of a failure
  # or manual stop. It is recommended not to change the default
  # path unless you know what you are doing.
  datapersistancepath: "/home/shikharcode/Main/opensource/shiroxy"

  # This configuration tells shiroxy whether to start the DNS challenge
  # solver. It starts an HTTP server that listens on port 80.
  enablednschallengesolver: ""

  # Timeout specifies the timeout for different scenarios
  timeout:
    # This sets the maximum time to wait for a connection to a
    # backend server to be established.
    connect: "10s"

    # This sets the maximum inactivity time on the client side.
    client: "10s"

    # This sets the maximum inactivity time on the server side.
    server: "10s"

  # This specifies how the SSL certificates will be stored.
  # You have to specify this setting when you set the secure
  # target to "multiple" or set singletargetmode to "shiroxysinglesecure"
  storage:
    # This sets where you want to store your SSL certificates. You can
    # store the certificates either in Redis or let shiroxy handle the storage.
    # Possible values for location are "redis" and "memory".
    # "redis" means storing the certificates in Redis for speed,
    # reliability, and security. "memory" means shiroxy will handle the
    # storage process and store the certificates in primary memory
    # for faster retrieval.
    location: "memory"

    # This sets the host of the Redis server
    # redis_host: ""

    # This sets the port of the Redis server
    # redis_port: ""

  # This section specifies settings related to analytics
  analytics:
    # This sets the interval after which analytics will be recorded.
    # This value is in seconds. For example, if set to 10, analytics
    # will be recorded and saved every 10 seconds.
    collectioninterval: 10

    # This specifies the base path for analytics-related API calls
    routename: "/analytics"

  # On the error page, there is a button, and we allow modifying how
  # that button should behave. This specifies what the button label will
  # look like and what happens when someone clicks on that button.
  errorresponses:
    # This sets the label of that button
    errorpagebuttonname: "Comeata"

    # This sets the URL that the button leads to after someone
    # clicks on it.
    errorpagebuttonurl: "https://youtube.com"

  user:
    email: "yshikharfzd10@gmail.com"
    secret: "kjksdnfiwj"

# This section specifies settings related to the frontend of the reverse proxy
frontend:
  backend: "shiroxy-test"
  bind:
    # This is the port on which shiroxy will listen
    port: "8080"

    # Host on which shiroxy will listen
    host: ""

    # Domain SSL configuration
    secure:
      # Target specifies how many domains you want to secure.
      # If you want to secure only one domain, set the value of
      # target to "single". If you want to secure more than one
      # domain, set the value of target to "multiple".
      target: "single"
      # target: "multiple"

      # If you set the value of target to "single", you have to
      # specify how you want to secure the single domain.
      # Set the value of singletargetmode to either
      # "certandkey" or "shiroxysinglesecure". For "certandkey",
      # you have to provide the SSL certificate. For "shiroxysinglesecure",
      # you just need to specify the domain you want to secure, and
      # shiroxy will handle everything.
      singletargetmode: "certandkey"

      # This sets the location of the SSL certificate and SSL key.
      certandkey:
        # This sets the certificate's location
        cert: "cert-location"

        # This sets the key's location
        key: "key-location"

      # This sets which domain name you want to secure in single mode.
      # shiroxysinglesecure:
      #   domain: "shikharcode.com"

  # This is used to manage specific behaviors of HTTP connections and
  # client information forwarding.
  options:
    - "http-server-close"
    - "forwardfor"

  # This specifies whether to run/expose the frontend of shiroxy
  # in secure mode or not. If set to true, you are required to
  # configure the secure settings in the bind section of the frontend section.
  secure: false

  # This sets the behavior of the server when secure is set to true.
  # The `secureverify` parameter is used to enforce certificate
  # verification for secure connections. Possible values are `none`,
  # `optional`, and `required`.
  # `none` - No client certificate is requested during the handshake, and if any certificates are sent, they are not verified.
  # `optional` - A client certificate is requested during the handshake, but it does not require the client to send any certificates.
  # `required` - A client certificate is requested during the handshake, and at least one valid certificate is required from the client.
  secureverify: "required"

  # This indicates which load balancing algorithm to use. Shiroxy supports
  # `round-robin`, `sticky-session`, and `least-count`.
  balance: "round-robin"

  # Will be supported soon
  # defaultbackend: ""

  # Will be supported soon
  # fallbackbackend: ""

# This section holds configuration about the backend to which shiroxy will
# be forwarding the requests and how it is going to do that.
backend:
  # Name of the backend. In future versions, we will support multiple backends,
  # which is why we are keeping this parameter from version one of the
  # configuration file.
  name: "shiroxy-test"

  # This section sets how many servers that backend will have.
  servers:
    - id: "<backend-name-server-1>"
      # Host address of the service
      host: ""
      # Port on which that service is running
      port: "8001"
      # This sets the health URL of that service. If not set, the base URL
      # will be used as the health URL (i.e., `host:port`).
      healthurl: ""

    - id: "<backend-name-server-2>"
      host: ""
      port: "8002"
      healthurl: ""

  # This sets the mode through which the health of the server will be checked.
  # If set to `url`, you have to provide the health URL in the server section
  # of the backend.
  healthcheckmode: "home/url"

  # This sets the frequency by which the health of the services will be checked.
  healthchecktriggerduration: 5

# Will be supported soon
logging:
  enable: true
  mode: "native/http/native-http"
  schema:
    - "[date-time] "
  include:
    - "*"
    - ""

# Webhook for different events that happen after API calls
webhook:
  # Toggle enabling/disabling the webhook
  enable: true
  # Events on which shiroxy should fire the webhook
  events:
    - "domain-register-success"
    - "domain-register-failed"
    - "domain-ssl-success"
    - "domain-ssl-failed"
    - "domain-remove-success"
    - "domain-remove-failed"
    - "domain-update-success"
    - "domain-update-failed"
  # Webhook URL
  url: "https://example.com/webhooks"
```

### Default Section

The default section contains settings that are global to shiroxy.

```yaml
default:
  mode: "http"
  logpath: "/home/shikharcode/Main/opensource/shiroxy"
  datapersistancepath: "/home/shikharcode/Main/opensource/shiroxy"
  enablednschallengesolver: ""
  timeout:
    connect: "10s"
    client: "10s"
    server: "10s"
  storage:
    location: "memory"
  analytics:
    collectioninterval: 10
    routename: "/analytics"
  errorresponses:
    errorpagebuttonname: "Comeata"
    errorpagebuttonurl: "https://youtube.com"
  user:
    email: "yshikharfzd10@gmail.com"
    secret: "kjksdnfiwj"
```

- **mode**: The mode in which shiroxy will be running. Default is "http".
- **logpath**: Path where the logs will be stored.
- **datapersistancepath**: Path where data will be persisted in case of failure or manual stop.
- **enablednschallengesolver**: Whether to start the DNS challenge solver.
- **timeout**: Specifies the timeout for different scenarios (connect, client, server).
- **storage**: Specifies how the SSL certificates will be stored (location can be "redis" or "memory").
- **analytics**: Settings related to analytics (collection interval and API base path).
- **errorresponses**: Settings for the error page button (label and URL).
  user: User email and secret.

## Frontend Section

The frontend section specifies settings related to the frontend of the reverse proxy.

```yaml
frontend:
  backend: "shiroxy-test"
  bind:
    port: "8080"
    host: ""
    secure:
      target: "single"
      singletargetmode: "certandkey"
      certandkey:
        cert: "cert-location"
        key: "key-location"
    options:
      - "http-server-close"
      - "forwardfor"
    secure: false
    secureverify: "required"
    balance: "round-robin"
```

- **backend**: Name of the backend to which the frontend will forward requests.
- **bind**: Settings for binding the frontend (port, host, SSL configuration).
  - **secure**: Domain SSL configuration (target can be "single" or "multiple").
    - **singletargetmode**: Mode for securing a single domain ("certandkey" or "shiroxysinglesecure").
    - **certandkey**: Location of the SSL certificate and key.
  - **options**: HTTP connection options.
  - **secure**: Whether to run the frontend in secure mode.
  - **secureverify**: Enforces certificate verification for secure connections (possible values: "none", "optional", "required").
  - **balance**: Load balancing algorithm (possible values: "round-robin", "sticky-session", "least-count").

## Backend Section

The backend section holds configuration about the backend to which shiroxy will forward requests.

```yaml
backend:
  name: "shiroxy-test"
  servers:
    - id: "<backend-name-server-1>"
      host: ""
      port: "8001"
      healthurl: ""
    - id: "<backend-name-server-2>"
      host: ""
      port: "8002"
      healthurl: ""
  healthcheckmode: "home/url"
  healthchecktriggerduration: 5
```

- **name**: Name of the backend.
- **servers**: List of backend servers (each with unique id, host, port, and optional health URL).
- **healthcheckmode**: Mode for checking server health ("home/url").
- **healthchecktriggerduration**: Frequency of health checks (in seconds).

## Webhook Section

The webhook section configures webhooks for different events.

```yaml
webhook:
  enable: true
  events:
    - "domain-register-success"
    - "domain-register-failed"
    - "domain-ssl-success"
    - "domain-ssl-failed"
    - "domain-remove-success"
    - "domain-remove-failed"
    - "domain-update-success"
    - "domain-update-failed"
  url: "https://example.com/webhooks"
```

- **enable**: Whether to enable webhooks.
- **events**: List of events that trigger webhooks.
- **url**: URL to which the webhook is sent.

## Contributing

We welcome contributions! If you'd like to contribute to this project, please fork the repository and create a pull request with your changes. For major changes, please open an issue first to discuss what you would like to change. See [Contribution Guideline](https://github.com/ShikharY10/shiroxy/blob/main/CONTRIBUTION.md)
