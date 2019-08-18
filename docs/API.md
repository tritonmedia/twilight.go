## Upload Media

**Endpoint**: PUT - https://twilight/v1/media

**Description**: Uploads a media file to Twilight

#### Header

```
{
	X-Media-Type: <tv/movie>
	X-Media-Name: <Name of the media>
	X-Media-ID: <ID of the Media>
	X-Media-Quality: <Quality Bucket: 480/720/1080p
}
```

#### Body

**Type**: `multipart/form-data`

```
{
	file: <data>
}
```