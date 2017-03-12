
BitBucket pipelines YAML parser [![Build Status](https://travis-ci.org/bivas/bitbucket-pipelines.svg?branch=master)](https://travis-ci.org/bivas/bitbucket-pipelines)

Parse [BitBucket Pipelines](https://confluence.atlassian.com/bitbucket/bitbucket-pipelines-792496469.html) to Go structs.

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
