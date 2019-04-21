# FileDateDispatcher

## Presentation

**FileDateDispatcher** is a CLI tool designed to dispatch multimedia files by date.

Let's consider that you have a folder containing a bunch of files (pictures, movies, ...) and you would like to regroup these files by month (creation date, ...). That is what does **FileDateDispatcher**.

## Configuration

```json
{
    "loggingLevel":"info",
    "batchSize":10,
    "dateFields": [
        { "field":"CreateDate", "pattern":"2006:01:02 15:04:05" },
        { "field":"Media Create Date", "pattern":"2006:01:02 15:04:05" }
    ]
}
```

- **loggingLevel** : logging level (debug, info, warn, error, fatal, panic)
- **batchSize** : how many files are provided to exiftool per invocation
- **dateFields** : exiftool tags that have to be considered as valid date for dispatching
  - **dateFields.field** : exiftool tag key
  - **dateFields.pattern** : date pattern, based on golang specifications (https://golang.org/pkg/time/#Time.Format)

## Usage

### From binary

#### Dependency

**FileDataDispatcher** requires **exiftool** to be installed :
- Ubuntu : `sudo apt-get install libimage-exiftool-perl`

#### Compilation

`go build cmd/dispatcher.go`

#### Execution

Arguments :
- `-s` : source folder (required)
- `-d` : destination folder (required)
- `-l` : logging level (debug|info|warn|error|fatal|panic, default: info)
- `-b` : batch size (default: 10)

Example input :
- `/tmp/in/toto.jpg`
- `/tmp/in/a/tutu.avi`

Exécution : `./dispatcher -s /tmp/in -d /tmp/out`

Output :
- `/tmp/out/2019_01/toto.jpg`
- `/tmp/out/2019_02/tutu.avi`

### Docker

#### Building image

`docker build -t filedatedispatcher:latest .`

#### Execution

When executing the container, it will use :
- `/var/dispatcher/in` as a source folder (inside the container)
- `/var/dispatcher/out` as a target folder (inside the container)
- `/etc/dispatcher/dispatcher.json` as a configuration file (a [default one](https://github.com/barasher/FileDateDispatcher/blob/master/docker.json) is integrated)

So, to deal with folder from the host, you will have to specify volumes :

```
docker run
  --rm
  -v [sourceFolder from host]:/var/dispatcher/in
  -v [targetFolder from host]:/var/dispatcher/out
  -v [configurationFile from host]:/etc/dispatcher/dispatcher.json
  filedatedispatcher:latest`
```