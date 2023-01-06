# Where to store my file?
The module supports different ways of retrieving the flag file.  
Available retriever are:

- [S3 Bucket](./s3.md)
- [HTTP endpoint](./http.md)
- [Github](./github.md)
- [File](./file.md)

To retrieve a file you need to provide a [retriever](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Retriever) in your `ffclient.Config{}` during the initialization.
