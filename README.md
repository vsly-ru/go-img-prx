# go-img-prx

go-img-prx is a high-performance image proxy and resizing service written in Go. It allows on-the-fly image resizing, format conversion, and caching.

## Features

- On-the-fly image resizing
- Format conversion (supports JPG, PNG, and WebP)
- Image caching for improved performance
- Can run as a CLI tool or HTTP server
- Docker support for easy deployment

## Installation

### Prerequisites

- Go 1.23 or higher
- Docker (optional, for containerized deployment)

### Building from source

1. Clone the repository:
```
git clone https://github.com/vsly-ru/go-img-prx.git
cd go-img-prx
```
2. Build the application:
```
go build -o go-img-prx
```

### Docker


1. Build the Docker image:
```
docker build -t go-img-prx .
```
2. Run the container:
```
docker run -p 8080:8080 go-img-prx
# or with cache volume (to persist resized images)
docker run -p 8080:8080 -v go-img-prx-cache:/app/cache go-img-prx

```

## Usage

### CLI Mode
```
./go-img-prx -url <image_url> -f <format> -w <width> -h <height>
```

Example:
```
./go-img-prx -url https://picsum.photos/1280 -f webp -w 600 -h 600
```
### Server Mode

Start the server:
```
./go-img-prx -server
```

Then access images via URL:
```
http://localhost:8080/format:<format>/resize:<mode>:<width>:<height>/plain/<image_url>
```

- `image_url` url-encoded URL of the image to resize
- `format` output format (jpg, png, webp)
- `mode` resize mode (currently only 'fill' is supported)
- `width` desired width
- `height` desired height

Example:
```
http://localhost:8080/format:webp/resize:fit:800:600/plain/https%3A%2F%2Fpicsum.photos/1280
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the [MIT License](LICENSE).