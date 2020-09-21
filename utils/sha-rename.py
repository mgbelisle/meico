#!/usr/bin/env python

from hashlib import sha1
import os
import sys


json_file = sys.argv[1]
photos_dir = sys.argv[2]
sha1_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'src', 'img', 'sha1')

for root, dnames, fnames in os.walk(photos_dir):
    for fname in fnames:
        base, ext = os.path.splitext(fname)
        ext = ext.lower()
        with open(os.path.join(root, fname), 'rb') as fhandle:
            img = fhandle.read()
        hash = sha1(img).hexdigest()
        with open(os.path.join(sha1_dir, hash+ext), 'wb') as fhandle:
            fhandle.write(img)
        relpath = os.path.relpath(os.path.join(root, fname), start=photos_dir)
        with open(json_file) as fhandle:
            json = fhandle.read()
        with open(json_file, 'w') as fhandle:
            fhandle.write(json.replace(f'"{relpath}"', f'"{hash+ext}"'))
