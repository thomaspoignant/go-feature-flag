#!/usr/bin/env python
import json
import re

mike_json = input()
versions = json.loads(mike_json)
versionPattern = r'^(v?)(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(' \
                 r'?:0|[' \
                 r'1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$'
regexMatcher = re.compile(versionPattern)

sorted_versions = sorted(versions, key=lambda d: d['version'])
sorted_versions.reverse()
keep = []
to_delete = []
# Compute which version we want to keep
for v in sorted_versions:
    if len(v['aliases']) > 0:
        keep.append(v['version'])
        continue

    regex_group = regexMatcher.search(v['version'])
    minor = '{}{}.{}'.format(regex_group.group(1), regex_group.group(2), regex_group.group(3))
    with_minor = list(filter(lambda version: version.startswith(minor), keep))
    if len(with_minor) == 0:
        keep.append(v['version'])
    else:
        to_delete.append(v['version'])

print(" ".join(to_delete))
