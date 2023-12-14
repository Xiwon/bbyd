#! /bin/bash

go build .
./cmd myconf.toml

# my build script using another config file which is gitignored...
