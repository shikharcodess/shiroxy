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
- [Configuration File](https://github.com/ShikharY10/shiroxy/blob/main/configuration.md)
- [Contributing](#contributing)
- [License](https://github.com/ShikharY10/shiroxy/blob/main/LICENSE)

<!-- - [API Documentation](#api-documentation) -->

## Getting Started

To get started with shiroxy, clone the repository and follow the installation instructions.

```bash
git clone https://github.com/yourusername/shiroxy.git
cd shiroxy
go build -o build/shiroxy

./shiroxy -config /path/to/your/config.yml
```

## Contributing

We welcome contributions! If you'd like to contribute to this project, please fork the repository and create a pull request with your changes. For major changes, please open an issue first to discuss what you would like to change. See [Contribution Guideline](https://github.com/ShikharY10/shiroxy/blob/main/CONTRIBUTION.md)
