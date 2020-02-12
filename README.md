# filefind

Find files in huge directory trees.

Features:
 * Browse and search in a huge filesystem, which might not be available online or slow to access (e.g. NAS, S3, Glacier, ...)
 * Browse the directory tree. Use filters to quickly narrow down to interesing parts of the tree
 * Search in the directory tree using regular expressions


## QNAP notes

export PATH=$PATH:/share/Henning/go/bin/
export GO_PATH=/share/Henning/go_path/

go build
./filetree_to_csv -path /share/Qusb/ -base /share | gzip > qusb.zip
./filetree_to_csv -path /share/Qusb/ -base /share | gzip > qusb.zip 2> log.txt &


## Backlog

 * TODO Test umlauts, unicodes, etc. encoding
