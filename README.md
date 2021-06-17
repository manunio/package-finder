# Private Package Finder

A simple go program which goes through given urls, downloads js file from each urls,
parses it for package.json embedded in js files and list packages found in js file,
and then adds it to a local log file.

# package-finder.service [systemd]
- location: /etc/systemd/system/package-finder.service
- enable command: sudo systemctl enable package-finder.service 
- start command: sudo systemctl start package-finder.service 
