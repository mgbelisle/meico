#!/usr/bin/env python

from hashlib import sha1
import os
import sys


sha1_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', 'src', 'img', 'sha1')

for root, dnames, fnames in os.walk(sys.argv[1]):
    for fname in fnames:
        base, ext = os.path.splitext(fname)
        with open(os.path.join(root, fname), 'rb') as fhandle:
            content = fhandle.read()
        hash = sha1(content).hexdigest()
        with open(os.path.join(sha1_dir, hash+ext.lower()), 'wb') as fhandle:
            fhandle.write(content)