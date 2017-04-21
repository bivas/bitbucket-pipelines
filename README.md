
[BitBucket Pipelines](https://confluence.atlassian.com/bitbucket/bitbucket-pipelines-792496469.html) YAML runner/parser [![Build Status](https://travis-ci.org/bivas/bitbucket-pipelines.svg?branch=master)](https://travis-ci.org/bivas/bitbucket-pipelines)

# Runner 

See [Examples](examples/README.md)   

# Parser

Parse pipelines to Go structs.

Using the following example:
```
image: python:2.7
 
pipelines:
  default:
    - step:
        script:
          - python --version
          - python myScript.py
```

Will output the following struct:
```
{Image:python:2.7 Pipelines:{Default:[{Step:{Scripts:[python --version python myScript.py]}}] Branches:[] Tags:[] Bookmarks:[]}}
```
