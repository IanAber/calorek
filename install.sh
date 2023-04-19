#!/bin/bash
systemctl stop calorek
chmod +x bin/amd64/calorek
cp bin/amd64/calorek /usr/bin
cp -r web/* /calorek/web
systemctl start calorek
