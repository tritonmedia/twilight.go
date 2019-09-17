<p align="center">
  <img src="https://raw.githubusercontent.com/jaredallard/media-stack/master/.github/twilight.png" alt="Twilight Sparkle with books" />
</p>

<p align="center">
  <code>twilight.go</code>
</p>

<p align="center">Twilight organizes your media so you don't have too.</p>

## Role

Twilight determines the name of your media and where it gets stored. This
information is then passed to [identifier](https://github.com/tritonmedia/identifier)
to be exposed to the various api clients.

## Configuration

| EnvVar                    | Description                | Conditions             |
|---------------------------|----------------------------|------------------------|
| TWILIGHT_STORAGE_PROVIDER | Storage Provider to use    | Must be `s3/fs`        |
| S3_ACCESS_KEY             | S3 Access Key              | Only works in S3 mode  |
| S3_SECRET_KEY             | S3 Secret Key              | Only works in S3 mode  |
| S3_ENDPOINT               | S3 Endpoint                | Only works in S3 mode. |
| S3_BUCKET                 | S3 Bucket                  | Only works in S3 mode  |
| TWILIGHT_FS_BASE          | Base dir for FS storage    | Only works in FS mode  |
| TWILIGHT_DEBUG            | Enable debug logging       | Set to != "" to enable |
| RABBITMQ                  | RabbitMQ Endpoint          |                        |
| PORT                      | Port to run HTTP server on | Must be an integer     |

## License

Apache-2.0
