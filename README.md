# filefind

Find files in huge directory trees.

Features:
 * Browse and search in a huge filesystem, which might not be available online or slow to access (e.g. NAS, S3, Glacier, ...)
 * Browse the directory tree. Use filters to quickly narrow down to interesing parts of the tree
 * Search in the directory tree using regular expressions


## QNAP notes

export PATH=$PATH:/share/Henning/go/bin/
export GO_PATH=/share/Henning/go_path/

go run main_printonly.go -path /share/Qusb/ | gzip > qusb.zip
go run main_printonly.go -path /share/Qusb/ | gzip > qusb.zip 2> log.txt &


## Backlog

 * TODO Test umlauts, unicodes, etc. encoding: /share/Qusb/2011 Capgemini/Stream/2011-Q3 Masterarbeit Felix BÃ¶hm
 * TODO Set base path, remove from stored path (e.g. /share/Qusb/2011 Capgemini -> /2011 Capgemini/)
