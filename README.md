# filefind

Find files in huge directory trees.

Features:
 * Browse and search in a huge filesystem, which might not be available online or slow to access (e.g. NAS, S3, Glacier, ...)
 * Browse the directory tree. Use filters to quickly narrow down to interesing parts of the tree
 * Search in the directory tree using regular expressions


## QNAP notes

```bash
# On local machine
rsync -avzhe ssh ./ admin@henningsnas.local:/share/Henning/filefind/

# On QNAP
export PATH=$PATH:/share/Henning/go/bin/
export GO_PATH=/share/Henning/go_path/

# Export csv
cd /share/Henning/filefind/cmd/filetree_to_csv
go build
./filetree_to_csv -path /share/Qusb/ -base /share | gzip > qusb.zip
./filetree_to_csv -path /share/Qusb/ -base /share | gzip > qusb.zip 2> log.txt &

# Export mirror tree
cd /share/Henning/filefind/cmd/filetree_mirror
go build
./filetree_mirror -src /share/Qusb/ -dest ./mirror/ &
tar cfz mirror.tar.gz ./mirror/

# Back on local machine
scp admin@henningsnas.local:/share/Henning/filefind/cmd/filetree_mirror/mirror.tar.gz .
tar xfz mirror.tar.gz
```

## Backlog
 * TODO Test umlauts, unicodes, etc. encoding
