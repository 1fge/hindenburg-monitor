# Hindenburg Research Site Monitor

## Overview

This project is designed to monitor various endpoints on the Hindenburg Research site, enabling quick notifications for new entries of companies they have researched. It can detect updates through WordPress posts, media endpoints, or directly from their sitemap. The goal is to provide an efficient way to fetch the latest publications before they move the markets.

## Structure

The project is organized into two main directories:

- **`internal-http-pkgs`**: Contains third-party request libraries for Go, handling all aspects related to TLS/HTTP2 to ensure proper fingerprinting and secure connections. This is almost certainly overkill, but it should help futureproof the code base in case they add more protection. This setup should simplify dependency management.

- **`monitor`**: Holds the core code necessary to monitor the Hindenburg Research site as described. It includes mechanisms to fetch endpoint data, detect changes, and send notifications via a configured Discord webhook.

## Prerequisites

- Go installed on your system
- Discord webhook URL set in your environment variables (`WEBHOOKURL`)

## Setup

1. **Environment Variable**: Ensure that the `WEBHOOKURL` environment variable is set with your Discord webhook URL. This is crucial for the application to send notifications about site changes.

2. **Dependencies**: Navigate to the project's root directory and install all required dependencies:
    ```bash
    go get
    ```

    You'll likely need to disable Go modules when using the project:
    ```bash
    export GO111MODULE=off
    ```

## Running the Project

To run the project, simply execute the following command in the `monitor` directory:

```bash
go run .
```
This will start the monitoring process, checking the specified endpoints on the Hindenburg Research site for changes and sending notifications to the configured Discord webhook URL.

## Customizing Refresh Delay
To modify the frequency of checks, you can adjust the refreshDelay variable in main.go. The default delay is set to 2500ms. To change this, simply edit the value of refreshDelay to your desired interval before running the code:
```go
var refreshDelay = 2500 * time.Millisecond // Change this value to your preferred delay
```

## Debugging
The project contains a debug flag within the code in the monitor directory, specifically set up to use Charles Proxy on its default port of 8888. This can be useful for  monitoring the network traffic for development purposes or troubleshooting issues. To enable this debug mode, ensure that the corresponding flag in the code is set before running the application.
