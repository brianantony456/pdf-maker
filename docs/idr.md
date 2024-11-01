# Pdf Reader design

## App startup

### Db & Table design
- creates empty database if it doesn't exist.
- Allows to add columns of certain type.
- On restart re-use the existing DB.
- Create rows and add data.

### Ui Design
- Add column of specific types
- Add rows
- Pdf format saved in the DB.
- Print option to generate a pdf to print

## Reliability
- Backup data timely to a gmail drive. (Whats better ?)

## Tools used
- For easy scripts: `Makefile` with `make`
- For Live Reload [air](https://github.com/air-verse/air)
- Web Framework [Gin](https://github.com/gin-gonic/gin)
- Used embed package to create a single binary https://github.com/gin-gonic/gin/blob/master/docs/doc.md#graceful-shutdown-or-restart

## Directory structure
- bin/ - Contains created binaries
