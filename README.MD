# FileDateDispatcher

## Presentation

**FileDateDispatcher** is a CLI tool designed to dispatch multimedia files by date.

Let's consider that you have a folder containing a bunch of files (pictures, movies, ...) and you would like to regroup these files by month (creation date, ...). That is what does **FileDateDispatcher**.

## Usage

###

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