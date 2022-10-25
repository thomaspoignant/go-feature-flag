# Where to store my file?
The module supports different ways of retrieving the flag file.  
Available retriever are:

- [S3 Bucket](s3)
- [HTTP endpoint](http)
- [Github](github)
- [File](file)

To retrieve a file you need to provide a [retriever](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Retriever) in your `ffclient.Config{}` during the initialization.  
If the existing retriever does not work with your system you can extend the system and use a [custom retriever](custom.md).
